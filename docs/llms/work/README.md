# Agent Work Directory

This folder is a **scratchpad for AI agents** working on this codebase. Files here (except this README) are gitignored.

## Purpose

Use this directory to:

1. **Track your work** - Create a worklog file for your session
2. **Explain your thinking** - Document decisions, trade-offs, and reasoning
3. **Record results** - Note what you tried, what worked, what didn't
4. **Plan before coding** - Draft approaches before implementation

## Worklog Format

Create a file named `worklog-YYYY-MM-DD-<brief-description>.md`:

```markdown
# Worklog: <Task Description>

**Date:** YYYY-MM-DD
**Agent:** <agent type>
**Ticket:** <reference to backlog item if applicable>

## Goal

<What you're trying to accomplish>

## Research Phase

### Existing Patterns Found
- <file:line> - <what it does, why it's relevant>

### SDK/API Verification
- <command run> - <result>

### Design Decisions
- <decision> - <reasoning>

## Implementation Log

### Attempt 1
- **Approach:** <what you tried>
- **Result:** <what happened>
- **Verification:** <how you confirmed it works/doesn't>

## Final Status

- [ ] Code compiles (`go build ./...`)
- [ ] Tests pass (`go test ./...`)
- [ ] New tests written for new behavior
- [ ] Existing patterns followed
- [ ] No assumptions about APIs - all verified

## Notes for Future Work

<Anything the next agent should know>
```

## Rules

1. **Research before implementing** - Read existing code, run `go doc`, verify APIs exist
2. **Verify as you go** - Compile after each significant change
3. **Document failures** - Failed approaches are valuable information
4. **No box-checking tests** - Tests should verify behavior, not just exist for coverage
