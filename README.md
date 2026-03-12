<div align="center">
  <h1>ClawRemove</h1>
  <p><strong>Agent Environment Inspector</strong></p>
  <p><em>Inspect, audit and clean environments where AI agents run.</em></p>
  <p>
    <a href="https://github.com/tianrking/ClawRemove/actions/workflows/ci.yml"><img src="https://github.com/tianrking/ClawRemove/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
    <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-1f6feb" alt="MIT License"></a>
    <img src="https://img.shields.io/badge/go-1.25%2B-00ADD8?logo=go" alt="Go 1.25+">
    <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-111827" alt="Platform support">
    <a href="https://github.com/tianrking/ClawRemove/releases"><img src="https://img.shields.io/github/v/release/tianrking/ClawRemove" alt="Latest release"></a>
  </p>
  <p>English | <a href="./README.zh-CN.md">中文</a> | <a href="./README.es.md">Español</a></p>
</div>

ClawRemove is an **Agent Environment Inspector** - a tool to inspect, audit, and clean environments where AI agents run.

It focuses on AI agent runtime, tools, and artifacts - not general system cleaning or security scanning.

## What ClawRemove Does

- **Detect** AI runtimes (Ollama, LM Studio, GPT4All, LocalAI)
- **Audit** agent installations (OpenClaw, NanoBot, Cursor, Windsurf)
- **Analyze** AI storage (models, caches, vector databases)
- **Clean** agent environments safely
- **Check** for exposed API keys in AI tool configs

## Quick Start

```bash
# Full environment audit
clawremove audit

# Check for exposed API keys
clawremove security

# Analyze AI storage usage
clawremove hygiene

# Clean up an agent
clawremove cleanup --product openclaw
```

## Documentation

- English: `README.md`
- 中文: [README.zh-CN.md](./README.zh-CN.md)
- Español: [README.es.md](./README.es.md)
- Development plan for agents: [docs/PLAN.md](./docs/PLAN.md)
- Architecture: [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)

## Supported AI Agents

| Agent | Description | State Directory |
|-------|-------------|-----------------|
| OpenClaw | AI assistant platform | ~/.openclaw |
| NanoBot | Python-based AI agent | ~/.nanobot |
| PicoClaw | Go-based AI agent | ~/.picoclaw |
| OpenFang | AI agent framework | ~/.openfang |
| ZeroClaw | Lightweight AI agent | ~/.zeroclaw |
| NanoClaw | Mini AI agent | ~/.nanoclaw |

## Scope

ClawRemove does not try to optimize a machine, patch the registry blindly, install services, or run forever in the background.

It focuses on one job:

1. identify product-owned artifacts
2. build a deletion plan
3. execute only the approved actions
4. verify what remains

It is a removal-first tool, not a system maintenance suite.

## Project Status

ClawRemove is in active development.

The current release target is a production-grade AI Agent removal CLI built on top of a provider-based engine that supports OpenClaw, NanoBot, PicoClaw, and future AI agents without rewriting the core.

Each product provider provides facts as well as runtime tool and skill contracts for controlled analysis and future expansion. For a detailed guide on creating and injecting new providers, see the [Provider Authoring Guide](docs/PROVIDER_AUTHORING.md).

## Core Principles

- Source-driven discovery: product facts come from actual analyzed install and storage behavior.
- Plan before action: destructive changes should be inspectable before execution.
- High-risk actions are opt-in: killing processes and removing containers or images are never implicit.
- Minimal footprint: no telemetry, no persistent service, no hidden state database.
- Provider architecture: each supported AI agent is a dedicated rule pack.
- Provider capabilities: each supported product can expose its own skills and read-only tool set.
- Controlled intelligence: any future LLM assistance is advisory only unless the deterministic engine can justify the action.

## What ClawRemove Detects

Depending on platform and provider rules, ClawRemove can discover:

- state directories
- workspaces (provider-declared workspace subdirectory names)
- temp and log paths
- app bundles and app data
- launchd, systemd, and scheduled-task registrations
- npm, pnpm, bun, pip, and Homebrew installations
- shell completion and profile traces (content-scanned against provider markers)
- matching processes
- listening ports (provider-declared port list)
- crontab references
- Docker and Podman containers and images
- Windows registry keys and values
- Environment variables
- Hosts file entries

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
  Product provider id. Options: `openclaw`, `nanobot`, `picoclaw`.
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
internal/products/nanobot  NanoBot provider
internal/products/picoclaw PicoClaw provider
internal/skills            provider skill catalog
internal/tools             provider tool catalog
internal/model             reports, evidence, capabilities, and verification models
internal/system            system command runner
docs                       roadmap and development plan
scripts                    build helpers
dist                       local build artifacts
```

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

## Roadmap

The long-term roadmap lives in [docs/PLAN.md](./docs/PLAN.md).

That document is written in English on purpose so human contributors and autonomous agents can use the same source of truth for continued development.

## Contributing

We welcome contributions! Please see [docs/PROVIDER_AUTHORING.md](./docs/PROVIDER_AUTHORING.md) for how to add new AI agent providers.

## License

MIT License - see [LICENSE](./LICENSE) for details.
