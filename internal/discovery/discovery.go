package discovery

import (
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"unicode"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/platform"
	"github.com/tianrking/ClawRemove/internal/system"
)

type Discoverer struct {
	runner  system.Runner
	facts   model.ProductFacts
	host    platform.Host
	adapter platform.Adapter
}

func New(runner system.Runner, facts model.ProductFacts, host platform.Host) Discoverer {
	return Discoverer{
		runner:  runner,
		facts:   facts,
		host:    host,
		adapter: platform.NewAdapter(host),
	}
}

func (d Discoverer) Discover(ctx context.Context) (model.Discovery, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return model.Discovery{}, err
	}

	// Phase 1: Get stateDirs first (needed by workspaces)
	stateDirs := d.discoverStateDirs(home)

	// Phase 2: Parallel discovery for independent operations
	var (
		wg                 sync.WaitGroup

		workspaceDirs      []string
		tempPaths          []string
		shellProfiles      []string
		appPaths           []string
		cliPaths           []string
		packages           []model.PackageRef
		services           []model.ServiceRef
		processes          []model.ProcessRef
		listeners          []string
		crontabLines       []string
		containers         []model.ContainerRef
		images             []model.ImageRef
		registryKeys       []model.RegistryRef
		envVars            []model.EnvVarRef
		hostsEntries       []string
	)

	// Launch parallel goroutines for independent discoveries
	wg.Add(16)
	go func() {
		defer wg.Done()
		workspaceDirs = d.discoverWorkspaces(home, stateDirs)
	}()
	go func() {
		defer wg.Done()
		tempPaths = d.discoverTempPaths()
	}()
	go func() {
		defer wg.Done()
		shellProfiles = d.discoverShellProfiles(home)
	}()
	go func() {
		defer wg.Done()
		appPaths = d.discoverAppPaths(home)
	}()
	go func() {
		defer wg.Done()
		cliPaths = d.discoverCLIPaths(home)
	}()
	go func() {
		defer wg.Done()
		packages = d.discoverPackages(ctx)
	}()
	go func() {
		defer wg.Done()
		services = d.discoverServices(ctx, home)
	}()
	go func() {
		defer wg.Done()
		processes = d.discoverProcesses(ctx)
	}()
	go func() {
		defer wg.Done()
		listeners = d.discoverListeners(ctx)
	}()
	go func() {
		defer wg.Done()
		crontabLines = d.discoverCrontab(ctx)
	}()
	go func() {
		defer wg.Done()
		containers = d.discoverContainers(ctx)
	}()
	go func() {
		defer wg.Done()
		images = d.discoverImages(ctx)
	}()
	go func() {
		defer wg.Done()
		registryKeys = d.discoverRegistryKeys(ctx)
	}()
	go func() {
		defer wg.Done()
		envVars = d.discoverEnvVars(ctx)
	}()
	go func() {
		defer wg.Done()
		hostsEntries = d.discoverHostsEntries(ctx)
	}()
	// Extra goroutine to wait and close a signal channel if needed
	go func() {
		wg.Wait()
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	return model.Discovery{
		Platform:      d.host.OS,
		HomeDir:       home,
		StateDirs:     stateDirs,
		WorkspaceDirs: workspaceDirs,
		TempPaths:     tempPaths,
		ShellProfiles: shellProfiles,
		AppPaths:      appPaths,
		CLIPaths:      cliPaths,
		Packages:      packages,
		Services:      services,
		Processes:     processes,
		Listeners:     listeners,
		CrontabLines:  crontabLines,
		Containers:    containers,
		Images:        images,
		RegistryKeys:  registryKeys,
		EnvVars:       envVars,
		HostsEntries:  hostsEntries,
	}, nil
}

func (d Discoverer) discoverStateDirs(home string) []string {
	var out []string
	for _, name := range d.facts.StateDirNames {
		out = append(out, filepath.Join(home, name))
	}

	entries, _ := os.ReadDir(home)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), ".openclaw-") {
			out = append(out, filepath.Join(home, entry.Name()))
		}
	}
	return uniqExistingish(out)
}

func (d Discoverer) discoverWorkspaces(home string, stateDirs []string) []string {
	var out []string

	// Use provider-declared workspace dir names first
	workspaceDirs := d.facts.WorkspaceDirNames
	if len(workspaceDirs) == 0 {
		// Fallback to convention: look for "workspace" subdirectory in each stateDir
		for _, stateDir := range stateDirs {
			out = append(out, filepath.Join(stateDir, "workspace"))
		}
	} else {
		for _, stateDir := range stateDirs {
			for _, wname := range workspaceDirs {
				out = append(out, filepath.Join(stateDir, wname))
			}
		}
	}

	// Dynamic workspace variants discovered inside any matching state dir
	for _, stateDir := range stateDirs {
		if entries, err := os.ReadDir(stateDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				if strings.HasPrefix(entry.Name(), "workspace") {
					out = append(out, filepath.Join(stateDir, entry.Name()))
				}
			}
		}
	}

	return uniqExistingish(out)
}

func (d Discoverer) discoverTempPaths() []string {
	root := os.TempDir()
	entries, _ := os.ReadDir(root)
	var out []string
	for _, prefix := range d.facts.TempPrefixes {
		if prefix == "openclaw" {
			out = append(out, filepath.Join(root, prefix))
		}
	}
	for _, entry := range entries {
		name := entry.Name()
		for _, prefix := range d.facts.TempPrefixes {
			if strings.HasPrefix(name, prefix) {
				out = append(out, filepath.Join(root, name))
				break
			}
		}
	}
	return uniqExistingish(out)
}

func (d Discoverer) discoverShellProfiles(home string) []string {
	var out []string
	for _, rel := range d.facts.ShellProfileGlobs {
		absPath := filepath.Join(home, filepath.FromSlash(rel))
		// Only include profile if it exists AND contains a marker string
		if !fileContainsMarker(absPath, d.facts.Markers) {
			continue
		}
		out = append(out, absPath)
	}
	return dedupe(out)
}

func (d Discoverer) discoverAppPaths(home string) []string {
	var out []string
	for _, raw := range d.facts.AppPaths {
		if filepath.IsAbs(raw) {
			out = append(out, raw)
			continue
		}
		out = append(out, filepath.Join(home, filepath.FromSlash(raw)))
	}
	return uniqExistingish(out)
}

func (d Discoverer) discoverCLIPaths(home string) []string {
	var out []string
	for _, rel := range d.facts.CLIPaths {
		out = append(out, filepath.Join(home, filepath.FromSlash(rel)))
	}
	return uniqExistingish(out)
}

func (d Discoverer) discoverPackages(ctx context.Context) []model.PackageRef {
	var out []model.PackageRef
	for _, ref := range d.facts.PackageRefs {
		if !d.runner.Exists(ctx, ref.Manager) {
			continue
		}
		switch ref.Manager {
		case "npm", "pnpm", "bun":
			out = append(out, ref)
		case "brew":
			if d.host.OS != "darwin" {
				continue
			}
			args := []string{"list"}
			if ref.Kind == "cask" {
				args = append(args, "--cask", ref.Name)
			} else {
				args = append(args, "--formula", ref.Name)
			}
			if d.runner.Run(ctx, "brew", args...).OK {
				out = append(out, ref)
			}
		}
	}
	return uniqPackages(out)
}

func (d Discoverer) discoverServices(ctx context.Context, home string) []model.ServiceRef {
	switch d.host.OS {
	case "darwin":
		return d.discoverDarwinServices(home)
	case "linux":
		return d.discoverLinuxServices(home)
	case "windows":
		return d.discoverWindowsServices(ctx)
	default:
		return nil
	}
}

func (d Discoverer) discoverDarwinServices(home string) []model.ServiceRef {
	var out []model.ServiceRef
	dirs := []struct {
		scope string
		dir   string
	}{
		{scope: "user", dir: filepath.Join(home, "Library", "LaunchAgents")},
		{scope: "system", dir: "/Library/LaunchAgents"},
		{scope: "system", dir: "/Library/LaunchDaemons"},
	}
	for _, item := range dirs {
		entries, _ := os.ReadDir(item.dir)
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".plist") {
				continue
			}
			name := strings.ToLower(entry.Name())
			if hasMarker(name, d.facts.Markers) {
				out = append(out, model.ServiceRef{
					Platform: "darwin",
					Scope:    item.scope,
					Name:     strings.TrimSuffix(entry.Name(), ".plist"),
					Path:     filepath.Join(item.dir, entry.Name()),
				})
			}
		}
	}

	// Also check for running launchd services
	if d.runner.Exists(context.Background(), "launchctl") {
		result := d.runner.Run(context.Background(), "launchctl", "list")
		if result.OK {
			for _, line := range strings.Split(result.Stdout, "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "PID") {
					continue
				}
				fields := strings.Fields(line)
				if len(fields) >= 3 {
					serviceName := fields[2]
					if hasMarker(strings.ToLower(serviceName), d.facts.Markers) {
						// Check if we haven't already added this service
						found := false
						for _, s := range out {
							if s.Name == serviceName {
								found = true
								break
							}
						}
						if !found {
							out = append(out, model.ServiceRef{
								Platform: "darwin",
								Scope:    "user",
								Name:     serviceName,
							})
						}
					}
				}
			}
		}
	}

	return uniqServices(out)
}

func (d Discoverer) discoverLinuxServices(home string) []model.ServiceRef {
	var out []model.ServiceRef
	dirs := []struct {
		scope string
		dir   string
	}{
		{scope: "user", dir: filepath.Join(home, ".config", "systemd", "user")},
		{scope: "system", dir: "/etc/systemd/system"},
		{scope: "system", dir: "/usr/lib/systemd/system"},
		{scope: "system", dir: "/lib/systemd/system"},
		{scope: "system", dir: "/run/systemd/system"},
	}
	for _, item := range dirs {
		entries, _ := os.ReadDir(item.dir)
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			// Check for .service, .timer, and .socket files
			isService := strings.HasSuffix(name, ".service")
			isTimer := strings.HasSuffix(name, ".timer")
			isSocket := strings.HasSuffix(name, ".socket")

			if !isService && !isTimer && !isSocket {
				continue
			}

			baseName := name
			for _, suffix := range []string{".service", ".timer", ".socket"} {
				baseName = strings.TrimSuffix(baseName, suffix)
			}

			if hasMarker(strings.ToLower(baseName), d.facts.Markers) {
				ref := model.ServiceRef{
					Platform: "linux",
					Scope:    item.scope,
					Name:     baseName,
					Path:     filepath.Join(item.dir, name),
				}
				if isTimer {
					ref.Name = name // Keep full name for timers
				}
				out = append(out, ref)
			}
		}
	}

	// Also check for running systemd services
	if d.runner.Exists(context.Background(), "systemctl") {
		// Check user services
		result := d.runner.Run(context.Background(), "systemctl", "--user", "list-units", "--type=service", "--no-pager")
		if result.OK {
			out = append(out, parseSystemdUnits(result.Stdout, "user", d.facts.Markers)...)
		}

		// Check system services
		result = d.runner.Run(context.Background(), "systemctl", "list-units", "--type=service", "--no-pager")
		if result.OK {
			out = append(out, parseSystemdUnits(result.Stdout, "system", d.facts.Markers)...)
		}
	}

	return uniqServices(out)
}

// parseSystemdUnits parses systemctl list-units output.
func parseSystemdUnits(output, scope string, markers []string) []model.ServiceRef {
	var out []model.ServiceRef
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "UNIT") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}
		unitName := fields[0]
		if !strings.HasSuffix(unitName, ".service") {
			continue
		}
		serviceName := strings.TrimSuffix(unitName, ".service")
		if hasMarker(strings.ToLower(serviceName), markers) {
			out = append(out, model.ServiceRef{
				Platform: "linux",
				Scope:    scope,
				Name:     serviceName,
			})
		}
	}
	return out
}

func (d Discoverer) discoverWindowsServices(ctx context.Context) []model.ServiceRef {
	cmd := d.adapter.ScheduledTaskListCommand()
	if len(cmd) == 0 || !d.runner.Exists(ctx, cmd[0]) {
		return nil
	}
	result := d.runner.Run(ctx, cmd[0], cmd[1:]...)
	if !result.OK {
		return nil
	}
	var out []model.ServiceRef
	for _, block := range strings.Split(result.Stdout, "\n\n") {
		lower := strings.ToLower(block)
		if !hasMarker(lower, d.facts.Markers) {
			continue
		}
		for _, line := range strings.Split(block, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(strings.ToLower(line), "taskname:") {
				out = append(out, model.ServiceRef{
					Platform: "windows",
					Scope:    "user",
					Name:     strings.TrimSpace(strings.TrimPrefix(line, "TaskName:")),
				})
				break
			}
		}
	}
	return uniqServices(out)
}

func (d Discoverer) discoverProcesses(ctx context.Context) []model.ProcessRef {
	cmd := d.adapter.ProcessListCommand()
	if len(cmd) == 0 || !d.runner.Exists(ctx, cmd[0]) {
		return nil
	}

	result := d.runner.Run(ctx, cmd[0], cmd[1:]...)
	if !result.OK {
		return nil
	}
	if d.host.OS == "windows" {
		if !result.OK {
			return nil
		}
		reader := csv.NewReader(strings.NewReader(result.Stdout))
		rows, _ := reader.ReadAll()
		var out []model.ProcessRef
		for _, row := range rows {
			line := strings.ToLower(strings.Join(row, " "))
			if hasMarker(line, d.facts.Markers) {
				out = append(out, model.ProcessRef{Command: strings.Join(row, " | ")})
			}
		}
		return out
	}

	var out []model.ProcessRef
	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		lower := strings.ToLower(line)
		if line == "" || strings.Contains(lower, "claw-remove") || !hasMarker(lower, d.facts.Markers) {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			out = append(out, model.ProcessRef{Command: line})
			continue
		}
		out = append(out, model.ProcessRef{
			PID:     atoi(fields[0]),
			PPID:    atoi(fields[1]),
			Command: strings.Join(fields[2:], " "),
		})
	}
	return out
}

func (d Discoverer) discoverListeners(ctx context.Context) []string {
	candidates := d.adapter.ListenerCommands()
	for _, candidate := range candidates {
		if !d.runner.Exists(ctx, candidate[0]) {
			continue
		}
		result := d.runner.Run(ctx, candidate[0], candidate[1:]...)
		if !result.OK {
			continue
		}
		var out []string
		for _, line := range strings.Split(result.Stdout, "\n") {
			line = strings.TrimSpace(line)
			lower := strings.ToLower(line)
			if line == "" {
				continue
			}
			markerMatch := hasMarker(lower, d.facts.Markers)
			portMatch := false
			for _, port := range d.facts.ListenerPorts {
				if strings.Contains(line, itoa(port)) {
					portMatch = true
					break
				}
			}
			if markerMatch || portMatch {
				out = append(out, line)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return nil
}

func (d Discoverer) discoverCrontab(ctx context.Context) []string {
	if d.host.OS == "windows" || !d.runner.Exists(ctx, "crontab") {
		return nil
	}
	result := d.runner.Run(ctx, "crontab", "-l")
	var out []string
	for _, line := range strings.Split(result.Stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if hasMarker(strings.ToLower(line), d.facts.Markers) {
			out = append(out, line)
		}
	}
	return out
}

func (d Discoverer) discoverContainers(ctx context.Context) []model.ContainerRef {
	var out []model.ContainerRef
	for _, runtimeName := range []string{"docker", "podman"} {
		if !d.runner.Exists(ctx, runtimeName) {
			continue
		}
		result := d.runner.Run(ctx, runtimeName, "ps", "-a", "--format", "{{.ID}}\t{{.Image}}\t{{.Names}}\t{{.Status}}")
		if !result.OK {
			continue
		}
		for _, line := range strings.Split(result.Stdout, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || !hasMarker(strings.ToLower(line), d.facts.Markers) {
				continue
			}
			parts := strings.Split(line, "\t")
			ref := model.ContainerRef{Runtime: runtimeName}
			if len(parts) > 0 {
				ref.ID = parts[0]
			}
			if len(parts) > 1 {
				ref.Image = parts[1]
			}
			if len(parts) > 2 {
				ref.Name = parts[2]
			}
			if len(parts) > 3 {
				ref.Status = parts[3]
			}
			out = append(out, ref)
		}
	}
	return uniqContainers(out)
}

func (d Discoverer) discoverImages(ctx context.Context) []model.ImageRef {
	var out []model.ImageRef
	for _, runtimeName := range []string{"docker", "podman"} {
		if !d.runner.Exists(ctx, runtimeName) {
			continue
		}
		result := d.runner.Run(ctx, runtimeName, "images", "--format", "{{.Repository}}:{{.Tag}}\t{{.ID}}")
		if !result.OK {
			continue
		}
		for _, line := range strings.Split(result.Stdout, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || !hasMarker(strings.ToLower(line), d.facts.Markers) {
				continue
			}
			parts := strings.Split(line, "\t")
			ref := model.ImageRef{Runtime: runtimeName}
			if len(parts) > 0 {
				ref.Name = parts[0]
			}
			if len(parts) > 1 {
				ref.ID = parts[1]
			}
			out = append(out, ref)
		}
	}
	return uniqImages(out)
}

func uniqExisting(paths []string) []string {
	var out []string
	seen := map[string]struct{}{}
	for _, item := range paths {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		if _, err := os.Stat(item); err == nil {
			seen[item] = struct{}{}
			out = append(out, item)
		}
	}
	return out
}

func uniqExistingish(paths []string) []string {
	var out []string
	seen := map[string]struct{}{}
	for _, item := range paths {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	slices.Sort(out)
	return out
}

func uniqPackages(items []model.PackageRef) []model.PackageRef {
	seen := map[string]struct{}{}
	var out []model.PackageRef
	for _, item := range items {
		key := item.Manager + "|" + item.Kind + "|" + item.Name
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func uniqServices(items []model.ServiceRef) []model.ServiceRef {
	seen := map[string]struct{}{}
	var out []model.ServiceRef
	for _, item := range items {
		key := item.Platform + "|" + item.Scope + "|" + item.Name + "|" + item.Path
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func uniqContainers(items []model.ContainerRef) []model.ContainerRef {
	seen := map[string]struct{}{}
	var out []model.ContainerRef
	for _, item := range items {
		key := item.Runtime + "|" + item.ID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func uniqImages(items []model.ImageRef) []model.ImageRef {
	seen := map[string]struct{}{}
	var out []model.ImageRef
	for _, item := range items {
		key := item.Runtime + "|" + item.ID + "|" + item.Name
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func hasMarker(input string, markers []string) bool {
	for _, marker := range markers {
		if marker == "" {
			continue
		}
		if strings.Contains(marker, ".") {
			if strings.Contains(input, marker) {
				return true
			}
			continue
		}
		if containsWordLike(input, marker) {
			return true
		}
	}
	return false
}

func containsWordLike(input string, marker string) bool {
	start := 0
	for {
		idx := strings.Index(input[start:], marker)
		if idx < 0 {
			return false
		}
		idx += start
		beforeOK := idx == 0 || !isWordRune(rune(input[idx-1]))
		end := idx + len(marker)
		afterOK := end >= len(input) || !isWordRune(rune(input[end]))
		if beforeOK && afterOK {
			return true
		}
		start = idx + len(marker)
		if start >= len(input) {
			return false
		}
	}
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-'
}

func atoi(raw string) int {
	value := 0
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return value
		}
		value = value*10 + int(ch-'0')
	}
	return value
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := ""
	if n < 0 {
		neg = "-"
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return neg + string(buf[i:])
}

// fileContainsMarker reads the file at path and returns true if any line
// contains at least one marker from the given list.
func fileContainsMarker(path string, markers []string) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(raw), "\n") {
		lower := strings.ToLower(strings.TrimSpace(line))
		if lower == "" || strings.HasPrefix(lower, "#") {
			continue
		}
		if hasMarker(lower, markers) {
			return true
		}
	}
	return false
}

// dedupe returns a slice with duplicate strings removed, preserving order.
func dedupe(items []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, item := range items {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

// discoverRegistryKeys discovers Windows registry keys related to the product.
func (d Discoverer) discoverRegistryKeys(ctx context.Context) []model.RegistryRef {
	if d.host.OS != "windows" {
		return nil
	}

	var out []model.RegistryRef

	// Scan provider-defined registry paths
	for _, regPath := range d.facts.RegistryPaths {
		parts := strings.SplitN(regPath, "\\", 2)
		if len(parts) != 2 {
			continue
		}
		rootKey := parts[0]
		path := parts[1]

		cmd := d.adapter.RegistryQueryRecursiveCommand(rootKey, path)
		if len(cmd) == 0 || !d.runner.Exists(ctx, cmd[0]) {
			continue
		}

		result := d.runner.Run(ctx, cmd[0], cmd[1:]...)
		if !result.OK {
			continue
		}

		// Parse registry output and create refs
		out = append(out, parseRegistryOutput(result.Stdout, rootKey, path)...)
	}

	// Scan common locations for product markers
	commonPaths := []struct {
		rootKey string
		path    string
	}{
		{"HKCU", "Software"},
		{"HKLM", "SOFTWARE"},
		{"HKLM", "SOFTWARE\\WOW6432Node"},
	}

	for _, cp := range commonPaths {
		cmd := d.adapter.RegistryQueryCommand(cp.rootKey, cp.path)
		if len(cmd) == 0 || !d.runner.Exists(ctx, cmd[0]) {
			continue
		}

		result := d.runner.Run(ctx, cmd[0], cmd[1:]...)
		if !result.OK {
			continue
		}

		// Look for subkeys matching markers
		for _, line := range strings.Split(result.Stdout, "\n") {
			line = strings.TrimSpace(line)
			if hasMarker(strings.ToLower(line), d.facts.Markers) {
				// Extract the subkey name
				if idx := strings.LastIndex(line, "\\"); idx >= 0 {
					subkeyName := line[idx+1:]
					out = append(out, model.RegistryRef{
						RootKey: cp.rootKey,
						Path:    cp.path + "\\" + subkeyName,
					})
				}
			}
		}
	}

	return uniqRegistryRefs(out)
}

// parseRegistryOutput parses Windows reg query output into registry refs.
func parseRegistryOutput(output, rootKey, basePath string) []model.RegistryRef {
	var out []model.RegistryRef
	var currentPath string

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this is a key path line
		if strings.HasPrefix(line, rootKey+"\\") {
			currentPath = strings.TrimPrefix(line, rootKey+"\\")
			continue
		}

		// Check if this is a value line (format: "    ValueName    REG_TYPE    ValueData")
		if strings.HasPrefix(line, "    ") && currentPath != "" {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				valueName := fields[0]
				valueType := fields[1]
				var valueData string
				if len(fields) >= 3 {
					valueData = strings.Join(fields[2:], " ")
				}
				out = append(out, model.RegistryRef{
					RootKey: rootKey,
					Path:    currentPath,
					Value:   valueName,
					Type:    valueType,
					Data:    valueData,
				})
			}
		}
	}

	return out
}

// discoverEnvVars discovers environment variables related to the product.
func (d Discoverer) discoverEnvVars(ctx context.Context) []model.EnvVarRef {
	var out []model.EnvVarRef

	// Check provider-defined environment variable names
	for _, name := range d.facts.EnvVarNames {
		// Check user scope
		cmd := d.adapter.EnvGetCommand(name, false)
		if len(cmd) > 0 && d.runner.Exists(ctx, cmd[0]) {
			result := d.runner.Run(ctx, cmd[0], cmd[1:]...)
			if result.OK && result.Stdout != "" {
				value := strings.TrimSpace(result.Stdout)
				if value != "" && hasMarker(strings.ToLower(value), d.facts.Markers) {
					out = append(out, model.EnvVarRef{
						Name:  name,
						Value: value,
						Scope: "user",
					})
				}
			}
		}

		// Check system scope (Windows)
		if d.host.OS == "windows" {
			cmd := d.adapter.EnvGetCommand(name, true)
			if len(cmd) > 0 && d.runner.Exists(ctx, cmd[0]) {
				result := d.runner.Run(ctx, cmd[0], cmd[1:]...)
				if result.OK {
					// Parse registry output for value
					for _, line := range strings.Split(result.Stdout, "\n") {
						line = strings.TrimSpace(line)
						if strings.HasPrefix(line, name) {
							fields := strings.Fields(line)
							if len(fields) >= 3 {
								value := strings.Join(fields[2:], " ")
								if hasMarker(strings.ToLower(value), d.facts.Markers) {
									out = append(out, model.EnvVarRef{
										Name:  name,
										Value: value,
										Scope: "system",
									})
								}
							}
						}
					}
				}
			}
		}
	}

	// Also check PATH for product paths
	pathCmd := d.adapter.EnvGetCommand("PATH", false)
	if len(pathCmd) > 0 && d.runner.Exists(ctx, pathCmd[0]) {
		result := d.runner.Run(ctx, pathCmd[0], pathCmd[1:]...)
		if result.OK {
			pathValue := strings.TrimSpace(result.Stdout)
			for _, pathEntry := range strings.Split(pathValue, string(os.PathListSeparator)) {
				if hasMarker(strings.ToLower(pathEntry), d.facts.Markers) {
					out = append(out, model.EnvVarRef{
						Name:  "PATH",
						Value: pathEntry,
						Scope: "user",
					})
				}
			}
		}
	}

	return uniqEnvVarRefs(out)
}

// discoverHostsEntries discovers hosts file entries related to the product.
func (d Discoverer) discoverHostsEntries(ctx context.Context) []string {
	hostsPath := d.adapter.HostsFilePath()
	if hostsPath == "" {
		return nil
	}

	content, err := os.ReadFile(hostsPath)
	if err != nil {
		return nil
	}

	var out []string
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check if the line contains a marker
		if hasMarker(strings.ToLower(line), d.facts.Markers) {
			out = append(out, line)
		}
	}

	return dedupe(out)
}

func uniqRegistryRefs(items []model.RegistryRef) []model.RegistryRef {
	seen := map[string]struct{}{}
	var out []model.RegistryRef
	for _, item := range items {
		key := item.RootKey + "|" + item.Path + "|" + item.Value
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func uniqEnvVarRefs(items []model.EnvVarRef) []model.EnvVarRef {
	seen := map[string]struct{}{}
	var out []model.EnvVarRef
	for _, item := range items {
		key := item.Name + "|" + item.Scope + "|" + item.Value
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}
