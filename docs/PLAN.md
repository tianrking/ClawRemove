# ClawRemove Development Plan

## Latest Progress

- Added `internal/evidence` and moved verification/classification flow to consume evidence.
- Updated planning to consume evidence and attach full provenance to actions.
- Split LLM stack into clearer layers with `internal/llm/prompts` and `internal/llm/providers`.
- Split tool mediation into `internal/llm/mediation` so reactor focuses on orchestration.
- Added platform adapters (`darwin`, `linux`, `windows`) and routed controlled probes, discovery, and planning through adapters.
- Added multi-driver LLM support direction and implementation baseline for `openai`, `anthropic`, `openrouter`, `zhipu`, and `openai-compatible`.
- Added confidence-based planning policy that downgrades low-confidence destructive actions to report-only.
- Added optional provider/model routing trace output for multi-driver LLM chains.

## Mission

ClawRemove exists to be the cleanest and most trustworthy claw-family removal engine on macOS, Linux, and Windows.

The product goal is narrow by design:

- discover product-owned artifacts with strong evidence
- generate an auditable removal plan
- execute only approved actions
- verify what still remains after removal

ClawRemove is not a generic "PC cleaner", not a system tuner, and not a resident management agent.

It should behave like a controlled uninstall claw:

- smart enough to reason about claw-agent footprints
- strict enough to avoid unbounded system changes
- quiet enough to leave no new mess behind

## Product Direction

ClawRemove should feel predictable, surgical, and quiet:

- no hidden state database by default
- no background service
- no telemetry by default
- no broad wildcard deletion
- no destructive action without explicit reasoning
- no LLM authority over destructive execution

The long-term direction is a reusable removal engine with pluggable product providers.

The first provider is `openclaw`. Future providers can include other claw-family products, but OpenClaw remains the current quality benchmark.

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
internal/plan
internal/executor
internal/output
internal/platform
internal/products/openclaw
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

- provider registry
- `openclaw` provider
- audit, plan, apply, verify commands
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
- expanded OpenClaw legacy aliases and app paths

Still missing or incomplete:

- deeper platform-specific service discovery edge cases and test coverage
- stable contributor workflow for adding new providers
- a richer provider skill and tool authoring workflow
- richer provider/runtime tool contracts for `skills` and `tools`

## Delivery Phases

### Phase 1: Release-Ready OpenClaw CLI

Goal: ship a reliable CLI that can be used in real environments for OpenClaw removal.

Work items:

- strengthen `verify` so it is not only a second audit pass
- expand OpenClaw historical naming coverage
- refine risk labeling for actions
- add more exact and strong evidence markers
- improve Windows scheduled-task and service cleanup coverage
- improve Linux service variants and user/system scope coverage
- improve macOS launch agent and app residue coverage
- package version metadata into binaries
- generate checksums and release archives

Exit criteria:

- cross-platform build artifacts are reproducible
- all supported commands have stable JSON output
- README examples match real CLI behavior
- CI validates tests and build on each push

### Phase 2: Engine Hardening

Goal: make the core engine robust enough for multiple providers.

Work items:

- formalize `ProductProvider` and evidence interfaces
- formalize provider skills and tool contracts
- add a `platform` adapter layer
- add a dedicated `evidence` layer between discovery and planning
- split exact, strong, and heuristic evidence into explicit types
- standardize action metadata: reason, evidence, risk, opt-in requirement
- add regression tests around discovery and planning
- add snapshot tests for JSON output

Exit criteria:

- provider logic is isolated from engine logic
- new providers can be added without editing core planning semantics
- output schema changes are intentional and reviewed

### Phase 3: Controlled AI Advisor

Goal: add ReAct-style assistance without turning ClawRemove into an unsafe autonomous agent.

Work items:

- add an `internal/llm` package with provider-agnostic interfaces and multi-model routing
- define a strict tool schema for read-only evidence gathering
- keep destructive execution outside the LLM boundary
- add an `explain` or `advisor` flow for operator guidance
- make model output structured and machine-validated
- document safe prompt and tool rules
- let advisors consume provider-declared skills and tool inventories
- split the current LLM subsystem into prompts, providers, and reactor packages

Exit criteria:

- the LLM can explain findings without being able to directly mutate the system
- advisory output is deterministic enough to be reviewed and logged
- execution remains owned by the core engine
### Phase 4: Multi-Provider Expansion

Goal: support additional claw-family products without compromising safety.

Work items:

- document provider authoring rules
- add provider fixtures and conformance tests
- add one additional provider only after OpenClaw quality is strong
- add provider-specific verification rules

Exit criteria:

- adding a provider is a bounded change
- provider discovery does not regress existing products
- heuristics remain report-only unless promoted by strong evidence

### Phase 5: Desktop Controller Readiness

Goal: prepare the engine for a future GUI or upper-computer controller.

Work items:

- stabilize internal request and response models
- preserve clean separation between engine and CLI rendering
- ensure every action can be surfaced in UI with reason and risk
- document machine-consumable output contracts

Exit criteria:

- GUI can call the same engine flows without forking business logic
- output remains understandable to both humans and automation

## Immediate Backlog

Priority order for the next development iterations:

1. Turn `internal/skills` and `internal/tools` into runtime contracts.
2. Add GitHub release automation.
3. Add deeper platform adapter tests for service/process/listener edge cases.

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
