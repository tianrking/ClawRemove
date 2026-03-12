package platform

import (
	"reflect"
	"testing"

	"github.com/tianrking/ClawRemove/internal/model"
)

func TestWindowsAdapter_ServiceCommands(t *testing.T) {
	adapter := windowsAdapter{}
	
	statusCmd := adapter.ServiceStatusCommand(model.ServiceRef{Name: "OpenClawSvc"}, "")
	expectStatus := []string{"schtasks", "/Query", "/TN", "OpenClawSvc", "/V", "/FO", "LIST"}
	if !reflect.DeepEqual(statusCmd, expectStatus) {
		t.Errorf("expected %v, got %v", expectStatus, statusCmd)
	}

	disableCmd := adapter.ServiceDisableCommand(model.ServiceRef{Name: "OpenClawUpdater"})
	expectDisable := []string{"schtasks", "/Delete", "/F", "/TN", "OpenClawUpdater"}
	if !reflect.DeepEqual(disableCmd, expectDisable) {
		t.Errorf("expected %v, got %v", expectDisable, disableCmd)
	}
}

func TestWindowsAdapter_ProcessCommands(t *testing.T) {
	adapter := windowsAdapter{}
	
	listCmd := adapter.ProcessListCommand()
	expectList := []string{"tasklist", "/V", "/FO", "CSV"}
	if !reflect.DeepEqual(listCmd, expectList) {
		t.Errorf("expected %v, got %v", expectList, listCmd)
	}

	if statusCmd := adapter.ProcessStatusCommand(0); statusCmd != nil {
		t.Errorf("expected nil for 0 PID, got %v", statusCmd)
	}

	validStatus := adapter.ProcessStatusCommand(1234)
	expectStatus := []string{"tasklist", "/FI", "PID eq 1234", "/V", "/FO", "CSV"}
	if !reflect.DeepEqual(validStatus, expectStatus) {
		t.Errorf("expected %v, got %v", expectStatus, validStatus)
	}

	if termCmd := adapter.ProcessTerminateCommand(-10); termCmd != nil {
		t.Errorf("expected nil for negative PID, got %v", termCmd)
	}

	validTerm := adapter.ProcessTerminateCommand(5678)
	expectTerm := []string{"taskkill", "/PID", "5678", "/F"}
	if !reflect.DeepEqual(validTerm, expectTerm) {
		t.Errorf("expected %v, got %v", expectTerm, validTerm)
	}
}

func TestWindowsAdapter_NetworkCommands(t *testing.T) {
	adapter := windowsAdapter{}
	cmds := adapter.ListenerCommands()
	if len(cmds) != 1 {
		t.Errorf("expected 1 command sequence, got %d", len(cmds))
	}
	if !reflect.DeepEqual(cmds[0], []string{"netstat", "-ano"}) {
		t.Errorf("netstat command signature mismatch")
	}

	tasks := adapter.ScheduledTaskListCommand()
	expectTasks := []string{"schtasks", "/Query", "/FO", "LIST", "/V"}
	if !reflect.DeepEqual(tasks, expectTasks) {
		t.Errorf("expected %v, got %v", expectTasks, tasks)
	}
}
