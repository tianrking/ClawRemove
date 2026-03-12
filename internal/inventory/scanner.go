package inventory

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tianrking/ClawRemove/internal/platform"
	"github.com/tianrking/ClawRemove/internal/system"
)

// Scanner discovers AI runtimes, frameworks, and infrastructure.
type Scanner struct {
	runner system.Runner
	host   platform.Host
	home   string
}

// NewScanner creates a new AI inventory scanner.
func NewScanner(runner system.Runner, host platform.Host) *Scanner {
	home, _ := os.UserHomeDir()
	return &Scanner{
		runner: runner,
		host:   host,
		home:   home,
	}
}

// AIInventory contains discovered AI components.
type AIInventory struct {
	// Runtimes are local LLM runtimes like Ollama, LM Studio.
	Runtimes []RuntimeInfo `json:"runtimes"`
	// Frameworks are AI SDKs and libraries.
	Frameworks []FrameworkInfo `json:"frameworks"`
	// ModelCaches are cached model files.
	ModelCaches []ModelCacheInfo `json:"modelCaches"`
	// VectorStores are vector database installations.
	VectorStores []VectorStoreInfo `json:"vectorStores"`
	// Agents are AI agent installations.
	Agents []AgentInfo `json:"agents"`
	// Services are AI-related system services.
	Services []ServiceInfo `json:"services"`
}

// RuntimeInfo represents a local LLM runtime.
type RuntimeInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version,omitempty"`
	Path      string `json:"path,omitempty"`
	Running   bool   `json:"running"`
	Ports     []int  `json:"ports,omitempty"`
	ModelCount int   `json:"modelCount,omitempty"`
}

// FrameworkInfo represents an AI framework/SDK.
type FrameworkInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Path    string `json:"path,omitempty"`
	Manager string `json:"manager,omitempty"` // npm, pip, etc.
}

// ModelCacheInfo represents cached model files.
type ModelCacheInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size,omitempty"`
	Type string `json:"type,omitempty"` // llama, bert, etc.
}

// VectorStoreInfo represents a vector database.
type VectorStoreInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path,omitempty"`
	Running bool   `json:"running"`
	Port    int    `json:"port,omitempty"`
}

// AgentInfo represents an AI agent installation.
type AgentInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Version string `json:"version,omitempty"`
}

// ServiceInfo represents an AI-related system service.
type ServiceInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // systemd, launchd, cron
	Running bool   `json:"running"`
}

// Scan performs a full AI inventory scan.
func (s *Scanner) Scan(ctx context.Context) AIInventory {
	return AIInventory{
		Runtimes:    s.scanRuntimes(ctx),
		Frameworks:  s.scanFrameworks(ctx),
		ModelCaches: s.scanModelCaches(),
		VectorStores: s.scanVectorStores(ctx),
		Agents:      s.scanAgents(),
		Services:    s.scanServices(ctx),
	}
}

// scanRuntimes detects local LLM runtimes.
func (s *Scanner) scanRuntimes(ctx context.Context) []RuntimeInfo {
	var runtimes []RuntimeInfo

	// Ollama
	if ollama := s.detectOllama(ctx); ollama != nil {
		runtimes = append(runtimes, *ollama)
	}

	// LM Studio
	if lmstudio := s.detectLMStudio(); lmstudio != nil {
		runtimes = append(runtimes, *lmstudio)
	}

	// GPT4All
	if gpt4all := s.detectGPT4All(); gpt4all != nil {
		runtimes = append(runtimes, *gpt4all)
	}

	// LocalAI
	if localai := s.detectLocalAI(ctx); localai != nil {
		runtimes = append(runtimes, *localai)
	}

	return runtimes
}

func (s *Scanner) detectOllama(ctx context.Context) *RuntimeInfo {
	// Check for ollama binary
	if !s.runner.Exists(ctx, "ollama") {
		return nil
	}

	info := &RuntimeInfo{
		Name:    "Ollama",
		Running: false,
	}

	// Check if ollama service is running
	result := s.runner.Run(ctx, "pgrep", "-f", "ollama")
	info.Running = result.OK && result.Stdout != ""

	// Get version
	result = s.runner.Run(ctx, "ollama", "--version")
	if result.OK {
		info.Version = strings.TrimSpace(result.Stdout)
	}

	// Check models directory
	modelsDir := filepath.Join(s.home, ".ollama", "models")
	if files, err := os.ReadDir(modelsDir); err == nil {
		info.ModelCount = countModelFiles(files)
	}

	info.Path = modelsDir
	info.Ports = []int{11434}

	return info
}

func (s *Scanner) detectLMStudio() *RuntimeInfo {
	// Check for LM Studio directories
	lmstudioPaths := []string{
		filepath.Join(s.home, ".lmstudio"),
		filepath.Join(s.home, "Library", "Application Support", "LM Studio"),
		filepath.Join(s.home, "AppData", "Roaming", "LM Studio"),
	}

	for _, p := range lmstudioPaths {
		if _, err := os.Stat(p); err == nil {
			return &RuntimeInfo{
				Name:  "LM Studio",
				Path:  p,
				Ports: []int{1234},
			}
		}
	}
	return nil
}

func (s *Scanner) detectGPT4All() *RuntimeInfo {
	gpt4allPaths := []string{
		filepath.Join(s.home, ".gpt4all"),
		filepath.Join(s.home, "Library", "Application Support", "gpt4all"),
		filepath.Join(s.home, "AppData", "Roaming", "nomic.ai", "GPT4All"),
	}

	for _, p := range gpt4allPaths {
		if _, err := os.Stat(p); err == nil {
			return &RuntimeInfo{
				Name: "GPT4All",
				Path: p,
			}
		}
	}
	return nil
}

func (s *Scanner) detectLocalAI(ctx context.Context) *RuntimeInfo {
	if !s.runner.Exists(ctx, "local-ai") {
		return nil
	}

	info := &RuntimeInfo{
		Name:    "LocalAI",
		Running: false,
	}

	result := s.runner.Run(ctx, "pgrep", "-f", "local-ai")
	info.Running = result.OK && result.Stdout != ""

	info.Ports = []int{8080}
	return info
}

// scanFrameworks detects AI frameworks and SDKs.
func (s *Scanner) scanFrameworks(ctx context.Context) []FrameworkInfo {
	var frameworks []FrameworkInfo

	// Python packages
	pythonPkgs := []string{
		"langchain", "openai", "anthropic", "llama-index",
		"transformers", "torch", "tensorflow", "keras",
		"chromadb", "pinecone-client", "weaviate-client",
		"tiktoken", "sentence-transformers",
	}

	for _, pkg := range pythonPkgs {
		if s.runner.Exists(ctx, "pip") || s.runner.Exists(ctx, "pip3") {
			pipCmd := "pip"
			if s.runner.Exists(ctx, "pip3") {
				pipCmd = "pip3"
			}
			result := s.runner.Run(ctx, pipCmd, "show", pkg)
			if result.OK {
				version := extractVersion(result.Stdout)
				frameworks = append(frameworks, FrameworkInfo{
					Name:    pkg,
					Version: version,
					Manager: "pip",
				})
			}
		}
	}

	// Node.js packages
	nodePkgs := []string{
		"openai", "@anthropic-ai/sdk", "langchain",
		"@langchain/core", "@pinecone-database/pinecone",
		"chromadb", "tiktoken",
	}

	// Check global npm packages
	if s.runner.Exists(ctx, "npm") {
		result := s.runner.Run(ctx, "npm", "list", "-g", "--depth=0")
		if result.OK {
			for _, pkg := range nodePkgs {
				if strings.Contains(result.Stdout, pkg) {
					frameworks = append(frameworks, FrameworkInfo{
						Name:    pkg,
						Manager: "npm",
					})
				}
			}
		}
	}

	return frameworks
}

// scanModelCaches detects cached model files.
func (s *Scanner) scanModelCaches() []ModelCacheInfo {
	var caches []ModelCacheInfo

	cacheDirs := []struct {
		name string
		path string
	}{
		{"Hugging Face", filepath.Join(s.home, ".cache", "huggingface")},
		{"Torch Hub", filepath.Join(s.home, ".cache", "torch")},
		{"TensorFlow Models", filepath.Join(s.home, ".cache", "tensorflow")},
		{"Ollama Models", filepath.Join(s.home, ".ollama", "models")},
		{"LM Studio Models", filepath.Join(s.home, ".lmstudio", "models")},
		{"GPT4All Models", filepath.Join(s.home, ".cache", "gpt4all")},
	}

	for _, cache := range cacheDirs {
		if info, err := os.Stat(cache.path); err == nil && info.IsDir() {
			size := dirSize(cache.path)
			caches = append(caches, ModelCacheInfo{
				Name: cache.name,
				Path: cache.path,
				Size: size,
			})
		}
	}

	return caches
}

// scanVectorStores detects vector database installations.
func (s *Scanner) scanVectorStores(ctx context.Context) []VectorStoreInfo {
	var stores []VectorStoreInfo

	// ChromaDB
	chromaPaths := []string{
		filepath.Join(s.home, ".chromadb"),
		filepath.Join(s.home, ".local", "share", "chromadb"),
	}
	for _, p := range chromaPaths {
		if _, err := os.Stat(p); err == nil {
			stores = append(stores, VectorStoreInfo{
				Name: "ChromaDB",
				Path: p,
			})
		}
	}

	// Check for running vector DB services
	if s.runner.Exists(ctx, "docker") {
		result := s.runner.Run(ctx, "docker", "ps", "--format", "{{.Names}}")
		if result.OK {
			for _, line := range strings.Split(result.Stdout, "\n") {
				line = strings.ToLower(line)
				if strings.Contains(line, "chroma") {
					stores = append(stores, VectorStoreInfo{
						Name:    "ChromaDB (Docker)",
						Running: true,
					})
				}
				if strings.Contains(line, "pinecone") {
					stores = append(stores, VectorStoreInfo{
						Name:    "Pinecone (Docker)",
						Running: true,
					})
				}
				if strings.Contains(line, "weaviate") {
					stores = append(stores, VectorStoreInfo{
						Name:    "Weaviate (Docker)",
						Running: true,
					})
				}
				if strings.Contains(line, "qdrant") {
					stores = append(stores, VectorStoreInfo{
						Name:    "Qdrant (Docker)",
						Running: true,
					})
				}
			}
		}
	}

	return stores
}

// scanAgents detects installed AI agents.
func (s *Scanner) scanAgents() []AgentInfo {
	var agents []AgentInfo

	agentDirs := []struct {
		name string
		path string
	}{
		{"OpenClaw", filepath.Join(s.home, ".openclaw")},
		{"NanoBot", filepath.Join(s.home, ".nanobot")},
		{"PicoClaw", filepath.Join(s.home, ".picoclaw")},
		{"OpenFang", filepath.Join(s.home, ".openfang")},
		{"ZeroClaw", filepath.Join(s.home, ".zeroclaw")},
		{"NanoClaw", filepath.Join(s.home, ".nanoclaw")},
		{"AutoGPT", filepath.Join(s.home, ".autogpt")},
		{"Cursor", filepath.Join(s.home, ".cursor")},
		{"Aider", filepath.Join(s.home, ".aider")},
		{"Cline", filepath.Join(s.home, ".cline")},
	}

	for _, agent := range agentDirs {
		if _, err := os.Stat(agent.path); err == nil {
			agents = append(agents, AgentInfo{
				Name: agent.name,
				Path: agent.path,
			})
		}
	}

	return agents
}

// scanServices detects AI-related system services.
func (s *Scanner) scanServices(ctx context.Context) []ServiceInfo {
	var services []ServiceInfo

	switch s.host.OS {
	case "darwin":
		services = append(services, s.scanDarwinServices(ctx)...)
	case "linux":
		services = append(services, s.scanLinuxServices(ctx)...)
	case "windows":
		services = append(services, s.scanWindowsServices(ctx)...)
	}

	return services
}

func (s *Scanner) scanDarwinServices(ctx context.Context) []ServiceInfo {
	var services []ServiceInfo

	// Check launchd services
	result := s.runner.Run(ctx, "launchctl", "list")
	if result.OK {
		aiServices := []string{"ollama", "openclaw", "nanobot", "localai", "chromadb"}
		for _, svc := range aiServices {
			if strings.Contains(strings.ToLower(result.Stdout), svc) {
				services = append(services, ServiceInfo{
					Name:    svc,
					Type:    "launchd",
					Running: true,
				})
			}
		}
	}

	return services
}

func (s *Scanner) scanLinuxServices(ctx context.Context) []ServiceInfo {
	var services []ServiceInfo

	if s.runner.Exists(ctx, "systemctl") {
		aiServices := []string{"ollama", "openclaw", "nanobot", "localai", "chromadb"}
		for _, svc := range aiServices {
			result := s.runner.Run(ctx, "systemctl", "is-active", svc)
			if result.OK && strings.TrimSpace(result.Stdout) == "active" {
				services = append(services, ServiceInfo{
					Name:    svc,
					Type:    "systemd",
					Running: true,
				})
			}
		}
	}

	return services
}

func (s *Scanner) scanWindowsServices(ctx context.Context) []ServiceInfo {
	var services []ServiceInfo

	// Check Windows services via PowerShell
	if s.runner.Exists(ctx, "powershell") {
		aiServices := []string{"Ollama", "LocalAI", "ChromaDB"}
		for _, svc := range aiServices {
			result := s.runner.Run(ctx, "powershell", "-Command",
				"Get-Service -Name '*"+svc+"*' -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Status")
			if result.OK && strings.Contains(result.Stdout, "Running") {
				services = append(services, ServiceInfo{
					Name:    svc,
					Type:    "windows-service",
					Running: true,
				})
			}
		}
	}

	return services
}

// Helper functions

func countModelFiles(files []os.DirEntry) int {
	count := 0
	for _, f := range files {
		if !f.IsDir() {
			name := strings.ToLower(f.Name())
			if strings.HasSuffix(name, ".gguf") ||
				strings.HasSuffix(name, ".bin") ||
				strings.HasSuffix(name, ".safetensors") {
				count++
			}
		}
	}
	return count
}

func extractVersion(output string) string {
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "Version:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
		}
	}
	return ""
}

func dirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
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

// init ensures runtime package is referenced
var _ = runtime.GOOS
