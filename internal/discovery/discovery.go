package discovery

import (
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"unicode"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/system"
)

type Discoverer struct {
	runner system.Runner
	facts  model.ProductFacts
}

func New(runner system.Runner, facts model.ProductFacts) Discoverer {
	return Discoverer{runner: runner, facts: facts}
}

func (d Discoverer) Discover(ctx context.Context) (model.Discovery, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return model.Discovery{}, err
	}

	stateDirs := d.discoverStateDirs(home)
	return model.Discovery{
		Platform:      runtime.GOOS,
		HomeDir:       home,
		StateDirs:     stateDirs,
		WorkspaceDirs: d.discoverWorkspaces(home, stateDirs),
		TempPaths:     d.discoverTempPaths(),
		ShellProfiles: d.discoverShellProfiles(home),
		AppPaths:      d.discoverAppPaths(home),
		CLIPaths:      d.discoverCLIPaths(home),
		Packages:      d.discoverPackages(ctx),
		Services:      d.discoverServices(ctx, home),
		Processes:     d.discoverProcesses(ctx),
		Listeners:     d.discoverListeners(ctx),
		CrontabLines:  d.discoverCrontab(ctx),
		Containers:    d.discoverContainers(ctx),
		Images:        d.discoverImages(ctx),
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
	out := []string{filepath.Join(home, ".openclaw", "workspace")}
	if entries, err := os.ReadDir(filepath.Join(home, ".openclaw")); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && (entry.Name() == "workspace" || strings.HasPrefix(entry.Name(), "workspace-")) {
				out = append(out, filepath.Join(home, ".openclaw", entry.Name()))
			}
		}
	}

	for _, dir := range stateDirs {
		base := filepath.Base(dir)
		if strings.HasPrefix(base, ".openclaw-") {
			profile := strings.TrimPrefix(base, ".openclaw-")
			out = append(out, filepath.Join(home, ".openclaw", "workspace-"+profile))
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
		out = append(out, filepath.Join(home, filepath.FromSlash(rel)))
	}
	return uniqExisting(out)
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
			if runtime.GOOS != "darwin" {
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
	switch runtime.GOOS {
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
	}
	for _, item := range dirs {
		entries, _ := os.ReadDir(item.dir)
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".service") {
				continue
			}
			name := strings.TrimSuffix(entry.Name(), ".service")
			if hasMarker(strings.ToLower(name), d.facts.Markers) {
				out = append(out, model.ServiceRef{
					Platform: "linux",
					Scope:    item.scope,
					Name:     name,
					Path:     filepath.Join(item.dir, entry.Name()),
				})
			}
		}
	}
	return uniqServices(out)
}

func (d Discoverer) discoverWindowsServices(ctx context.Context) []model.ServiceRef {
	if !d.runner.Exists(ctx, "schtasks") {
		return nil
	}
	result := d.runner.Run(ctx, "schtasks", "/Query", "/FO", "LIST", "/V")
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
	if runtime.GOOS == "windows" {
		result := d.runner.Run(ctx, "tasklist", "/V", "/FO", "CSV")
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

	result := d.runner.Run(ctx, "ps", "ax", "-o", "pid=,ppid=,command=")
	if !result.OK {
		return nil
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
	candidates := [][]string{}
	switch runtime.GOOS {
	case "darwin":
		candidates = append(candidates, []string{"lsof", "-nP", "-iTCP", "-sTCP:LISTEN"})
	case "linux":
		candidates = append(candidates, []string{"ss", "-lptn"}, []string{"netstat", "-lntp"})
	case "windows":
		candidates = append(candidates, []string{"netstat", "-ano"})
	}

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
			if line != "" && (hasMarker(lower, d.facts.Markers) || strings.Contains(line, "18789") || strings.Contains(line, "19001")) {
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
	if runtime.GOOS == "windows" || !d.runner.Exists(ctx, "crontab") {
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
