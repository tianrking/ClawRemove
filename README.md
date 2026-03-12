# ClawRemove

ClawRemove is a professional cross-platform claw uninstaller engine.

Its purpose is narrow and deliberate: discover, plan, execute, and verify clean removal of supported claw products without behaving like a generic “cleaner” that sprays changes across the system.

ClawRemove is designed to be:

- evidence-driven
- conservative by default
- safe to audit before execution
- portable across macOS, Linux, and Windows
- extensible through product providers
- suitable for both CLI-first workflows and future desktop control software

## Documentation

- English: `README.md`
- 中文: [README.zh-CN.md](./README.zh-CN.md)
- Español: [README.es.md](./README.es.md)

## Scope

ClawRemove does not try to optimize a machine, patch the registry blindly, install services, or run forever in the background.

It focuses on one job:

1. identify product-owned artifacts
2. build a deletion plan
3. execute only the approved actions
4. verify what remains

## Current Provider

Currently included:

- `openclaw`

The engine is already structured for future providers such as other claw-family products, but the current implementation is intentionally focused on OpenClaw first.

## Core Principles

- Source-driven discovery: product facts come from actual analyzed install/storage behavior.
- Plan before action: destructive changes should be inspectable before execution.
- High-risk actions are opt-in: killing processes and removing containers/images are never implicit.
- Minimal footprint: no telemetry, no persistent service, no hidden state database.
- Provider architecture: each supported claw product is a dedicated rule pack.

## Commands

```bash
claw-remove products
claw-remove audit --product openclaw --json
claw-remove plan --product openclaw --json
claw-remove apply --product openclaw --dry-run
claw-remove apply --product openclaw
claw-remove verify --product openclaw --json
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
  Runs a post-removal audit-style verification pass.

## Flags

Shared flags:

- `--product`
  Product provider id. Current default: `openclaw`.
- `--json`
  Emit structured machine-readable output.
- `--dry-run`
  Report intended changes without applying them.
- `--keep-cli`
  Preserve package uninstall actions and CLI wrapper deletion.
- `--keep-app`
  Preserve desktop app bundles and app-specific data.
- `--keep-workspace`
  Preserve workspace directories.
- `--keep-shell`
  Preserve shell completion/profile integration cleanup.
- `--kill-processes`
  Opt in to terminating matching live processes.
- `--remove-docker`
  Opt in to removing matching Docker/Podman containers and images.

## What ClawRemove Detects

Depending on platform and provider rules, ClawRemove can discover:

- state directories
- workspaces
- temp and log paths
- app bundles and app data
- launchd, systemd, and scheduled-task registrations
- npm, pnpm, bun, and Homebrew installations
- shell completion/profile traces
- matching processes
- listening ports
- crontab references
- Docker and Podman containers/images

## Safety Model

ClawRemove intentionally separates actions by risk:

- Low risk
  Known provider-owned paths such as state, temp, wrapper, and app artifacts.
- Medium risk
  Service unload/disable actions and package manager uninstall actions.
- High risk
  Process termination and container/image removal.

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
scripts                    build helpers
dist                       local build artifacts
```

The architecture is built to support:

- more claw-family providers later
- a future desktop controller or “upper computer” UI
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

## Repository Notes

This repository is intentionally organized so that the tracked product source lives in `ClawRemove/`.

Outer workspace material can be used for:

- provider research
- agent-driven analysis
- external references
- scratch experiments

Those outer artifacts are intentionally excluded from version control.
