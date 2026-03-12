package platform

import "github.com/tianrking/ClawRemove/internal/model"

type linuxAdapter struct{}

func (linuxAdapter) ServiceStatusCommand(service model.ServiceRef, _ string) []string {
	args := []string{"status", service.Name + ".service", "--no-pager"}
	if service.Scope == "user" {
		args = append([]string{"--user"}, args...)
	}
	return append([]string{"systemctl"}, args...)
}

func (linuxAdapter) ServiceDisableCommand(service model.ServiceRef) []string {
	args := []string{"disable", "--now", service.Name + ".service"}
	if service.Scope == "user" {
		args = append([]string{"--user"}, args...)
	}
	return append([]string{"systemctl"}, args...)
}

func (linuxAdapter) ProcessListCommand() []string {
	return []string{"ps", "ax", "-o", "pid=,ppid=,command="}
}

func (linuxAdapter) ProcessStatusCommand(pid int) []string {
	if pid <= 0 {
		return nil
	}
	return []string{"ps", "-p", itoa(pid), "-o", "pid=,ppid=,etime=,command="}
}

func (linuxAdapter) ProcessTerminateCommand(pid int) []string {
	if pid <= 0 {
		return nil
	}
	return []string{"kill", "-TERM", itoa(pid)}
}

func (linuxAdapter) ListenerCommands() [][]string {
	return [][]string{
		{"ss", "-lptn"},
		{"netstat", "-lntp"},
	}
}

func (linuxAdapter) ScheduledTaskListCommand() []string {
	return nil
}

// Registry methods - not applicable on Linux
func (linuxAdapter) RegistryQueryCommand(rootKey, path string) []string {
	return nil
}

func (linuxAdapter) RegistryQueryRecursiveCommand(rootKey, path string) []string {
	return nil
}

func (linuxAdapter) RegistryDeleteKeyCommand(rootKey, path string) []string {
	return nil
}

func (linuxAdapter) RegistryDeleteValueCommand(rootKey, path, value string) []string {
	return nil
}

func (linuxAdapter) EnvGetCommand(name string, systemScope bool) []string {
	// On Linux, we read environment from shell
	return []string{"sh", "-c", "echo \"$" + name + "\""}
}

func (linuxAdapter) HostsFilePath() string {
	return "/etc/hosts"
}
