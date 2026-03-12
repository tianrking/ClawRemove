package platform

import "github.com/tianrking/ClawRemove/internal/model"

type genericAdapter struct{}

func (genericAdapter) ServiceStatusCommand(service model.ServiceRef, _ string) []string {
	_ = service
	return nil
}

func (genericAdapter) ProcessStatusCommand(pid int) []string {
	if pid <= 0 {
		return nil
	}
	return []string{"ps", "-p", itoa(pid), "-o", "pid=,ppid=,etime=,command="}
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
