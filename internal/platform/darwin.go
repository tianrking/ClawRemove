package platform

import "github.com/tianrking/ClawRemove/internal/model"

type darwinAdapter struct{}

func (darwinAdapter) ServiceStatusCommand(service model.ServiceRef, uid string) []string {
	if uid == "" {
		uid = "0"
	}
	return []string{"launchctl", "print", "gui/" + uid + "/" + service.Name}
}

func (darwinAdapter) ServiceDisableCommand(service model.ServiceRef) []string {
	return []string{"launchctl", "bootout", "gui/$UID/" + service.Name}
}

func (darwinAdapter) ProcessListCommand() []string {
	return []string{"ps", "ax", "-o", "pid=,ppid=,command="}
}

func (darwinAdapter) ProcessStatusCommand(pid int) []string {
	if pid <= 0 {
		return nil
	}
	return []string{"ps", "-p", itoa(pid), "-o", "pid=,ppid=,etime=,command="}
}

func (darwinAdapter) ProcessTerminateCommand(pid int) []string {
	if pid <= 0 {
		return nil
	}
	return []string{"kill", "-TERM", itoa(pid)}
}

func (darwinAdapter) ListenerCommands() [][]string {
	return [][]string{{"lsof", "-nP", "-iTCP", "-sTCP:LISTEN"}}
}

func (darwinAdapter) ScheduledTaskListCommand() []string {
	return nil
}

// Registry methods - not applicable on macOS
func (darwinAdapter) RegistryQueryCommand(rootKey, path string) []string {
	return nil
}

func (darwinAdapter) RegistryQueryRecursiveCommand(rootKey, path string) []string {
	return nil
}

func (darwinAdapter) RegistryDeleteKeyCommand(rootKey, path string) []string {
	return nil
}

func (darwinAdapter) RegistryDeleteValueCommand(rootKey, path, value string) []string {
	return nil
}

func (darwinAdapter) EnvGetCommand(name string, systemScope bool) []string {
	// On macOS/Linux, we read environment from shell
	return []string{"sh", "-c", "echo \"$" + name + "\""}
}

func (darwinAdapter) HostsFilePath() string {
	return "/etc/hosts"
}
