---
name: update-dependency
description: Safely update a Go dependency in this Terraform provider, checking for breaking changes and verifying Terraform 1.5.x / OpenTofu compatibility.
argument-hint: Name of the dependency to update (e.g. github.com/hashicorp/terraform-plugin-framework). Leave blank to process all pending dependabot bumps.
agent: agent
tools: ["read_file", "run_in_terminal", "grep_search", "file_search", "fetch_webpage"]
---

# Update Dependency

Safely update one or more Go dependencies in this Terraform provider.

## Related prompts

- `/regen-client` — regenerate the REST client if oapi-codegen is updated
- `/check-api-drift` — check if the Solace API spec has changed
- `/fix-dependabot-pr` — fix a failing dependabot PR (often calls this prompt)

## Context

> **Toolchain:** Tool versions (`go`, `gh`, `oapi-codegen`) are managed by mise. Activate before running commands:
> ```powershell
> mise activate pwsh | Out-String | Invoke-Expression
> ```

- **Minimum supported Terraform:** 1.5.x — do not introduce plugin framework features requiring Terraform 1.6+
- **OpenTofu:** must remain compatible with recent OpenTofu releases
- **Go version:** use the version in `go.mod` — do not upgrade Go itself unless asked
- **Critical packages** with extra compatibility checks:
  - `github.com/hashicorp/terraform-plugin-framework` — check for deprecated APIs and new required interfaces
  - `github.com/hashicorp/terraform-plugin-testing` — test helper API changes
  - `github.com/oapi-codegen/oapi-codegen` — may require client regen if updated
  - `github.com/oapi-codegen/runtime` — runtime types used by generated client

## Steps

1. **Check git state**
   - Run `git status` and report any uncommitted changes.
   - If there are uncommitted changes, warn the user — a failed update mid-way could leave `go.mod`/`go.sum` in a broken state. Ask whether to proceed or stash first.

2. **Identify target dependency**
   - Read `go.mod` to confirm the current version.
   - If the user named a specific package, focus on that.
   - Otherwise, list all outdated direct dependencies:
     ```
     go list -u -m -f '{{if .Update}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}' all 2>/dev/null | grep -v '^$'
     ```

3. **Check release notes / changelog**
   - For `terraform-plugin-framework` and `terraform-plugin-testing`, fetch the GitHub releases page and scan for:
     - Breaking API changes
     - Deprecated functions/types
     - New required interface methods
     - Terraform core version requirements
   - For `oapi-codegen`, note if regeneration of `internal/missioncontrol/client.go` is needed.

4. **Perform the update**
   ```
   go get <package>@latest   (or specific version if user provided one)
   go mod tidy
   ```

5. **Scan for compilation issues**
   - Run `go build ./...`
   - If build fails, identify which files reference changed APIs and fix them.
   - Do **not** hand-edit `internal/missioncontrol/client.go` — if that file breaks, run the regen-client prompt instead.

6. **Run tests**
   ```
   TF_ACC=1 go test ./internal/provider/... -v -timeout 120s
   ```
   - Tests require the fake server — they do not need real Solace Cloud credentials.
   - Report any test failures with root cause analysis.

7. **Compatibility check**
   - Confirm no new plugin framework features are being used that require Terraform > 1.5.x.
   - Confirm `go.mod` `go` directive hasn't been bumped beyond what was requested.

8. **Summary**
   Report: old version → new version, any API changes found, any fixes applied, test result.
