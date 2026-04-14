package codex

func codexAgentTeamsSkillBody() string {
	return `
# Agent Teams Skill

## Overview

This document is a reserved placeholder for a future native Codex multi-agent / teams surface.

**Activation flag**: ` + "`@auto go SPEC-ID --team`" + `

Today, Codex should continue to use the default ` + "`spawn_agent(...)`" + ` subagent pipeline. Do not reinterpret ` + "`--team`" + ` as extra parallel orchestration in the harness.

## Current Behavior

- ` + "`@auto go`" + ` without flags: use the default subagent pipeline, but if runtime policy blocks implicit spawning, ask before proceeding
- ` + "`@auto go --auto`" + `: treat this as explicit approval to run the default subagent pipeline without an extra confirmation round
- ` + "`@auto go --solo`" + `: disable subagents and stay in the main session
- ` + "`@auto go --team`" + `: keep compatibility with future native multi-agent naming, but continue with the default subagent pipeline for now

## Why This Is Reserved

- Codex already supports subagents natively via ` + "`spawn_agent(...)`" + `
- Public Codex docs do not define a separate local CLI Team API equivalent to Claude Code Agent Teams
- Overloading ` + "`--team`" + ` to mean "extra ` + "`spawn_agent(...)`" + ` fan-out" would conflict with the likely future meaning of native multi-agent support

## What To Use Instead

- Use ` + "`.codex/skills/agent-pipeline.md`" + ` for the default execution model
- Use ` + "`.codex/agents/*.toml`" + ` as the role source of truth for spawned workers
- Use ` + "`.codex/skills/worktree-isolation.md`" + ` when parallel ownership boundaries are explicit

## Revisit Condition

Enable a real ` + "`--team`" + ` route only when Codex exposes a documented native multi-agent surface that is distinct from ordinary subagent spawning.
`
}

func codexAgentPipelineSkillBody() string {
	return `
# Agent Pipeline Skill

Default multi-agent execution model for ` + "`@auto go`" + ` in Codex.

## Activation

This skill is the default for ` + "`@auto go SPEC-ID`" + `.

| Flag | Mode | Codex meaning |
|------|------|---------------|
| none | Subagent pipeline | Main session orchestrates specialists phase-by-phase |
| ` + "`--team`" + ` | Reserved compatibility flag | Keep the default subagent pipeline until Codex ships a documented native multi-agent surface |
| ` + "`--solo`" + ` | Single session | No worker spawning; implement directly in the main session |
| ` + "`--multi`" + ` | Multi-provider review | Run additional review/validation passes when configured, prefer orchestra-backed review when available |

See .codex/skills/agent-teams.md for the reserved ` + "`--team`" + ` policy and .codex/skills/worktree-isolation.md for parallel ownership rules.

## Codex Auto Semantics

- In Codex, ` + "`--auto`" + ` means "skip approval gates" and also counts as explicit approval for the default ` + "`spawn_agent(...)`" + ` subagent pipeline.
- Without ` + "`--auto`" + `, if the runtime policy blocks implicit worker spawning, the main session must explain the constraint and ask before switching to subagents.
- ` + "`--team`" + ` remains a reserved compatibility flag until Codex ships a distinct native multi-agent surface.

## Phase 0.5: Autonomy Policy

Before spawning workers, decide whether the pipeline can proceed autonomously:

- If ` + "`--auto`" + ` is set, continue without confirmation and treat it as explicit approval for the default subagent pipeline.
- If user intent is ambiguous, ask one concise plain-text question in the main session.
- Do not rely on Claude-only permission or question APIs.

## Pipeline Overview

` + "```text" + `
Phase 1:   Planning        -> planner
Phase 1.5: Test Scaffold   -> tester        (optional)
Gate 1:    Approval        -> main session  (skip with --auto)
Phase 1.8: Doc Fetch       -> main session  (fetch current docs if needed)
Phase 2:   Implementation  -> executor x N  (parallel only with disjoint ownership)
Gate 2:    Validation      -> validator
Phase 2.5: Annotation      -> annotator
Phase 3:   Testing         -> tester
Phase 3.5: UX Verify       -> frontend-specialist (optional)
Phase 4:   Review          -> reviewer + security-auditor
` + "```" + `

## Quality Mode

Quality mode influences model choice, not platform semantics:

- Ultra: pass ` + "`model=\"opus\"`" + ` to spawned workers
- Balanced: use each role's default model
- Adaptive: choose stronger models only for high-complexity tasks

Reference: .codex/skills/adaptive-quality.md

## Phase Guidance

### Phase 1: Planning

Spawn a planner when the task has enough scope to justify decomposition.

` + "```python" + `
spawn_agent(
    agent_type="planner",
    fork_context=True,
    message="""
    Read SPEC-XXX.
    Produce an execution table with task id, owner role, mode, and file ownership.
    Mark only truly independent tasks as parallel.
    """,
)
` + "```" + `

### Phase 1.5: Test Scaffold

When enabled, spawn a tester to write failing tests before implementation. Generated scaffold tests are read-only for later executors unless the plan explicitly reassigns them.

### Phase 1.8: Doc Fetch

This phase stays in the main session. Use current documentation tools available in the session and inject only the relevant excerpts into later worker prompts.

### Phase 2: Implementation

Parallel implementation is valid only with disjoint ownership.

` + "```python" + `
spawn_agent(
    agent_type="executor",
    fork_context=True,
    message="""
    Own only: pkg/auth/*.
    Follow TDD for task T1.
    Return changed files, tests run, and unresolved issues.
    """,
)
` + "```" + `

When workers return, review and integrate their results in the main session. Do not assume Codex auto-merges worktree branches.

### Gate 2: Validation

Spawn a validator after implementation lands. If validation fails, respawn a focused fixer instead of rerunning the full pipeline blindly.

### Phase 2.5: Annotation

Run annotator after validation PASS. Harness-only markdown changes may skip this phase.

### Phase 3 / 3.5: Testing and UX Verification

- Tester raises coverage and adds edge-case tests
- Frontend-specialist runs only when changed files include frontend UI

### Phase 4: Review

Run reviewer and security-auditor in parallel when the change scope justifies both. When ` + "`--multi`" + ` is set, prefer an additional orchestra-backed review/decision pass after local validation if the CLI/config supports it. Consolidate findings in the main session.

## Parallelism Rules

| Condition | Execution |
|----------|-----------|
| Non-overlapping ownership | Parallel workers allowed |
| Shared file or shared migration | Sequential execution |
| Order dependency between tasks | Sequential execution |
| One worker blocked on another's output | Wait, integrate, then continue |

## Retry Policy

- Validation: up to 3 retries, or 5 with ` + "`--loop`" + `
- Review: up to 2 retries, or 3 with ` + "`--loop`" + `
- Repeated worker failure: shrink scope or fall back to the main session

## Result Integration

Each worker should return:

- changed files
- verification run
- blockers or assumptions

The main session owns final integration, status updates, and the decision to continue, retry, or stop.
`
}
