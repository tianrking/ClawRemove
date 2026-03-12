package model

type Options struct {
	Command       string
	Product       string
	DryRun        bool
	Yes           bool
	Quiet         bool
	JSON          bool
	AI            bool
	AuditOnly     bool
	KeepCLI       bool
	KeepApp       bool
	KeepWorkspace bool
	KeepShell     bool
	KillProcesses bool
	RemoveDocker  bool
	Version       bool
	SourceRoot    string
}

type Host struct {
	OS      string `json:"os"`
	Arch    string `json:"arch"`
	ExeExt  string `json:"exeExt,omitempty"`
	HomeEnv string `json:"homeEnv,omitempty"`
}

type ProductFacts struct {
	ID                string
	DisplayName       string
	StateDirNames     []string
	WorkspaceDirNames []string
	ConfigNames       []string
	Markers           []string
	ShellProfileGlobs []string
	TempPrefixes      []string
	AppPaths          []string
	CLIPaths          []string
	PackageRefs       []PackageRef
	ListenerPorts     []int
	// Windows-specific
	RegistryPaths []string // Registry paths to scan (e.g., "HKCU\\Software\\OpenClaw")
	// Environment
	EnvVarNames []string // Environment variable names to check (e.g., "OPENCLAW_PATH")
}

type PackageRef struct {
	Manager string `json:"manager"`
	Name    string `json:"name"`
	Kind    string `json:"kind,omitempty"`
}

type ServiceRef struct {
	Platform string `json:"platform"`
	Scope    string `json:"scope"`
	Name     string `json:"name"`
	Path     string `json:"path,omitempty"`
}

type ProcessRef struct {
	PID     int    `json:"pid,omitempty"`
	PPID    int    `json:"ppid,omitempty"`
	Command string `json:"command"`
}

type ContainerRef struct {
	Runtime string `json:"runtime"`
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Image   string `json:"image,omitempty"`
	Status  string `json:"status,omitempty"`
}

type ImageRef struct {
	Runtime string `json:"runtime"`
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
}

// RegistryRef represents a Windows registry key or value.
type RegistryRef struct {
	RootKey string `json:"rootKey"` // HKLM, HKCU, HKCR, HKU, HKCC
	Path    string `json:"path"`    // Subkey path
	Value   string `json:"value,omitempty"`   // Value name (empty for default)
	Data    string `json:"data,omitempty"`    // Value data
	Type    string `json:"type,omitempty"`    // REG_SZ, REG_DWORD, etc.
}

// EnvVarRef represents an environment variable reference.
type EnvVarRef struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Scope string `json:"scope,omitempty"` // "user" or "system" (Windows)
}

type Discovery struct {
	Platform      string         `json:"platform"`
	HomeDir       string         `json:"homeDir"`
	StateDirs     []string       `json:"stateDirs"`
	WorkspaceDirs []string       `json:"workspaceDirs"`
	TempPaths     []string       `json:"tempPaths"`
	ShellProfiles []string       `json:"shellProfiles"`
	AppPaths      []string       `json:"appPaths"`
	CLIPaths      []string       `json:"cliPaths"`
	Packages      []PackageRef   `json:"packages"`
	Services      []ServiceRef   `json:"services"`
	Processes     []ProcessRef   `json:"processes"`
	Listeners     []string       `json:"listeners"`
	CrontabLines  []string       `json:"crontabLines"`
	Containers    []ContainerRef `json:"containers"`
	Images        []ImageRef     `json:"images"`
	RegistryKeys  []RegistryRef  `json:"registryKeys,omitempty"`  // Windows only
	EnvVars       []EnvVarRef    `json:"envVars,omitempty"`       // Environment variables
	HostsEntries  []string       `json:"hostsEntries,omitempty"`  // Hosts file entries
}

type ActionKind string

const (
	ActionRemovePath    ActionKind = "remove_path"
	ActionRunCommand    ActionKind = "run_command"
	ActionEditFile      ActionKind = "edit_file"
	ActionReportOnly    ActionKind = "report_only"
	ActionRemoveRegistry ActionKind = "remove_registry" // Windows registry key/value removal
)

type Action struct {
	Kind       ActionKind `json:"kind"`
	Target     string     `json:"target"`
	Reason     string     `json:"reason"`
	Evidence   string     `json:"evidence,omitempty"`
	Rule       string     `json:"rule,omitempty"`
	Source     string     `json:"source,omitempty"`
	Confidence float64    `json:"confidence,omitempty"`
	Command    []string   `json:"command,omitempty"`
	Platform   string     `json:"platform,omitempty"`
	Risk       string     `json:"risk,omitempty"`
	HighRisk   bool       `json:"highRisk,omitempty"`
	Markers    []string   `json:"markers,omitempty"`
}

type Plan struct {
	Actions []Action `json:"actions"`
}

type Result struct {
	OK      bool   `json:"ok"`
	Action  string `json:"action"`
	Target  string `json:"target"`
	Reason  string `json:"reason"`
	DryRun  bool   `json:"dryRun,omitempty"`
	Skipped bool   `json:"skipped,omitempty"`
	Error   string `json:"error,omitempty"`
}

type Report struct {
	OK           bool                 `json:"ok"`
	Product      string               `json:"product"`
	Command      string               `json:"command"`
	DryRun       bool                 `json:"dryRun"`
	AuditOnly    bool                 `json:"auditOnly"`
	Host         Host                 `json:"host"`
	Capabilities ProviderCapabilities `json:"capabilities"`
	Discovery    Discovery            `json:"discovery"`
	Evidence     EvidenceSet          `json:"evidence"`
	Verify       Verification         `json:"verify"`
	Plan         Plan                 `json:"plan"`
	Results      []Result             `json:"results"`
	Advice       *Advice              `json:"advice,omitempty"`
}

type Residual struct {
	Kind       string  `json:"kind"`
	Target     string  `json:"target"`
	Evidence   string  `json:"evidence"`
	Reason     string  `json:"reason"`
	Risk       string  `json:"risk,omitempty"`
	Rule       string  `json:"rule,omitempty"`
	Source     string  `json:"source,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

type VerificationSummary struct {
	Exact     int `json:"exact"`
	Strong    int `json:"strong"`
	Heuristic int `json:"heuristic"`
}

type Verification struct {
	Verified    bool                `json:"verified"`
	Summary     VerificationSummary `json:"summary"`
	Residuals   []Residual          `json:"residuals"`
	Confirmed   []Residual          `json:"confirmed"`
	Investigate []Residual          `json:"investigate"`
}

type Recommendation struct {
	Kind     string `json:"kind"`
	Target   string `json:"target"`
	Reason   string `json:"reason"`
	Risk     string `json:"risk"`
	OptIn    bool   `json:"optIn"`
	Evidence string `json:"evidence"`
}

type Advice struct {
	Mode            string           `json:"mode"`
	Authority       string           `json:"authority"`
	ThoughtSummary  string           `json:"thoughtSummary"`
	NeededEvidence  []string         `json:"neededEvidence"`
	Recommendations []Recommendation `json:"recommendations"`
	RiskNotes       []string         `json:"riskNotes"`
	Trace           []string         `json:"trace,omitempty"`
	UserMessage     string           `json:"userMessage"`
}

// EnvironmentReport is the output of the environment audit command.
type EnvironmentReport struct {
	Platform  string           `json:"platform"`
	Host      Host             `json:"host"`
	Runtime   RuntimeSection   `json:"runtime"`
	Agents    AgentsSection    `json:"agents"`
	Artifacts ArtifactsSection `json:"artifacts"`
	Security  SecuritySection  `json:"security"`
	Hygiene   HygieneSection   `json:"hygiene"`
}

// RuntimeSection contains detected AI runtimes.
type RuntimeSection struct {
	Detected []RuntimeItem `json:"detected"`
	Summary  string        `json:"summary"`
}

// RuntimeItem represents a detected AI runtime.
type RuntimeItem struct {
	Name       string `json:"name"`
	Version    string `json:"version,omitempty"`
	Path       string `json:"path,omitempty"`
	Running    bool   `json:"running"`
	ModelsSize int64  `json:"modelsSize,omitempty"`
	Port       int    `json:"port,omitempty"`
}

// AgentsSection contains detected agent tools and frameworks.
type AgentsSection struct {
	Applications []AgentItem `json:"applications"`
	Frameworks   []AgentItem `json:"frameworks"`
	Summary      string      `json:"summary"`
}

// AgentItem represents a detected agent or framework.
type AgentItem struct {
	Name    string `json:"name"`
	Path    string `json:"path,omitempty"`
	Version string `json:"version,omitempty"`
	Manager string `json:"manager,omitempty"` // npm, pip, etc.
}

// ArtifactsSection contains detected AI artifacts.
type ArtifactsSection struct {
	Models     []ArtifactItem `json:"models"`
	VectorDBs  []ArtifactItem `json:"vectorDbs"`
	Caches     []ArtifactItem `json:"caches"`
	TotalSize  int64          `json:"totalSize"`
	Summary    string         `json:"summary"`
}

// ArtifactItem represents a detected artifact.
type ArtifactItem struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size,omitempty"`
}

// SecuritySection contains security findings.
type SecuritySection struct {
	Findings    []SecurityFinding `json:"findings"`
	HighRisk    int               `json:"highRisk"`
	Summary     string            `json:"summary"`
}

// SecurityFinding represents a security issue.
type SecurityFinding struct {
	Type        string `json:"type"`
	Provider    string `json:"provider"`
	Location    string `json:"location"`
	Line        int    `json:"line,omitempty"`
	Severity    string `json:"severity"`
	Remediation string `json:"remediation"`
}

// HygieneSection contains storage analysis.
type HygieneSection struct {
	ModelsSize    int64  `json:"modelsSize"`
	CacheSize     int64  `json:"cacheSize"`
	VectorDBSize  int64  `json:"vectorDbSize"`
	LogSize       int64  `json:"logSize"`
	TotalSize     int64  `json:"totalSize"`
	Recommendations []string `json:"recommendations"`
	Summary       string `json:"summary"`
}
