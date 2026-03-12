package platform

import "github.com/tianrking/ClawRemove/internal/model"

type genericAdapter struct{}

func (genericAdapter) ServiceStatusCommand(service model.ServiceRef, _ string) []string {
	_ = service
	return nil
}

func (genericAdapter) ServiceDisableCommand(service model.ServiceRef) []string {
	_ = service
	return nil
}

func (genericAdapter) ProcessListCommand() []string {
	return []string{"ps", "ax", "-o", "pid=,ppid=,command="}
}

func (genericAdapter) ProcessStatusCommand(pid int) []string {
	if pid <= 0 {
		return nil
	}
	return []string{"ps", "-p", itoa(pid), "-o", "pid=,ppid=,etime=,command="}
}

func (genericAdapter) ProcessTerminateCommand(pid int) []string {
	if pid <= 0 {
		return nil
	}
	return []string{"kill", "-TERM", itoa(pid)}
}

func (genericAdapter) ListenerCommands() [][]string {
	return nil
}

func (genericAdapter) ScheduledTaskListCommand() []string {
	return nil
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var buf [32]byte
	i := len(buf)
	for value > 0 {
		i--
		buf[i] = byte('0' + (value % 10))
		value /= 10
	}
	return string(buf[i:])
}
