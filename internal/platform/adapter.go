package platform

import "github.com/tianrking/ClawRemove/internal/model"

type Adapter interface {
	ServiceStatusCommand(service model.ServiceRef, uid string) []string
	ServiceDisableCommand(service model.ServiceRef) []string
	ProcessListCommand() []string
	ProcessStatusCommand(pid int) []string
	ProcessTerminateCommand(pid int) []string
	ListenerCommands() [][]string
	ScheduledTaskListCommand() []string
}

func NewAdapter(host Host) Adapter {
	switch host.OS {
	case "darwin":
		return darwinAdapter{}
	case "linux":
		return linuxAdapter{}
	case "windows":
		return windowsAdapter{}
	default:
		return genericAdapter{}
	}
}
