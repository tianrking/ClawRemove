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

func (linuxAdapter) ProcessStatusCommand(pid int) []string {
	if pid <= 0 {
		return nil
	}
	return []string{"ps", "-p", itoa(pid), "-o", "pid=,ppid=,etime=,command="}
}
