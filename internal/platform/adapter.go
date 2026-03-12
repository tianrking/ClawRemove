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
	// Registry methods (Windows only, return nil on other platforms)
	RegistryQueryCommand(rootKey, path string) []string
	RegistryQueryRecursiveCommand(rootKey, path string) []string
	RegistryDeleteKeyCommand(rootKey, path string) []string
	RegistryDeleteValueCommand(rootKey, path, value string) []string
	// Environment methods
	EnvGetCommand(name string, systemScope bool) []string
	// Hosts file
	HostsFilePath() string
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
