package platform

import (
	"path/filepath"
	"runtime"
)

type Host struct {
	OS      string `json:"os"`
	Arch    string `json:"arch"`
	ExeExt  string `json:"exeExt,omitempty"`
	HomeEnv string `json:"homeEnv"`
}

func Detect() Host {
	host := Host{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
	if runtime.GOOS == "windows" {
		host.ExeExt = ".exe"
		host.HomeEnv = "USERPROFILE"
		return host
	}
	host.HomeEnv = "HOME"
	return host
}

func (h Host) Join(parts ...string) string {
	return filepath.Join(parts...)
}
