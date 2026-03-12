# ClawRemove Architecture

## Purpose

This document defines the architectural direction for ClawRemove as a controlled, cross-platform claw removal engine.

It is not just a code map. It records:

- what is already structurally correct
- what is still too tightly coupled
- what the target architecture should become
- how future agents should evolve the code without breaking safety

## Current Assessment

The current architecture is directionally good, not bloated, and already better than a script-based remover.

Why it is good today:

- discovery, planning, execution, verification, and advisory concerns are split into separate packages
- provider-specific facts already live under `internal/products/*`
- the LLM path is advisory-only and does not own destructive execution
- verification is no longer just a second audit pass
- provider capabilities now exist as a first-class concept

Where it is still imperfect:

- `internal/core/engine.go` still orchestrates too much directly
- `internal/llm/reactor.go` contains tool protocol, tool execution, and prompt logic in one package
- provider skills and tools are still metadata catalogs rather than fully isolated runtime contracts
- `plan` still builds from raw discovery instead of a stronger evidence graph
- `verify` is useful, but not yet a full evidence engine with traceable provenance
- `platform` exists, but is still a thin host abstraction rather than a true adapter layer

Conclusion:

The codebase is not badly bloated, but some packages are beginning to accumulate too many responsibilities. This is the right time to lock in the target architecture before more features pile on.

## Reference Influence

After comparing the local `openclaw`, `nanobot`, and `picoclaw` trees, a few architectural patterns are worth keeping:

- provider-centric organization
- explicit `skills` and `tools` domains
- separation between product logic and transport or UI layers
- room for multiple model providers without tying the whole system to one API

ClawRemove should borrow those structural strengths without inheriting their runtime sprawl.

Because ClawRemove has a much narrower mission than those projects, its ideal architecture should stay smaller and stricter:

- fewer subsystems
- stronger deterministic boundaries
- clearer removal-focused domain modeling

## Architectural Principles

- deterministic engine owns execution
- providers own product knowledge
- platform adapters own OS-specific behavior
- verification owns evidence strength and leftover classification
- LLMs own explanation, hypothesis, and investigation guidance only
- read-only tools must stay strictly separate from destructive actions
- documentation must evolve with the architecture

## Current Layer Map

```text
cmd/claw-remove
  -> internal/app
  -> internal/core
     -> internal/discovery
     -> internal/plan
     -> internal/executor
     -> internal/verify
     -> internal/llm
     -> internal/products
     -> internal/platform
     -> internal/output
```

This is already understandable and mostly clean.

The main risk is that `core`, `llm`, and `plan` become accumulation points if new behavior is added without another structural pass.

## Target Architecture

The target structure should become:

```text
cmd/claw-remove

internal/app
internal/core
internal/discovery
internal/evidence
internal/plan
internal/executor
internal/verify
internal/platform
internal/output

internal/products/openclaw
internal/products/<future-provider>

internal/skills
internal/tools

internal/llm
internal/llm/prompts
internal/llm/providers
internal/llm/reactor
```

## Target Responsibilities

### `app`

- CLI parsing
- interactive confirmation flow
- command composition

### `core`

- orchestration only
- no product-specific rules
- no prompt-specific logic
- no OS-specific command rules

### `discovery`

- collect raw findings from filesystem, packages, services, processes, and related sources
- avoid making strong semantic claims

### `evidence`

This layer should be added next.

It should turn discovery output into explicit evidence objects:

- exact
- strong
- heuristic
- provenance
- related provider rule

This has now started landing as `internal/evidence`, but it still needs richer provenance and stronger integration across planning and verification.

### `plan`

- build actions from evidence, not just from raw discovered items
- keep destructive action policy deterministic
- keep opt-in boundaries explicit

### `executor`

- perform approved actions
- never call the LLM
- remain idempotent where possible

### `verify`

- classify leftovers after audit or apply
- summarize confirmed versus investigate-only residue
- remain deterministic

### `platform`

This should grow into true adapters:

- `platform/darwin`
- `platform/linux`
- `platform/windows`

OS-specific logic should move there over time instead of staying mixed inside discovery, plan, or probes.

### `products/*`

Each provider should own:

- facts
- aliases
- package refs
- service naming rules
- verification special cases
- capabilities

### `skills`

Provider skills should become a real contract, not only metadata.

Examples:

- residue analysis
- safe removal review
- legacy alias reconciliation
- post-removal verification review

### `tools`

Provider tools should become a real contract for controlled investigation.

Examples:

- path probes
- service probes
- package probes
- shell profile probes

All tools must be:

- read-only
- bounded to already discovered targets
- auditable
- model-safe

### `llm`

The LLM system should eventually split into:

- provider routing
- prompt construction
- reactor loop
- tool mediation
- structured output validation

The current implementation is good enough for now, but it is still concentrated in a single `reactor.go`.

The split has started with dedicated `prompts` and `providers` packages, but tool mediation and reactor control still need a cleaner boundary.

## Coupling Review

### Good decoupling already present

- provider facts are not hardcoded into engine
- execution is separated from planning
- LLM is not the executor
- verification is separated from execution

### Coupling that still exists

- `plan` depends directly on discovery layout rather than a richer evidence layer
- `llm/reactor.go` knows too much about tool names and tool execution
- provider capabilities are defined in product packages but not yet strongly consumed by `skills` and `tools`
- some OS logic still lives in discovery and plan rather than a dedicated platform adapter

## Controlled AI Boundary

This boundary is mandatory and should not be weakened:

- the LLM may explain findings
- the LLM may request read-only follow-up probes
- the LLM may suggest what should be reviewed
- the LLM may not authorize deletion
- the LLM may not execute shell commands
- the executor must remain deterministic

This is what lets ClawRemove behave like a smart agent without turning into another unsafe resident claw.

## Recommended Next Refactor

The next architectural refactor should follow this order:

1. Deepen `internal/evidence` so it carries provenance, provider rules, and stronger planning inputs.
2. Finish splitting the LLM subsystem into `prompts`, `providers`, and `reactor`.
3. Move OS-specific logic toward `internal/platform/*`.
4. Turn `internal/skills` and `internal/tools` into real contracts, not only catalogs.
5. Make `plan` consume evidence everywhere instead of relying on raw discovery lookups.

## Architecture Decision For This Version

For the current version, ClawRemove is considered:

- clear enough to continue building on
- not yet bloated
- partially decoupled, but not fully
- strong enough to support OpenClaw well
- ready for disciplined evolution, not uncontrolled feature growth

That means the architecture should now prioritize refinement over adding random surface area.
