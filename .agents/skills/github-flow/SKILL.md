---
name: github-flow
description: Feature branch → commit → squash merge workflow
metadata:
  workflow: github-flow
---

## What is GitHub Flow

A branching strategy where `main` is always deployable, and all changes are made on feature branches then merged via pull requests.

## Rules

- **No direct commits to main** — always work on feature branches
- **Commit per logical change** — separate commits for each concern, not one giant commit
- **Commit only to the current feature branch** — other features go in separate branches
- **Squash merge only** — `gh pr merge --squash --delete-branch`
- No force push, no merge on CI failure

## Branch Naming

Prefix (English) + `/` + kebab-case: `feat/`, `fix/`, `hotfix/`, `refactor/` (e.g. `feat/add-session-restore`)

## Commit Separation

- **One commit = one logical change** — separation of concerns
- **lint, typecheck, build, test must pass before commit**
- **One feature per branch** — don't mix multiple features in a single branch

## Commit Messages

`<prefix>: <description>` — both prefix and description in English (e.g. `feat: add session restore`, `fix: handle duplicate session names`)

### Prefixes

| Prefix | Purpose |
|--------|---------|
| `feat` | New feature |
| `fix` | Bug fix |
| `ui` | UI/style change (no functional change) |
| `refactor` | Code refactoring |
| `config` | Configuration/environment change |
| `docs` | Documentation change |
| `chore` | Build/package/cleanup |

## Workflow

1. **Create branch**: `git checkout main && git pull origin main && git checkout -b <type>/<name>`
2. **Develop & commit incrementally**: one commit per logical unit, ensure lint/typecheck/build/test pass
3. **Create PR**: `git push -u origin <branch> && gh pr create`
4. **Review & iterate**: address feedback, confirm CI passes
5. **Squash merge**: `gh pr merge --squash --delete-branch`
6. **Deploy**: auto-deploy via CI/CD after merge to main, or manual deploy

## Hotfixes

`hotfix/<name>` branch → minimal change commit → immediate PR → squash merge → postmortem after deploy

## Example: Add project config feature

```
# 1. Create feature branch from main
git checkout main && git pull origin main
git checkout -b feat/add-project-config

# 2. Develop & commit incrementally (one commit per logical unit)
# Add config struct
git add internal/config/config.go
git commit -m "feat: add project config struct"

# Implement TUI
git add internal/tui/
git commit -m "feat: add config screen UI"

# Fix bug
git commit -m "fix: fix config file parsing error"

# Add helper util
git add internal/util/
git commit -m "refactor: extract path utility"

# 3. Create PR
git push -u origin feat/add-project-config
gh pr create --title "feat: add per-project config" --body "YAML-based project session management"

# 4. Review & iterate (address feedback with additional commits)
git commit -m "refactor: simplify config loading logic"
git push

# 5. Squash merge
gh pr merge --squash --delete-branch

# 6. Deploy (CI/CD auto or manual)
```