package platform

import (
	"reflect"
	"testing"

	"github.com/tianrking/ClawRemove/internal/model"
)

func TestGenericAdapter_ServiceCommands(t *testing.T) {
	adapter := genericAdapter{}
	
	if cmd := adapter.ServiceStatusCommand(model.ServiceRef{Name: "test"}, ""); cmd != nil {
		t.Errorf("expected nil service status for generic, got %v", cmd)
	}

	if cmd := adapter.ServiceDisableCommand(model.ServiceRef{Name: "test"}); cmd != nil {
		t.Errorf("expected nil service disable for generic, got %v", cmd)
	}
}

func TestGenericAdapter_ProcessCommands(t *testing.T) {
	adapter := genericAdapter{}
	
	if listCmd := adapter.ProcessListCommand(); len(listCmd) == 0 {
		t.Error("ProcessListCommand returned empty")
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

func TestGenericAdapter_NetworkCommands(t *testing.T) {
	adapter := genericAdapter{}
	if cmds := adapter.ListenerCommands(); cmds != nil {
		t.Errorf("expected nil listener commands for generic, got %v", cmds)
	}

	if tasks := adapter.ScheduledTaskListCommand(); tasks != nil {
		t.Errorf("expected nil scheduled task list for generic, got %v", tasks)
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{999, "999"},
		{1234567890, "1234567890"},
	}

	for _, tt := range tests {
		got := itoa(tt.input)
		if got != tt.want {
			t.Errorf("itoa(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
