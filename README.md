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

## Commands

```bash
claw-remove products
claw-remove audit --product openclaw --json
claw-remove plan --product openclaw --json
claw-remove apply --product openclaw --dry-run
claw-remove apply --product openclaw
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
  Executes the planned actions.
- `verify`
  Runs a post-removal verification pass.
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
- `--dry-run`
  Report intended changes without applying them.
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
internal/plan              safe action planning
internal/executor          command and file execution
internal/output            human and JSON reporting
internal/products          provider registry
internal/products/openclaw OpenClaw provider
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

## Roadmap

The long-term roadmap lives in [docs/PLAN.md](./docs/PLAN.md).

That document is written in English on purpose so human contributors and autonomous agents can use the same source of truth for continued development.
