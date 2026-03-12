<div align="center">
  <h1>ClawRemove</h1>
  <p><strong>A surgical, cross-platform claw removal engine.</strong></p>
  <p>
    <a href="https://github.com/tianrking/ClawRemove/actions/workflows/ci.yml"><img src="https://github.com/tianrking/ClawRemove/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
    <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-1f6feb" alt="MIT License"></a>
    <img src="https://img.shields.io/badge/go-1.25%2B-00ADD8?logo=go" alt="Go 1.25+">
    <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-111827" alt="Platform support">
    <a href="https://github.com/tianrking/ClawRemove/releases"><img src="https://img.shields.io/github/v/release/tianrking/ClawRemove" alt="Latest release"></a>
  </p>
  <p>English | <a href="./README.zh-CN.md">中文</a> | <a href="./README.es.md">Español</a></p>
</div>

ClawRemove is a professional cross-platform claw removal engine written in Go.

Its purpose is narrow and deliberate: discover, plan, execute, and verify clean removal of OpenClaw and other claw-family agents without behaving like a generic cleaner that sprays changes across the system.

ClawRemove is designed to be:

- evidence-driven
- conservative by default
- safe to audit before execution
- portable across macOS, Linux, and Windows
- extensible through product providers
- suitable for both CLI-first workflows and future desktop control software

ClawRemove should be understood as a controlled uninstall claw:

- it understands how claw agents install, persist, and leave residue
- it uses that understanding to remove them cleanly
- it does not become a noisy resident agent itself

## Documentation

- English: `README.md`
- 中文: [README.zh-CN.md](./README.zh-CN.md)
- Español: [README.es.md](./README.es.md)
- Development plan for agents: [docs/PLAN.md](./docs/PLAN.md)
- Architecture: [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)

## Scope

ClawRemove does not try to optimize a machine, patch the registry blindly, install services, or run forever in the background.

It focuses on one job:

1. identify product-owned artifacts
2. build a deletion plan
3. execute only the approved actions
4. verify what remains

It is a removal-first tool, not a system maintenance suite.

## Project Status

ClawRemove is in active build-out.

The current release target is a production-grade OpenClaw removal CLI built on top of a provider-based engine that can later support additional claw-family products without rewriting the core.

Each product provider is now expected to describe not only facts, but also provider-specific skills and tools for controlled analysis and future expansion.

Release scripts now inject version metadata into binaries and produce per-platform archives with SHA256 checksums.

## Current Provider

Currently included:

- `openclaw`

The engine is already structured for future providers such as other claw-family products, but the current implementation is intentionally focused on OpenClaw first.

## Core Principles

- Source-driven discovery: product facts come from actual analyzed install and storage behavior.
- Plan before action: destructive changes should be inspectable before execution.
- High-risk actions are opt-in: killing processes and removing containers or images are never implicit.
- Minimal footprint: no telemetry, no persistent service, no hidden state database.
- Provider architecture: each supported claw product is a dedicated rule pack.
- Provider capabilities: each supported product can expose its own skills and read-only tool set.
- Controlled intelligence: any future LLM assistance is advisory only unless the deterministic engine can justify the action.

## Why ClawRemove

- Built for removal, not generic machine cleanup.
- Default behavior is conservative and reviewable.
- Evidence matters more than heuristics.
- JSON output is suitable for automation and future desktop tooling.
- The repository is structured for continued agent-driven iteration.
- The long-term design allows an AI-assisted analyst without letting the model directly mutate the system.

## Controlled AI Direction

ClawRemove may later integrate an LLM-assisted analysis layer, but only under strict constraints:

- the LLM can explain findings and ask for more evidence
- the LLM can help classify uncertainty and improve operator guidance
- the LLM cannot directly issue destructive shell commands
- the deterministic engine remains the final authority for execution

This keeps ClawRemove useful like an agent while remaining auditable like a proper system tool.

The first implementation of this architecture is now present:

- multi-provider LLM support for OpenAI, Anthropic, and other OpenAI-compatible APIs
- a controlled ReAct loop
- an explicit `internal/evidence` layer
- explicit evidence provenance (`rule`, `source`, `confidence`) consumed by planning and verification
- a split LLM stack with `internal/llm/prompts`, `internal/llm/providers`, and `internal/llm/mediation`
- platform adapters for darwin/linux/windows used by controlled probes, discovery, and planning
- a read-only tool protocol over in-memory discovery and plan data
- provider-specific skills and tools metadata
- a hard boundary that prevents the model from issuing destructive commands directly

The current architecture assessment and target structure are documented in [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md).

## Commands

```bash
claw-remove products
claw-remove audit --product openclaw --json
claw-remove plan --product openclaw --json
claw-remove apply --product openclaw --dry-run
claw-remove apply --product openclaw
claw-remove apply --product openclaw --yes
claw-remove verify --product openclaw --json
claw-remove explain --product openclaw --json
```

### Command Summary

- `products`
  Lists compiled-in product providers.
- `audit`
  Read-only discovery and residual analysis.
- `plan`
  Produces a deletion plan without applying it.
- `apply`
  Executes the planned actions after an interactive safety confirmation.
- `verify`
  Runs a post-removal verification pass with residual classification.
- `explain`
  Produces controlled advisory analysis on top of deterministic discovery.

## Flags

Shared flags:

- `--product`
  Product provider id. Current default: `openclaw`.
- `--json`
  Emit structured machine-readable output.
- `--ai`
  Include controlled advisory analysis in the report.
- `--version`
  Print version information and exit.
- `--dry-run`
  Report intended changes without applying them.
- `--yes`
  Skip interactive confirmation after you have already reviewed the plan.
- `--keep-cli`
  Preserve package uninstall actions and CLI wrapper deletion.
- `--keep-app`
  Preserve desktop app bundles and app-specific data.
- `--keep-workspace`
  Preserve workspace directories.
- `--keep-shell`
  Preserve shell completion and profile integration cleanup.
- `--kill-processes`
  Opt in to terminating matching live processes.
- `--remove-docker`
  Opt in to removing matching Docker or Podman containers and images.

## LLM Configuration

ClawRemove can attach a controlled advisor to `audit`, `verify`, and `explain`.

The advisor is optional. If no LLM configuration is present, ClawRemove falls back to deterministic advisory output.

Supported providers:

- `openai`
- `anthropic`
- `openrouter`
- `zhipu`
- `openai-compatible`

Environment variables:

- `CLAWREMOVE_LLM_PROVIDERS`
  Comma-separated provider chain. Example: `openai,openrouter,zhipu`.
- `CLAWREMOVE_LLM_PROVIDER`
  Single-provider shorthand (legacy fallback).
- `CLAWREMOVE_LLM_API_KEY`
  Generic API key override for the configured provider.
- `OPENAI_API_KEY`
  Fallback key when `CLAWREMOVE_LLM_PROVIDER=openai`.
- `ANTHROPIC_API_KEY`
  Fallback key when `CLAWREMOVE_LLM_PROVIDER=anthropic`.
- `OPENROUTER_API_KEY`
  Fallback key when provider is `openrouter`.
- `ZHIPU_API_KEY`
  Fallback key when provider is `zhipu`.
- `BIGMODEL_API_KEY`
  Secondary fallback key for `zhipu`.
- `CLAWREMOVE_LLM_BASE_URL`
  Provider base URL override.
- `CLAWREMOVE_LLM_MODEL`
  Model name override.
- `CLAWREMOVE_LLM_MODELS`
  Comma-separated model fallback chain shared by configured providers.
- `CLAWREMOVE_LLM_MAX_TOKENS`
  Output token budget for advisory responses.
- `CLAWREMOVE_LLM_MAX_STEPS`
  Maximum controlled ReAct steps.
- `CLAWREMOVE_LLM_TIMEOUT_SECONDS`
  Request timeout in seconds.
- `CLAWREMOVE_LLM_TRACE`
  Enable routing trace output in `advice.trace` (attempt chain + selected provider/model).

Defaults:

- `openai`
  Base URL: `https://api.openai.com/v1`
  Model: `gpt-4.1-mini`
- `anthropic`
  Base URL: `https://api.anthropic.com/v1`
  Model: `claude-3-5-sonnet-latest`
- `openai-compatible`
  Base URL: `https://api.openai.com/v1`
  Model: `gpt-4.1-mini`

Provider-specific overrides are also supported:

- `CLAWREMOVE_LLM_OPENAI_BASE_URL`, `CLAWREMOVE_LLM_OPENAI_MODELS`, `CLAWREMOVE_LLM_OPENAI_API_KEY`
- `CLAWREMOVE_LLM_ANTHROPIC_BASE_URL`, `CLAWREMOVE_LLM_ANTHROPIC_MODELS`, `CLAWREMOVE_LLM_ANTHROPIC_API_KEY`
- `CLAWREMOVE_LLM_OPENROUTER_BASE_URL`, `CLAWREMOVE_LLM_OPENROUTER_MODELS`, `CLAWREMOVE_LLM_OPENROUTER_API_KEY`
- `CLAWREMOVE_LLM_ZHIPU_BASE_URL`, `CLAWREMOVE_LLM_ZHIPU_MODELS`, `CLAWREMOVE_LLM_ZHIPU_API_KEY`

Example with OpenAI:

```bash
export CLAWREMOVE_LLM_PROVIDER="openai"
export OPENAI_API_KEY="..."
export CLAWREMOVE_LLM_MODEL="gpt-4.1-mini"
claw-remove explain --product openclaw --ai --json
```

Example with Anthropic:

```bash
export CLAWREMOVE_LLM_PROVIDER="anthropic"
export ANTHROPIC_API_KEY="..."
export CLAWREMOVE_LLM_MODEL="claude-3-5-sonnet-latest"
claw-remove explain --product openclaw --ai --json
```

Example with another OpenAI-compatible provider:

```bash
export CLAWREMOVE_LLM_PROVIDER="openai-compatible"
export CLAWREMOVE_LLM_BASE_URL="https://your-provider.example/v1"
export CLAWREMOVE_LLM_API_KEY="..."
export CLAWREMOVE_LLM_MODEL="your-model-name"
claw-remove explain --product openclaw --ai --json
```

Example with a multi-provider fallback chain:

```bash
export CLAWREMOVE_LLM_PROVIDERS="openai,openrouter,zhipu,anthropic"
export OPENAI_API_KEY="..."
export OPENROUTER_API_KEY="..."
export ZHIPU_API_KEY="..."
export ANTHROPIC_API_KEY="..."
export CLAWREMOVE_LLM_MODELS="gpt-4.1-mini,glm-4.5-air,claude-3-5-sonnet-latest"
claw-remove explain --product openclaw --ai --json
```

## Controlled Tool Protocol

The LLM does not receive shell access.

Instead, it can request only read-only tools over the existing report:

- `summary`
- `verification`
- `state_dirs`
- `workspace_dirs`
- `services`
- `packages`
- `processes`
- `containers`
- `plan_actions`
- `path_probe`
- `service_probe`
- `package_probe`
- `process_probe`
- `shell_profile_probe`

Those tools either inspect in-memory data already collected by the deterministic engine or perform tightly scoped read-only probes against already discovered targets. They do not mutate files, do not delete data, and do not broaden the scan surface beyond known findings.

## Provider Skills And Tools

ClawRemove now treats product providers as capability bundles instead of plain fact tables.

Each provider can define:

- `facts`
  Install paths, markers, package refs, shell traces, and other deterministic fingerprints.
- `skills`
  High-level product-specific analysis abilities that the advisor and future controllers can reason about.
- `tools`
  Read-only probes that are safe for the advisor to call when investigating already discovered targets.

This structure is intended to scale as ClawRemove adds more products, more models, and richer operator guidance without turning the engine into an unbounded autonomous agent.

## Architecture Status

Current state:

- clear enough to build on
- not bloated yet
- partially decoupled
- still needs a dedicated evidence layer
- still needs stronger platform adapters
- still needs a fuller split inside the LLM subsystem
- still needs broader platform adapter coverage beyond probe command routing

ClawRemove is intentionally being shaped toward a stricter architecture now, before more providers and models make the code harder to untangle.

`cmd/claw-remove/main.go` intentionally lives under `cmd/claw-remove/` even though it is a single file. This is idiomatic Go command layout and keeps space for additional binaries later without restructuring imports or release scripts.

## Verification Model

`verify` now classifies leftovers instead of behaving like a second plain audit pass.

Residuals are grouped as:

- `exact`
  Confirmed product-owned residue, such as state directories and app artifacts.
- `strong`
  Highly likely residue, such as installed packages, service registrations, and live matching processes.
- `heuristic`
  Evidence that still needs human review, such as listeners, shell profiles, and some cron or image matches.

This classification is also exposed to the LLM advisor so the model can reason over stronger evidence instead of guessing from raw discovery alone.

## Safe Removal Workflow

ClawRemove is designed to remove as much confirmed residue as possible while still making unsafe operations explicit.

Recommended workflow:

1. `audit`
   See what ClawRemove found.
2. `verify`
   Review confirmed residue versus investigate-only residue.
3. `explain --ai`
   Ask the controlled advisor to summarize what matters.
4. `apply`
   Review the dry-run plan shown by ClawRemove and confirm interactively.
5. `apply --yes`
   Use only for automation or after prior review.

By default, `apply` is not silent. It prints a preview and asks you to type a confirmation phrase before removal starts.

This keeps the tool comprehensive without turning it into an unsafe fully automatic remover.

If you need JSON output for automation, use `plan` or `verify` for review first, then call `apply --yes`.

## What ClawRemove Detects

Depending on platform and provider rules, ClawRemove can discover:

- state directories
- workspaces
- temp and log paths
- app bundles and app data
- launchd, systemd, and scheduled-task registrations
- npm, pnpm, bun, and Homebrew installations
- shell completion and profile traces
- matching processes
- listening ports
- crontab references
- Docker and Podman containers and images

## Safety Model

ClawRemove intentionally separates actions by risk:

- Low risk
  Known provider-owned paths such as state, temp, wrapper, and app artifacts.
- Medium risk
  Service unload or disable actions and package manager uninstall actions.
- High risk
  Process termination and container or image removal.

High-risk actions require explicit opt-in flags.

Heuristic-only findings should be reported instead of silently deleted.

## Architecture

```text
cmd/claw-remove            CLI entrypoint
internal/app               CLI command wiring
internal/core              engine orchestration
internal/discovery         source-driven discovery layer
internal/evidence          evidence building between discovery and plan
internal/plan              safe action planning
internal/executor          command and file execution
internal/llm               advisor coordination
internal/llm/prompts       prompt definitions
internal/llm/providers     model provider clients
internal/output            human and JSON reporting
internal/platform          host and platform abstractions
internal/products          provider registry
internal/products/openclaw OpenClaw provider
internal/skills            provider skill catalog
internal/tools             provider tool catalog
internal/model             reports, evidence, capabilities, and verification models
internal/system            system command runner
docs                       roadmap and development plan
scripts                    build helpers
dist                       local build artifacts
```

The architecture is built to support:

- more claw-family providers later
- a future desktop controller or upper-computer UI
- stable JSON reporting for automation

## Build

### Local

```bash
go test ./...
go build -o dist/claw-remove ./cmd/claw-remove
```

### Release Matrix

```bash
./scripts/build.sh
```

Windows PowerShell:

```powershell
./scripts/build.ps1
```

Current release targets:

- `dist/claw-remove-darwin-amd64`
- `dist/claw-remove-darwin-arm64`
- `dist/claw-remove-linux-amd64`
- `dist/claw-remove-linux-arm64`
- `dist/claw-remove-windows-amd64.exe`
- `dist/claw-remove-windows-arm64.exe`

## Example Workflow

Audit first:

```bash
claw-remove audit --product openclaw --json
```

Generate a plan:

```bash
claw-remove plan --product openclaw --json
```

Run a dry-run apply:

```bash
claw-remove apply --product openclaw --dry-run
```

Execute for real:

```bash
claw-remove apply --product openclaw
```

Verify residual state:

```bash
claw-remove verify --product openclaw --json
```

Ask for a controlled explanation:

```bash
claw-remove explain --product openclaw --json
```

Ask for an LLM-assisted explanation:

```bash
claw-remove explain --product openclaw --ai --json
```

Run a safe interactive removal:

```bash
claw-remove apply --product openclaw
```

Run a non-interactive removal only after review:

```bash
claw-remove apply --product openclaw --yes
```

## Roadmap

The long-term roadmap lives in [docs/PLAN.md](./docs/PLAN.md).

That document is written in English on purpose so human contributors and autonomous agents can use the same source of truth for continued development.
