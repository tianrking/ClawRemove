package platform

import "github.com/tianrking/ClawRemove/internal/model"

type windowsAdapter struct{}

func (windowsAdapter) ServiceStatusCommand(service model.ServiceRef, _ string) []string {
	return []string{"schtasks", "/Query", "/TN", service.Name, "/V", "/FO", "LIST"}
}

func (windowsAdapter) ServiceDisableCommand(service model.ServiceRef) []string {
	return []string{"schtasks", "/Delete", "/F", "/TN", service.Name}
}

func (windowsAdapter) ProcessListCommand() []string {
	return []string{"tasklist", "/V", "/FO", "CSV"}
}

func (windowsAdapter) ProcessStatusCommand(pid int) []string {
	if pid <= 0 {
		return nil
	}
	return []string{"tasklist", "/FI", "PID eq " + itoa(pid), "/V", "/FO", "CSV"}
}

func (windowsAdapter) ProcessTerminateCommand(pid int) []string {
	if pid <= 0 {
		return nil
	}
	return []string{"taskkill", "/PID", itoa(pid), "/F"}
}

func (windowsAdapter) ListenerCommands() [][]string {
	return [][]string{{"netstat", "-ano"}}
}

func (windowsAdapter) ScheduledTaskListCommand() []string {
	return []string{"schtasks", "/Query", "/FO", "LIST", "/V"}
}
