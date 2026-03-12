package platform

import "github.com/tianrking/ClawRemove/internal/model"

type windowsAdapter struct{}

func (windowsAdapter) ServiceStatusCommand(service model.ServiceRef, _ string) []string {
	return []string{"schtasks", "/Query", "/TN", service.Name, "/V", "/FO", "LIST"}
}

func (windowsAdapter) ServiceDisableCommand(service model.ServiceRef) []string {
	return []string{"schtasks", "/Delete", "/F", "/TN", service.Name}
}

func (windowsAdapter) ProcessListCommand() []string {
	return []string{"tasklist", "/V", "/FO", "CSV"}
}

func (windowsAdapter) ProcessStatusCommand(pid int) []string {
	if pid <= 0 {
		return nil
	}
	return []string{"tasklist", "/FI", "PID eq " + itoa(pid), "/V", "/FO", "CSV"}
}

func (windowsAdapter) ProcessTerminateCommand(pid int) []string {
	if pid <= 0 {
		return nil
	}
	return []string{"taskkill", "/PID", itoa(pid), "/F"}
}

func (windowsAdapter) ListenerCommands() [][]string {
	return [][]string{{"netstat", "-ano"}}
}

func (windowsAdapter) ScheduledTaskListCommand() []string {
	return []string{"schtasks", "/Query", "/FO", "LIST", "/V"}
}

// RegistryQueryCommand returns the command to query a registry key.
func (windowsAdapter) RegistryQueryCommand(rootKey, path string) []string {
	return []string{"reg", "query", rootKey + "\\" + path}
}

// RegistryQueryRecursiveCommand returns the command to query a registry key recursively.
func (windowsAdapter) RegistryQueryRecursiveCommand(rootKey, path string) []string {
	return []string{"reg", "query", rootKey + "\\" + path, "/s"}
}

// RegistryDeleteKeyCommand returns the command to delete a registry key.
func (windowsAdapter) RegistryDeleteKeyCommand(rootKey, path string) []string {
	return []string{"reg", "delete", rootKey + "\\" + path, "/f"}
}

// RegistryDeleteValueCommand returns the command to delete a registry value.
func (windowsAdapter) RegistryDeleteValueCommand(rootKey, path, value string) []string {
	return []string{"reg", "delete", rootKey + "\\" + path, "/v", value, "/f"}
}

// EnvGetCommand returns the command to get an environment variable.
func (windowsAdapter) EnvGetCommand(name string, systemScope bool) []string {
	if systemScope {
		return []string{"reg", "query", "HKLM\\SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Environment", "/v", name}
	}
	return []string{"reg", "query", "HKCU\\Environment", "/v", name}
}

// HostsFilePath returns the path to the Windows hosts file.
func (windowsAdapter) HostsFilePath() string {
	return "C:\\Windows\\System32\\drivers\\etc\\hosts"
}
