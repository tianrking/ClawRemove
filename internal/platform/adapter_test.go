package platform

import (
	"testing"

	"github.com/tianrking/ClawRemove/internal/model"
)

func TestNewAdapter(t *testing.T) {
	cases := []Host{
		{OS: "darwin"},
		{OS: "linux"},
		{OS: "windows"},
		{OS: "unknown"},
	}
	for _, c := range cases {
		adapter := NewAdapter(c)
		if adapter == nil {
			t.Fatalf("expected adapter for os=%s", c.OS)
		}
	}
}

func TestLinuxServiceCommandUsesScope(t *testing.T) {
	adapter := linuxAdapter{}
	service := model.ServiceRef{Name: "openclaw-gateway", Scope: "user"}
	cmd := adapter.ServiceStatusCommand(service, "")
	if len(cmd) < 3 || cmd[0] != "systemctl" || cmd[1] != "--user" {
		t.Fatalf("unexpected linux user service command: %#v", cmd)
	}
}
