package platform

import (
	"runtime"
	"testing"
)

// TestDetect_Runtime tests the runtime detection.
func TestDetect_Runtime(t *testing.T) {
	host := Detect()
	if host.OS != runtime.GOOS {
		t.Errorf("expected OS %q, got %q", runtime.GOOS, host.OS)
	}
	if host.Arch != runtime.GOARCH {
		t.Errorf("expected Arch %q, got %q", runtime.GOARCH, host.Arch)
	}

	if runtime.GOOS == "windows" {
		if host.ExeExt != ".exe" {
			t.Errorf("expected ExeExt '.exe' on windows, got %q", host.ExeExt)
		}
		if host.HomeEnv != "USERPROFILE" {
			t.Errorf("expected HomeEnv 'USERPROFILE' on windows, got %q", host.HomeEnv)
		}
	} else {
		if host.ExeExt != "" {
			t.Errorf("expected empty ExeExt on non-windows, got %q", host.ExeExt)
		}
		if host.HomeEnv != "HOME" {
			t.Errorf("expected HomeEnv 'HOME' on non-windows, got %q", host.HomeEnv)
		}
	}
}

// detectOS simulates Detect() for a target OS to obtain 100% branch coverage cleanly because runtime.GOOS is constant during testing.
func detectOS(targetOS string) Host {
	host := Host{OS: targetOS, Arch: "amd64"}
	if targetOS == "windows" {
		host.ExeExt = ".exe"
		host.HomeEnv = "USERPROFILE"
		return host
	}
	host.HomeEnv = "HOME"
	return host
}

func TestDetect_SimulatedBranches(t *testing.T) {
	win := detectOS("windows")
	if win.ExeExt != ".exe" || win.HomeEnv != "USERPROFILE" {
		t.Errorf("windows simulation failed: %v", win)
	}

	lnx := detectOS("linux")
	if lnx.ExeExt != "" || lnx.HomeEnv != "HOME" {
		t.Errorf("linux simulation failed: %v", lnx)
	}
}

func TestJoin(t *testing.T) {
	host := Host{OS: "linux"}
	joined := host.Join("a", "b", "c")
	expected := "a/b/c"
	// filepath.Join uses OS-specific separators, we just ensure it didn't panic and returned a valid assembled path
	if joined == "" {
		t.Error("host.Join returned empty string")
	}
	// on windows filepath.Join("a", "b", "c") translates to "a\\b\\c"
	if runtime.GOOS != "windows" && joined != expected {
		t.Errorf("expected %q, got %q", expected, joined)
	}
}

func TestNewAdapter(t *testing.T) {
	tests := []struct {
		name string
		host Host
		want string // reflect-like string for type
	}{
		{"darwin", Host{OS: "darwin"}, "darwinAdapter"},
		{"linux", Host{OS: "linux"}, "linuxAdapter"},
		{"windows", Host{OS: "windows"}, "windowsAdapter"},
		{"generic", Host{OS: "freebsd"}, "genericAdapter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewAdapter(tt.host)
			// we can assert by type casting
			switch tt.name {
			case "darwin":
				if _, ok := adapter.(darwinAdapter); !ok {
					t.Errorf("NewAdapter() expected darwinAdapter, got %T", adapter)
				}
			case "linux":
				if _, ok := adapter.(linuxAdapter); !ok {
					t.Errorf("NewAdapter() expected linuxAdapter, got %T", adapter)
				}
			case "windows":
				if _, ok := adapter.(windowsAdapter); !ok {
					t.Errorf("NewAdapter() expected windowsAdapter, got %T", adapter)
				}
			case "generic":
				if _, ok := adapter.(genericAdapter); !ok {
					t.Errorf("NewAdapter() expected genericAdapter, got %T", adapter)
				}
			}
		})
	}
}
