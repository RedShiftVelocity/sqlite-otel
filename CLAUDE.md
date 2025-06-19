# CLAUDE.md - Development Rules

## Git Workflow Rules
1. **NEVER push to main branch** - Always use PR workflow
2. **One PR at a time** - Complete current PR before starting next
3. **Small PRs only** - Keep changes minimal and focused
4. **Local testing required** - All builds must pass before pushing
5. **Detailed PR descriptions** - Explain what changed and why

## Required Actions Per PR
- Test locally: `go build`, `go test ./...`, `go run`
- Update `DevJournal.md` with:
  - Actions taken
  - Technical decisions
  - Challenges/solutions
  - Key learnings
- Create descriptive branch names
- Write comprehensive PR descriptions

## DevJournal.md Format
```
## [Date] - PR #X: [Title]
### Actions: [what was done]
### Decisions: [technical choices made]  
### Challenges: [problems and solutions]
### Learnings: [insights gained]
```

## Pre-PR Checklist
- [ ] Code builds successfully
- [ ] Tests pass
- [ ] App runs locally
- [ ] DevJournal.md updated
- [ ] PR description complete
- [ ] Changes are minimal/focused

## Branch Strategy
- Create feature branches from main
- Use format: `feature/description` or `fix/description`
- Keep branches short-lived
- Delete after merge

**Remember: Quality over speed. Small, working changes are better than large, broken ones.**