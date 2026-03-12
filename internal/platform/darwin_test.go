package platform

import (
	"reflect"
	"testing"

	"github.com/tianrking/ClawRemove/internal/model"
)

func TestDarwinAdapter_ServiceStatusCommand(t *testing.T) {
	adapter := darwinAdapter{}
	
	// Test without UID
	cmd1 := adapter.ServiceStatusCommand(model.ServiceRef{Name: "com.test.service"}, "")
	expect1 := []string{"launchctl", "print", "gui/0/com.test.service"}
	if !reflect.DeepEqual(cmd1, expect1) {
		t.Errorf("expected %v, got %v", expect1, cmd1)
	}

	// Test with UID
	cmd2 := adapter.ServiceStatusCommand(model.ServiceRef{Name: "com.test.service"}, "501")
	expect2 := []string{"launchctl", "print", "gui/501/com.test.service"}
	if !reflect.DeepEqual(cmd2, expect2) {
		t.Errorf("expected %v, got %v", expect2, cmd2)
	}
}

func TestDarwinAdapter_ServiceDisableCommand(t *testing.T) {
	adapter := darwinAdapter{}
	cmd := adapter.ServiceDisableCommand(model.ServiceRef{Name: "com.test.service"})
	expect := []string{"launchctl", "bootout", "gui/$UID/com.test.service"}
	if !reflect.DeepEqual(cmd, expect) {
		t.Errorf("expected %v, got %v", expect, cmd)
	}
}

func TestDarwinAdapter_ProcessCommands(t *testing.T) {
	adapter := darwinAdapter{}
	
	if listCmd := adapter.ProcessListCommand(); len(listCmd) == 0 {
		t.Error("ProcessListCommand returned empty")
	}

	if statusCmd := adapter.ProcessStatusCommand(-1); statusCmd != nil {
		t.Errorf("expected nil for negative PID, got %v", statusCmd)
	}
	
	if statusCmd := adapter.ProcessStatusCommand(0); statusCmd != nil {
		t.Errorf("expected nil for 0 PID, got %v", statusCmd)
	}

	validStatus := adapter.ProcessStatusCommand(123)
	expectStatus := []string{"ps", "-p", "123", "-o", "pid=,ppid=,etime=,command="}
	if !reflect.DeepEqual(validStatus, expectStatus) {
		t.Errorf("expected %v, got %v", expectStatus, validStatus)
	}

	if termCmd := adapter.ProcessTerminateCommand(-1); termCmd != nil {
		t.Errorf("expected nil for negative PID, got %v", termCmd)
	}

	validTerm := adapter.ProcessTerminateCommand(999)
	expectTerm := []string{"kill", "-TERM", "999"}
	if !reflect.DeepEqual(validTerm, expectTerm) {
		t.Errorf("expected %v, got %v", expectTerm, validTerm)
	}
}

func TestDarwinAdapter_NetworkCommands(t *testing.T) {
	adapter := darwinAdapter{}
	cmds := adapter.ListenerCommands()
	if len(cmds) != 1 {
		t.Errorf("expected 1 command sequence, got %d", len(cmds))
	}
	if !reflect.DeepEqual(cmds[0], []string{"lsof", "-nP", "-iTCP", "-sTCP:LISTEN"}) {
		t.Errorf("lsof command signature mismatch")
	}

	if tasks := adapter.ScheduledTaskListCommand(); tasks != nil {
		t.Errorf("expected nil scheduled task list for darwin, got %v", tasks)
	}
}
