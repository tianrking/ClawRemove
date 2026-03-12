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
}

type ActionKind string

const (
	ActionRemovePath ActionKind = "remove_path"
	ActionRunCommand ActionKind = "run_command"
	ActionEditFile   ActionKind = "edit_file"
	ActionReportOnly ActionKind = "report_only"
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
