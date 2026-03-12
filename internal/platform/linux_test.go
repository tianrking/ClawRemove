package platform

import (
	"reflect"
	"testing"

	"github.com/tianrking/ClawRemove/internal/model"
)

func TestLinuxAdapter_ServiceStatusCommand(t *testing.T) {
	adapter := linuxAdapter{}
	
	systemCmd := adapter.ServiceStatusCommand(model.ServiceRef{Name: "sshd", Scope: "system"}, "")
	expectSys := []string{"systemctl", "status", "sshd.service", "--no-pager"}
	if !reflect.DeepEqual(systemCmd, expectSys) {
		t.Errorf("expected %v, got %v", expectSys, systemCmd)
	}

	userCmd := adapter.ServiceStatusCommand(model.ServiceRef{Name: "docker", Scope: "user"}, "")
	expectUsr := []string{"systemctl", "--user", "status", "docker.service", "--no-pager"}
	if !reflect.DeepEqual(userCmd, expectUsr) {
		t.Errorf("expected %v, got %v", expectUsr, userCmd)
	}
}

func TestLinuxAdapter_ServiceDisableCommand(t *testing.T) {
	adapter := linuxAdapter{}
	
	systemCmd := adapter.ServiceDisableCommand(model.ServiceRef{Name: "nginx", Scope: "system"})
	expectSys := []string{"systemctl", "disable", "--now", "nginx.service"}
	if !reflect.DeepEqual(systemCmd, expectSys) {
		t.Errorf("expected %v, got %v", expectSys, systemCmd)
	}

	userCmd := adapter.ServiceDisableCommand(model.ServiceRef{Name: "x11vnc", Scope: "user"})
	expectUsr := []string{"systemctl", "--user", "disable", "--now", "x11vnc.service"}
	if !reflect.DeepEqual(userCmd, expectUsr) {
		t.Errorf("expected %v, got %v", expectUsr, userCmd)
	}
}

func TestLinuxAdapter_ProcessCommands(t *testing.T) {
	adapter := linuxAdapter{}
	
	if listCmd := adapter.ProcessListCommand(); len(listCmd) == 0 {
		t.Error("ProcessListCommand returned empty")
	}

	if statusCmd := adapter.ProcessStatusCommand(0); statusCmd != nil {
		t.Errorf("expected nil for 0 PID, got %v", statusCmd)
	}

	validStatus := adapter.ProcessStatusCommand(1234)
	expectStatus := []string{"ps", "-p", "1234", "-o", "pid=,ppid=,etime=,command="}
	if !reflect.DeepEqual(validStatus, expectStatus) {
		t.Errorf("expected %v, got %v", expectStatus, validStatus)
	}

	if termCmd := adapter.ProcessTerminateCommand(-10); termCmd != nil {
		t.Errorf("expected nil for negative PID, got %v", termCmd)
	}

	validTerm := adapter.ProcessTerminateCommand(5678)
	expectTerm := []string{"kill", "-TERM", "5678"}
	if !reflect.DeepEqual(validTerm, expectTerm) {
		t.Errorf("expected %v, got %v", expectTerm, validTerm)
	}
}

func TestLinuxAdapter_NetworkCommands(t *testing.T) {
	adapter := linuxAdapter{}
	cmds := adapter.ListenerCommands()
	if len(cmds) != 2 {
		t.Errorf("expected 2 command sequences, got %d", len(cmds))
	}
	if !reflect.DeepEqual(cmds[0], []string{"ss", "-lptn"}) {
		t.Errorf("ss command signature mismatch")
	}

	if tasks := adapter.ScheduledTaskListCommand(); tasks != nil {
		t.Errorf("expected nil scheduled task list for linux, got %v", tasks)
	}
}
