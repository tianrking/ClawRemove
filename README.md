<div align="center">
  <h1>ClawRemove</h1>
  <p><strong>Agent Environment Inspector</strong></p>
  <p><em>Inspect, audit and clean environments where AI agents run.</em></p>
  <p>
    <a href="https://github.com/tianrking/ClawRemove/actions/workflows/ci.yml"><img src="https://github.com/tianrking/ClawRemove/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
    <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-1f6feb" alt="MIT License"></a>
    <img src="https://img.shields.io/badge/go-1.25%2B-00ADD8?logo=go" alt="Go 1.25+">
    <img src="https://img.shields.io/badge/size-%3C10MB-success" alt="Size < 10MB">
    <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows%20%7C%20FreeBSD%20%7C%20NetBSD%20%7C%20OpenBSD-111827" alt="Platform support">
    <a href="https://github.com/tianrking/ClawRemove/releases"><img src="https://img.shields.io/github/v/release/tianrking/ClawRemove" alt="Latest release"></a>
  </p>
  <p>English | <a href="./README.zh-CN.md">中文</a> | <a href="./README.es.md">Español</a></p>
</div>

> **🚀 One tool, under 10MB, solves all your OpenClaw-related headaches.**
>
> No dependencies. No installation. Just download and run.

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
claw-remove environment

# Check for exposed API keys
claw-remove security

# Analyze AI storage usage
claw-remove hygiene

# Scan for cleanup candidates (old models, caches, logs)
claw-remove cleanup

# Interactive Terminal UI for environment scanning and cleanup
claw-remove tui

# Scan only a specific category
claw-remove cleanup --category model_version

# Clean up an agent
claw-remove apply --product openclaw
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
| Cursor | AI-powered IDE | ~/.cursor |
| Windsurf | Codeium AI IDE | ~/.windsurf |
| Aider | AI pair programming CLI | ~/.aider |
| Cline | VS Code AI extension | ~/.cline |
| Continue | Open-source AI code assistant | ~/.continue |

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

### Environment Inspection

```bash
# Full environment report
claw-remove environment

# AI inventory only
claw-remove inventory

# Security audit only
claw-remove security

# Storage hygiene only
claw-remove hygiene

# JSON output
claw-remove environment --json
claw-remove security --json
```

### Agent Cleanup

```bash
claw-remove products
claw-remove audit --product openclaw --json
claw-remove plan --product openclaw --json
claw-remove apply --product openclaw --dry-run
claw-remove apply --product openclaw
claw-remove apply --product openclaw --yes
claw-remove verify --product openclaw --json
claw-remove explain --product openclaw --json

### Interactive Cleanup
```bash
claw-remove tui
```
```

### Command Summary

**Environment Commands:**
- `environment`
  Full environment inspection report (runtime, agents, artifacts, security, hygiene).
- `inventory`
  AI runtime and agent inventory.
- `security`
  AI tool security audit (API key exposure).
- `hygiene`
  AI storage usage analysis.
- `cleanup`
  Scan and clean old models, orphaned caches, unused vector DBs, and rotated logs.

**Cleanup Commands:**
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
- `--category`
  Cleanup category filter. Options: `model_version`, `orphaned_cache`, `unused_vectordb`, `log_rotation`, `all`.
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

## AI Analysis (Optional)

ClawRemove can use AI to explain findings in plain language. This is **optional** - all core features work without AI.

### Core Features (No LLM Required)

| Command | Description |
|---------|-------------|
| `claw-remove environment` | Full environment inspection |
| `claw-remove inventory` | AI runtime and agent inventory |
| `claw-remove security` | API key exposure audit |
| `claw-remove hygiene` | Storage usage analysis |
| `claw-remove audit --product X` | Discover residuals |
| `claw-remove plan --product X` | Generate deletion plan |
| `claw-remove apply --product X` | Execute cleanup |
| `claw-remove verify --product X` | Verify cleanup results |

### AI-Enhanced Features (Requires LLM)

| Command | Description |
|---------|-------------|
| `claw-remove explain --ai` | Explain findings in plain language |
| `claw-remove audit --ai` | Audit + AI explanation |
| `claw-remove verify --ai` | Verify + AI explanation |

### What AI Can and Cannot Do

**AI Can:**
- ✅ Explain what was discovered in simple terms
- ✅ Suggest what to review or clean
- ✅ Help classify uncertain items

**AI Cannot:**
- ❌ Execute destructive commands
- ❌ Bypass safety checks
- ❌ Modify your system

### Quick Setup

**Option 1: OpenAI**
```bash
# Set your API key
export OPENAI_API_KEY="sk-your-key-here"

# Use AI analysis
claw-remove explain --product openclaw --ai
```

**Option 2: Anthropic Claude**
```bash
# Set your API key
export ANTHROPIC_API_KEY="sk-ant-your-key-here"
export CLAWREMOVE_LLM_PROVIDER="anthropic"

# Use AI analysis
claw-remove explain --product openclaw --ai
```

**Option 3: OpenAI-Compatible (any provider)**
```bash
# Configure your provider
export CLAWREMOVE_LLM_PROVIDER="openai-compatible"
export CLAWREMOVE_LLM_BASE_URL="https://your-provider.com/v1"
export CLAWREMOVE_LLM_API_KEY="your-key"

# Use AI analysis
claw-remove explain --product openclaw --ai
```

**Option 4: Anthropic-Compatible (e.g., ZhipuAI Coding)**
```bash
# Configure Anthropic-compatible provider
export CLAWREMOVE_LLM_PROVIDER="anthropic-compatible"
export CLAWREMOVE_LLM_BASE_URL="https://open.bigmodel.cn/api/coding/paas/v4"
export CLAWREMOVE_LLM_API_KEY="your-zhipu-api-key"
export CLAWREMOVE_LLM_MODEL="GLM-4.5"

# Use AI analysis
claw-remove explain --product openclaw --ai
```

### All LLM Configuration Options

| Variable | Description | Example |
|----------|-------------|---------|
| `CLAWREMOVE_LLM_PROVIDER` | Provider type | `openai`, `anthropic`, `openai-compatible`, `anthropic-compatible` |
| `CLAWREMOVE_LLM_API_KEY` | API key (generic) | `sk-xxx` |
| `OPENAI_API_KEY` | OpenAI key (fallback) | `sk-xxx` |
| `ANTHROPIC_API_KEY` | Anthropic key (fallback) | `sk-ant-xxx` |
| `CLAWREMOVE_LLM_BASE_URL` | Custom API URL | `https://api.example.com/v1` |
| `CLAWREMOVE_LLM_MODEL` | Model override | `GPT-5.4`, `claude-opus-4-6`, `GLM-4.5` |
| `CLAWREMOVE_LLM_TIMEOUT_SECONDS` | Request timeout | `60` |

### Supported Provider Types

| Provider | API Format | Use Case |
|----------|------------|----------|
| `openai` | OpenAI Native | OpenAI GPT models |
| `anthropic` | Anthropic Native | Claude models |
| `openai-compatible` | OpenAI `/chat/completions` | Custom providers (Ollama, vLLM, etc.) |
| `anthropic-compatible` | Anthropic `/messages` | ZhipuAI Coding, other Anthropic-format APIs |
| `openrouter` | OpenAI `/chat/completions` | OpenRouter aggregator |
| `zhipu` | OpenAI `/chat/completions` | ZhipuAI general API |

### Commands with AI

```bash
# Audit with AI explanation
claw-remove audit --product openclaw --ai

# Get AI analysis of findings
claw-remove explain --product openclaw --ai

# JSON output with AI
claw-remove explain --product openclaw --ai --json

# Full environment analysis with AI
claw-remove environment --ai
```

### ReAct Deep Analysis

When `--ai` is enabled, ClawRemove uses a **ReAct (Reasoning + Acting)** mechanism for comprehensive analysis:

**Analysis Capabilities:**
| Tool | Platform | Description |
|------|----------|-------------|
| `deep_analysis` | All | Comprehensive overview of all artifacts |
| `quick_scan` | All | Fast scan of common sensitive directories |
| `search_agent_traces` | All | Search for agent-specific patterns across filesystem |
| `credential_probe` | All | Detect exposed API keys and secrets |
| `config_probe` | All | Analyze configuration files for agent modifications |
| `file_content_search` | All | Search for specific patterns in files |
| `registry_probe` | Windows | Registry startup entries, uninstall keys |
| `env_probe` | All | PATH modifications, API keys, configs |
| `hosts_probe` | All | Hosts file domain mappings |
| `autostart_probe` | All | launchd, systemd, cron, registry auto-start |
| `shell_profile_probe` | All | Shell aliases, completions, PATH changes |
| `service_probe` | All | Running services status |

**What the AI Analyzes:**
- **Filesystem**: State directories, workspaces, temp files
- **Services**: System services, background processes
- **Environment**: PATH, API keys, custom variables
- **Network**: Hosts entries, port listeners
- **Autostart**: launchd (macOS), systemd (Linux), registry (Windows), cron
- **Credentials**: API keys, secrets, tokens in config files
- **SSH/Git**: Configuration modifications by agents

**Intelligent Features:**
- **Batch Tool Execution**: AI can run multiple probes in parallel
- **Confidence Tracking**: AI reports confidence level (0-100%)
- **Progress Indicators**: starting → gathering → analyzing → finalizing
- **Error Recovery**: Automatically adapts when tools fail

**Example Output:**
```
AI Analysis Summary:
- 49 confirmed residue items found
- Modifications detected in: filesystem, services, autostart
- Registry startup entry: HKCU\Software\OpenClaw
- PATH modified in ~/.zshrc
- Recommend: remove_confirmed_residue (risk=medium)
```

### Real-Time Streaming Output

When using `--ai` mode, ClawRemove shows real-time progress during AI analysis:

```bash
claw-remove explain --product openclaw --ai
```

**Live Progress Example:**
```
🤖 AI Analysis Starting...
   Provider: openclaw
   Command: explain

🔄 ReAct Step 1/10...
   📤 Calling LLM...
   💭 Thought: Analyzing discovered artifacts for agent modifications...
   🔧 Using tool: deep_analysis
   ✅ Tool result received

🔄 ReAct Step 2/10...
   📤 Calling LLM...
   💭 Thought: Investigating shell profile modifications...
   🔧 Using tool: shell_profile_probe
   ✅ Tool result received

✅ AI Analysis Complete!
   📝 Summary: Found 49 confirmed residue items. Recommend removal...
```

**Streaming Behavior:**
- Progress shown in real-time during each ReAct step
- Shows which tools the AI is using
- Displays AI's thought process summaries
- Only streams to terminal (disabled with `--json` or `--quiet`)

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

## Installation

### Pre-built Binaries

Download from [GitHub Releases](https://github.com/tianrking/ClawRemove/releases).

**macOS:**
```bash
# DMG (recommended)
# Download claw-remove-VERSION-macOS.dmg, open and drag to Applications

# Or via tarball
curl -sL https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-darwin-arm64.tar.gz | tar xz
sudo mv claw-remove /usr/local/bin/
```

**Linux:**
```bash
# Debian/Ubuntu (deb package)
wget https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove_VERSION_amd64.deb
sudo dpkg -i claw-remove_VERSION_amd64.deb

# RHEL/Fedora (rpm package)
wget https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-1.x86_64.rpm
sudo rpm -i claw-remove-VERSION-1.x86_64.rpm

# Arch Linux
wget https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-1-x86_64.pkg.tar.zst
sudo pacman -U claw-remove-VERSION-1-x86_64.pkg.tar.zst

# AppImage (universal, no installation required)
wget https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-x86_64.AppImage
chmod +x claw-remove-VERSION-x86_64.AppImage
./claw-remove-VERSION-x86_64.AppImage

# Or via tarball
curl -sL https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-linux-amd64.tar.gz | tar xz
sudo mv claw-remove /usr/local/bin/
```

**Windows:**
```powershell
# Download the ZIP for your architecture
# claw-remove-VERSION-windows-amd64.zip (x64)
# claw-remove-VERSION-windows-arm64.zip (ARM64)
# claw-remove-VERSION-windows-386.zip (32-bit)

# Extract and add to PATH
Expand-Archive claw-remove-VERSION-windows-amd64.zip -DestinationPath C:\Tools\claw-remove
$env:PATH += ";C:\Tools\claw-remove"
```

**BSD:**
```bash
# FreeBSD
curl -sL https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-freebsd-amd64.tar.gz | tar xz
sudo mv claw-remove /usr/local/bin/

# NetBSD / OpenBSD
curl -sL https://github.com/tianrking/ClawRemove/releases/latest/download/claw-remove-VERSION-netbsd-amd64.tar.gz | tar xz
sudo mv claw-remove /usr/pkg/bin/
```

### From Source

```bash
go install github.com/tianrking/ClawRemove/cmd/claw-remove@latest
```

## Build

### Local

```bash
go test ./...
go build -o dist/claw-remove ./cmd/claw-remove
```

### Release

```bash
# Build all platforms
./scripts/build.sh

# Or use GoReleaser (optional)
go install github.com/goreleaser/goreleaser/v2@latest
goreleaser build --snapshot --clean
```

### Release Artifacts

Each release includes:

| Format | Platform | Description |
|--------|----------|-------------|
| `.dmg` | macOS | Disk image installer |
| `.deb` | Linux (Debian/Ubuntu) | APT package |
| `.rpm` | Linux (RHEL/Fedora) | RPM package |
| `.pkg.tar.zst` | Linux (Arch) | Arch Linux package |
| `.AppImage` | Linux (universal) | Portable executable |
| `.zip` | Windows | Archive with binary |
| `.tar.gz` | All platforms | Compressed archive |

Supported platforms (22 total):

**macOS:**
- `darwin-amd64` (Intel)
- `darwin-arm64` (Apple Silicon)

**Linux:**
- `linux-amd64` (x86_64)
- `linux-arm64` (ARM64)
- `linux-386` (32-bit)
- `linux-arm` (ARM v7, Raspberry Pi)
- `linux-riscv64` (RISC-V)
- `linux-ppc64le` (IBM Power)
- `linux-s390x` (IBM Z)
- `linux-mips64` (MIPS64 big-endian)
- `linux-mips64le` (MIPS64 little-endian)

**Windows:**
- `windows-amd64.exe` (x86_64)
- `windows-arm64.exe` (ARM64)
- `windows-386.exe` (32-bit)

**BSD:**
- `freebsd-amd64`, `freebsd-arm64`
- `netbsd-amd64`
- `openbsd-amd64`

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

## Uninstalling ClawRemove

To completely remove ClawRemove from your system:

### macOS (DMG/tarball)

```bash
# Remove binary
sudo rm -f /usr/local/bin/claw-remove
# Or if installed via DMG
sudo rm -rf /Applications/claw-remove.app

# Remove shell completions (if installed)
rm -f ~/.zsh/completion/_claw-remove
rm -f ~/.bash_completion.d/claw-remove
```

### Linux (deb)

```bash
# Remove package
sudo dpkg --remove claw-remove

# Or completely purge (including configs)
sudo dpkg --purge claw-remove

# Manual cleanup (if installed from tarball)
sudo rm -f /usr/local/bin/claw-remove
sudo rm -f /usr/share/man/man1/claw-remove.1.gz
```

### Linux (rpm)

```bash
# Remove package
sudo rpm -e claw-remove

# Manual cleanup (if installed from tarball)
sudo rm -f /usr/bin/claw-remove
```

### Linux (AppImage)

```bash
# Simply delete the AppImage file
rm -f claw-remove-*.AppImage

# Remove desktop integration (if any)
rm -f ~/.local/share/applications/claw-remove.desktop
rm -f ~/.local/share/icons/claw-remove.png
```

### Windows

```powershell
# Remove from the installation directory
Remove-Item -Recurse -Force "C:\Tools\claw-remove"

# Or if added to PATH only
Remove-Item "C:\Tools\claw-remove\claw-remove.exe"
```

### From Source (go install)

```bash
# Remove binary
rm -f $(go env GOPATH)/bin/claw-remove

# Clean module cache (optional)
go clean -modcache
```

### Verify Complete Removal

```bash
# Should return "command not found"
which claw-remove || echo "ClawRemove successfully removed"
```

## Roadmap

The long-term roadmap lives in [docs/PLAN.md](./docs/PLAN.md).

That document is written in English on purpose so human contributors and autonomous agents can use the same source of truth for continued development.

## Contributing

We welcome contributions! Please see [docs/PROVIDER_AUTHORING.md](./docs/PROVIDER_AUTHORING.md) for how to add new AI agent providers.

## License

MIT License - see [LICENSE](./LICENSE) for details.
