---
name: fix-dependabot-pr
description: Checkout a failing dependabot PR, diagnose the failure from CI logs, plan a fix, and apply it after confirmation.
argument-hint: PR number to fix (e.g. 42). Leave blank to list open dependabot PRs and pick the oldest.
agent: agent
tools: ["read_file", "grep_search", "file_search", "run_in_terminal", "get_errors"]
---

# Fix Dependabot PR

Diagnose and fix a failing dependabot pull request for this Terraform provider.

## Context

> **Toolchain:** Tool versions (`go`, `gh`, `oapi-codegen`) are managed by mise. Activate before running commands:
> ```powershell
> mise activate pwsh | Out-String | Invoke-Expression
> ```

- Repo is `terraform-provider-gsolaceclustermgr`
- Tests require `TF_ACC=1` to run acceptance tests
- `internal/missioncontrol/client.go` is **generated** — never hand-edit it; use `go generate ./...`
- Tests use the fake server — no real Solace Cloud credentials needed
- Must remain compatible with **Terraform 1.5.x** and **OpenTofu**

## Related prompts

- `/update-dependency` — update a dependency and verify compat
- `/regen-client` — regenerate the REST client after an oapi-codegen bump
- `/check-api-drift` — check if the Solace API spec has changed

---

## Steps

### 1. Identify the PR

- If the user provided a PR number, use that.
- Otherwise, list open dependabot PRs oldest-first:
  ```
  gh pr list --author app/dependabot --state open --json number,title,createdAt,headRefName --jq 'sort_by(.createdAt) | .[] | "#\(.number) \(.createdAt[:10]) \(.title)"'
  ```
  Ask the user which PR to fix before proceeding.

### 2. Checkout the branch

```
gh pr checkout <PR number>
```

Show the user which branch is now checked out and what dependency is being bumped (read `go.mod` diff vs main).

### 3. Fetch CI failure logs

```
gh run list --branch <branch-name> --limit 5
gh run view <run-id> --log-failed
```

If no run exists yet, trigger one first:
```
gh workflow run test.yml --ref <branch-name>
gh run watch
```

### 4. Diagnose the failure

Analyse the logs and classify the root cause into one of these categories:

| Category | Indicators | Likely fix |
|----------|-----------|------------|
| **Build failure** | `go build` errors, undefined symbols | API change in dependency; fix call sites |
| **Generated client broken** | errors in `internal/missioncontrol/` | Run `/regen-client` |
| **Test failure** | `FAIL` in test output | Behaviour change; inspect failing test |
| **Plugin framework API change** | deprecated/removed interface method | Update resource/provider code |
| **Go toolchain bump** | `go.mod` `go` directive conflict | Check compat, may need toolchain update |
| **Terraform compat** | feature requires Terraform > 1.5.x | Do not adopt; find alternative |

Show the diagnosis clearly before proposing anything.

### 5. Present a fix plan — STOP AND WAIT FOR CONFIRMATION

Describe exactly what you will change:
- Which files will be modified
- Which commands will be run
- Whether any sub-prompts (`/regen-client`, etc.) will be invoked
- Any risks or unknowns

**Do not apply any changes until the user confirms.**

### 6. Apply the fix

After confirmation:
- Make the code changes
- Run `go build ./...` — must pass before continuing
- Run `TF_ACC=1 go test ./internal/provider/... -v -timeout 120s`
- If tests pass, commit and push:
  ```
  git add -A
  git commit -m "fix: resolve <dependency> bump breaking change"
  git push
  ```

### 7. Verify CI

```
gh run list --branch <branch-name> --limit 3
gh run watch
```

> **Note:** `gh run watch` is interactive. If it times out or is interrupted, use `gh run list --branch <branch-name>` to check status manually.

Report the final CI result.

### 8. Clean up

After CI passes and the PR is merged:
```
git checkout main
git pull
```
Do not delete the dependabot branch — GitHub closes it automatically on merge.
