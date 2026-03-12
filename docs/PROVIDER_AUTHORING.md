# Provider Authoring Guide

ClawRemove is designed to be an extensible removal engine. The core engine orchestrated evidence gathering, planning, and execution, but it relies on **Product Providers** to supply the actual rules, heuristics, and tools.

This guide explains how to add a new provider to ClawRemove safely and effectively.

## The Provider Interface

Every provider must implement the `products.Provider` interface:

```go
type Provider interface {
	ID() string
	DisplayName() string
	Facts() model.ProviderFacts
	Capabilities() model.ProviderCapabilities
	Tools() []tools.Tool
	Skills() []skills.Skill
}
```

### `ID()` and `DisplayName()`
- `ID()` must be a strict, lowercase, alphanumeric string without spaces (e.g., `"openclaw"`). This is used in the CLI via `--product openclaw`.
- `DisplayName()` is the human-readable string used in outputs and reports (e.g., `"OpenClaw"`).

### `Facts()`
Facts are deterministic, static paths and names that the engine will look for unconditionally.

```go
type ProviderFacts struct {
	StateDirNames []string // Scanned in home directories, temp directories, /var/log, etc.
	Markers       []string // Specific filenames indicating activity (e.g. .openclaw.lock)
	TempPrefixes  []string // Prefixes of temp files dropped by this product
	AppPaths      []string // Absolute paths to the binary or application directory
}
```

**Golden Rule:** Do not add broad wildcards or generic names (like `.log` or `data`) to `Facts`. Only add highly specific, product-owned strings.

### `Tools()` and `Skills()`

Tools and Skills form the **Runtime Contract** between the provider and the engine (including the LLM Advisor).

- **Tools** are atomic, read-only probes (e.g., `state_probe`, `ps_grep`) that can be executed to gather specific contextual evidence during analysis.
- **Skills** are higher-level analysis functions that process data and return actionable insights.

You must return concrete implementations of `tools.Tool` and `skills.Skill`. The `Capabilities()` method can then dynamically map these contracts back to metadata so the LLM can understand what the provider can do:

```go
func (p *myProvider) Capabilities() model.ProviderCapabilities {
	var skillMeta []model.ProviderSkill
	for _, skill := range p.Skills() {
		skillMeta = append(skillMeta, skill.Info())
	}
	var toolMeta []model.ProviderTool
	for _, t := range p.Tools() {
		toolMeta = append(toolMeta, t.Info())
	}
	return model.ProviderCapabilities{
		Skills: skillMeta,
		Tools:  toolMeta,
	}
}
```

## Creating a Provider Contract

### 1. Define a Target Strategy

Before writing code, analyze how the target software persists:
- Does it use standard OS services (systemd, launchd, schtasks)?
- Where does it drop state (`~/.name`, `%APPDATA%\Name`)?
- Does it leave shell profile hooks (`.bashrc`, `.zshrc`)?

### 2. Implement the Structure

Create a folder under `internal/products/` matching your product ID (e.g., `internal/products/myproduct`).

Create `provider.go` implementing the interface. Use the `openclaw` provider as a structural reference.

### 3. Register the Provider

Open `internal/products/registry.go` and add your provider to the `init()` block or the `Registry()` map:

```go
func init() {
	registry["myproduct"] = myproduct.New()
}
```

## Conformance Requirements

All providers are checked against the `conformance_test.go` suite in the `internal/products` package. Before submitting a PR, run:

```bash
go test -v ./internal/products
```

The conformance tests enforce:
- Valid `ID` formatting.
- Capability parity (metadata lengths must match the slices of runtime contracts).
- Graceful degradation for runtime tools (e.g., `Execute(nil)` shouldn't induce kernel panics).
- Non-empty descriptions for all Tools and Skills.

## Safety Constraints Review

When designing your provider, strictly adhere to the following ClawRemove principles:

1. **No Destructive Tools**: Your `tools.Tool` implementations must be **read-only** (e.g., reading logs, checking statuses). Do not invoke `rm`, `kill`, or process injection from inside a tool.
2. **Evidence-Based Planning**: If you need to remove something, write a heuristic in the `verify` layer or rely on standard facts. Do not mutate the system silently!
3. **Graceful Fallbacks**: If your tool is requested by an LLM but missing a required argument, gracefully return a deterministic error string (e.g., `"missing required 'target' argument"`), rather than panicking.
