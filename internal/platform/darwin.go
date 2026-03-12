package platform

import "github.com/tianrking/ClawRemove/internal/model"

type darwinAdapter struct{}

func (darwinAdapter) ServiceStatusCommand(service model.ServiceRef, uid string) []string {
	if uid == "" {
		uid = "0"
	}
	return []string{"launchctl", "print", "gui/" + uid + "/" + service.Name}
}

func (darwinAdapter) ProcessStatusCommand(pid int) []string {
	if pid <= 0 {
		return nil
	}
	return []string{"ps", "-p", itoa(pid), "-o", "pid=,ppid=,etime=,command="}
}
