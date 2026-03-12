package platform

import "github.com/tianrking/ClawRemove/internal/model"

type windowsAdapter struct{}

func (windowsAdapter) ServiceStatusCommand(service model.ServiceRef, _ string) []string {
	return []string{"schtasks", "/Query", "/TN", service.Name, "/V", "/FO", "LIST"}
}

func (windowsAdapter) ProcessStatusCommand(pid int) []string {
	if pid <= 0 {
		return nil
	}
	return []string{"tasklist", "/FI", "PID eq " + itoa(pid), "/V", "/FO", "CSV"}
}
