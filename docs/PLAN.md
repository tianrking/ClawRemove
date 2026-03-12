# ClawRemove Development Plan

## Strategic Directions

ClawRemove has three core directions that provide real value:

### 1. AI Environment Inspector (Primary)

The most reasonable direction - scan and inventory AI environments:

```bash
claw-remove audit
```

Outputs:
- AI runtimes detected (Ollama, LM Studio, GPT4All)
- AI frameworks installed (langchain, openai sdk, transformers)
- AI model artifacts (llama models, cached models)
- AI tools installed (cursor, openclaw, windsurf)
- Storage analysis (model cache sizes)

Similar to: `docker info`, `brew doctor`

### 2. AI Tool Uninstaller (Original)

Deep uninstall for AI tools:

```bash
claw-remove apply --product cursor
```

Supports:
- Cursor
- OpenClaw
- Windsurf
- Aider
- Cline
- Continue.dev

### 3. AI Machine Hygiene Tool (Future)

Check AI resource usage:

```bash
claw-remove hygiene
```

Reports:
- AI storage: 86GB (model caches)
- Vector databases running
- GPU memory usage by AI processes
- Orphaned model files

## Latest Progress

- **Multi-Provider Architecture**: ClawRemove is now an Agent Removal Framework supporting multiple AI agents.
- **New Providers Added**: NanoBot and PicoClaw providers added alongside OpenClaw.
- **Windows Registry Support**: Added Windows registry detection and cleanup capabilities.
- **Environment Variable Detection**: Added environment variable discovery for all platforms.
- **Hosts File Detection**: Added hosts file entry detection.
- **Enhanced Service Discovery**: Improved macOS launchd, Linux systemd (including timers/sockets), and Windows scheduled task detection.
- **Platform Adapter Interface**: Extended with registry, environment, and hosts file methods.
- **Unit Tests**: Added comprehensive unit tests for executor and discovery packages.
- **Runner Interface**: Refactored system.Runner to interface for better testability.

## Mission

ClawRemove exists to be the cleanest and most trustworthy AI agent removal engine on macOS, Linux, and Windows.

The product goal is narrow by design:

- discover product-owned artifacts with strong evidence
- generate an auditable removal plan
- execute only approved actions
- verify what still remains after removal

ClawRemove is not a generic "PC cleaner", not a system tuner, and not a resident management agent.

It should behave like a controlled uninstall tool:

- smart enough to reason about AI agent footprints
- strict enough to avoid unbounded system changes
- quiet enough to leave no new mess behind

## Supported AI Agents

ClawRemove currently supports removal of:

| Provider | State Dir | Config | Env Prefix | Default Port | Package Manager |
|----------|-----------|--------|------------|--------------|-----------------|
| OpenClaw | ~/.openclaw | openclaw.json | OPENCLAW_ | 18789 | npm, brew |
| NanoBot | ~/.nanobot | config.json | NANOBOT_ | 18790 | pip, pipx |
| PicoClaw | ~/.picoclaw | config.json | PICOCLAW_ | 18790 | binary |
| OpenFang | ~/.openfang | openfang.json | OPENFANG_ | 18791 | npm |
| ZeroClaw | ~/.zeroclaw | zeroclaw.json | ZEROCLAW_ | 18792 | npm |
| NanoClaw | ~/.nanoclaw | nanoclaw.json | NANOCLAW_ | 18793 | npm, pip |

## Future Expansion

ClawRemove is designed to be extensible. Future providers may include:

### AI Agent Frameworks
- LangGraph agents
- CrewAI
- AutoGen
- GPT Researcher
- AgentGPT

### Local LLM Tools
- Ollama
- LM Studio
- GPT4All
- LocalAI

### Development Tools
- Cursor
- Windsurf
- Continue.dev
- Aider
- Cline

The architecture allows adding new providers without modifying core engine code.

## Product Direction

ClawRemove should feel predictable, surgical, and quiet:

- no hidden state database by default
- no background service
- no telemetry by default
- no broad wildcard deletion
- no destructive action without explicit reasoning
- no LLM authority over destructive execution

The long-term direction is a reusable removal engine with pluggable product providers for AI agents.

## Non-Negotiable Engineering Standards

- Evidence before deletion. Every destructive action must have traceable evidence.
- Plan before apply. Users and agents must be able to audit intended changes before execution.
- High-risk actions require explicit opt-in.
- Idempotent execution. Re-running the same plan should not cause new damage.
- Provider isolation. Product-specific rules must stay inside provider packages.
- Platform isolation. macOS, Linux, and Windows behavior must not leak into generic engine code.
- Human-readable and machine-readable output must remain stable.
- Any future LLM integration must stay advisory unless deterministic evidence promotes an action into the executable plan.

## Architectural Target

```text
cmd/claw-remove
internal/app
internal/core
internal/discovery
internal/evidence
internal/plan
internal/executor
internal/output
internal/platform
internal/products/openclaw
internal/products/nanobot
internal/products/picoclaw
internal/products/<future-provider>
internal/skills
internal/tools
docs
scripts
```

### Core responsibilities

- `app`: command model and CLI flags
- `core`: orchestration for audit, plan, apply, and verify
- `discovery`: provider-aware evidence collection
- `plan`: safe action generation with risk grading
- `executor`: execution of file, command, and service actions
- `output`: text and JSON reporting
- `platform`: platform-native adapters and system integrations
- `products/*`: provider facts, aliases, heuristics, and verification logic
- `skills`: provider-declared high-level analysis capabilities
- `tools`: controlled read-only tool catalog used by advisors

## Current State

Implemented today:

- provider registry with multi-provider support
- `openclaw`, `nanobot`, `picoclaw` providers
- audit, plan, apply, verify, explain commands
- JSON and human-readable output
- multi-platform build scripts
- baseline CI
- multilingual README set
- controlled advisor scaffold and `explain` command
- multi-provider LLM client for OpenAI, Anthropic, and OpenAI-compatible APIs
- residual verification classifier with confirmed versus investigate buckets
- provider capability model with provider-specific skills and tools
- explicit `internal/evidence` layer and partial LLM split into prompts/providers/reactor
- platform adapters (darwin/linux/windows) wired into controlled probes, discovery, and planning paths
- explicit architecture assessment in `docs/ARCHITECTURE.md`
- multi-driver LLM configuration and fallback-chain capability
- version metadata injected into binaries
- generated checksums and release archives
- Windows registry detection and cleanup
- Environment variable detection
- Hosts file entry detection
- Enhanced service discovery (launchd, systemd, scheduled tasks)
- Comprehensive unit tests for executor and discovery

## Delivery Phases

### Phase 1: Release-Ready CLI ✅

Goal: ship a reliable CLI that can be used in real environments for AI agent removal.

Status: Complete

### Phase 2: Engine Hardening ✅

Goal: make the core engine robust enough for multiple providers.

Status: Complete

### Phase 3: Controlled AI Advisor ✅

Goal: add ReAct-style assistance without turning ClawRemove into an unsafe autonomous agent.

Status: Complete

### Phase 4: Multi-Provider Expansion 🔄

Goal: support additional AI agent products without compromising safety.

Work items:

- [x] OpenClaw provider
- [x] NanoBot provider
- [x] PicoClaw provider
- [ ] AutoGPT provider
- [ ] LangGraph agents provider
- [ ] Ollama agents provider
- [x] Provider fixtures and conformance tests
- [x] Provider-specific verification rules

### Phase 5: Desktop Controller Readiness

Goal: prepare the engine for a future GUI or upper-computer controller.

Work items:

- [ ] stabilize internal request and response models
- [x] preserve clean separation between engine and CLI rendering
- [ ] ensure every action can be surfaced in UI with reason and risk
- [x] document machine-consumable output contracts

## Immediate Backlog

Priority order for the next development iterations:

1. Add AutoGPT provider
2. Add LangGraph agents provider
3. Add Ollama agents provider
4. Improve Windows registry cleanup coverage
5. Add GUI/daemon removal detection

## Rules For Agents

Agents working on ClawRemove should preserve these constraints:

- do not add a background process
- do not add telemetry by default
- do not add persistent caches unless explicitly justified
- do not broaden deletion to fuzzy filesystem scans
- do not let product-specific rules leak into generic engine packages
- do not mark heuristic findings as auto-removable without review
- do not let an LLM directly own destructive execution

When in doubt, prefer reporting over deletion.

## Definition of Done

A feature is done only when all of the following are true:

- behavior is documented
- risk is explicit
- output is stable
- tests or validation steps were run
- the change does not weaken evidence standards
- the change does not make the tool noisier or more invasive

## Documentation Contract

`README.md` explains what ClawRemove is and how to use it.

`docs/PLAN.md` explains where the project is going and how agents should continue development safely.

`docs/ARCHITECTURE.md` explains how the system is structured today, what is still too coupled, and what the target architecture should become.

This file should stay current whenever major architecture or roadmap decisions change.

When code structure, provider capabilities, or advisor behavior changes, `README.md`, `docs/PLAN.md`, and `docs/ARCHITECTURE.md` must be updated in the same change.
