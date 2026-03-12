package model

type Options struct {
	Command       string
	Product       string
	DryRun        bool
	Yes           bool
	Quiet         bool
	JSON          bool
	AuditOnly     bool
	KeepCLI       bool
	KeepApp       bool
	KeepWorkspace bool
	KeepShell     bool
	KillProcesses bool
	RemoveDocker  bool
	SourceRoot    string
}

type ProductFacts struct {
	ID                string
	DisplayName       string
	StateDirNames     []string
	ConfigNames       []string
	Markers           []string
	ShellProfileGlobs []string
	TempPrefixes      []string
	AppPaths          []string
	CLIPaths          []string
	PackageRefs       []PackageRef
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
	Kind     ActionKind `json:"kind"`
	Target   string     `json:"target"`
	Reason   string     `json:"reason"`
	Command  []string   `json:"command,omitempty"`
	Platform string     `json:"platform,omitempty"`
	Risk     string     `json:"risk,omitempty"`
	HighRisk bool       `json:"highRisk,omitempty"`
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
	OK        bool      `json:"ok"`
	Product   string    `json:"product"`
	Command   string    `json:"command"`
	DryRun    bool      `json:"dryRun"`
	AuditOnly bool      `json:"auditOnly"`
	Discovery Discovery `json:"discovery"`
	Plan      Plan      `json:"plan"`
	Results   []Result  `json:"results"`
}
