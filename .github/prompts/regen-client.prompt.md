---
name: regen-client
description: Regenerate the Solace Mission Control REST client from the OpenAPI spec using oapi-codegen, then check if the fakeserver needs updating.
argument-hint: Optional path to a new OpenAPI spec to use. Leave blank to regen from the existing api/missioncontrol_api_v2.json.
agent: provider-maintainer
tools: ["read_file", "replace_string_in_file", "multi_replace_string_in_file", "run_in_terminal", "grep_search", "file_search"]
---

# Regenerate REST Client

Regenerate `internal/missioncontrol/client.go` from the OpenAPI spec using oapi-codegen.

## Related prompts

- `/check-api-drift` — check if the Solace API spec has changed (often run before this)
- `/update-dependency` — update oapi-codegen or other dependencies
- `/fix-dependabot-pr` — if regen is needed as part of fixing a dependabot PR

## Context

> **Toolchain:** Tool versions (`go`, `gh`, `oapi-codegen`) are managed by mise. Activate before running commands:
> ```powershell
> mise activate pwsh | Out-String | Invoke-Expression
> ```

- **Spec:** `api/missioncontrol_api_v2.json`
- **Config:** `oapi-config.yaml` (package name, output path, generation options)
- **Output:** `internal/missioncontrol/client.go` — **never hand-edit this file**
- **Fake server:** `internal/fakeserver/fakeserver.go` — manually maintained stub; must be kept in sync with the endpoints the provider uses

## Steps

1. **Check git state**
   - Run `git status` and report any uncommitted changes.
   - If there are uncommitted changes unrelated to this task, warn the user and ask whether to proceed.

2. **Prepare the spec**
   - If the user provided a new spec path, copy it to `api/missioncontrol_api_v2.json` first (ask for confirmation before overwriting).
   - If you are continuing from a `/check-api-drift` run that downloaded a newer spec to a temp file, copy that file now:
     ```
     Copy-Item $env:TEMP\missioncontrol_api_v2_latest.json api\missioncontrol_api_v2.json
     ```
     (On Linux/macOS: `cp /tmp/missioncontrol_api_v2_latest.json api/missioncontrol_api_v2.json`)
   - Otherwise use the existing spec.

3. **Run oapi-codegen**
   ```
   go generate ./...
   ```
   Or directly:
   ```
   oapi-codegen --config oapi-config.yaml api/missioncontrol_api_v2.json
   ```
   - If `oapi-codegen` is not on PATH, check `tools/tools.go` and run `go install` for it.
   - If generation fails, diagnose whether the spec is malformed or the `oapi-codegen` version is incompatible.

4. **Verify the generated file**
   - Confirm `internal/missioncontrol/client.go` was updated (check modification time or diff).
   - Run `go build ./...` to ensure the generated code compiles.

5. **Check for fakeserver drift**
   - Compare the endpoint paths and request/response structs used in `internal/fakeserver/fakeserver.go` against the newly generated types in `internal/missioncontrol/client.go`.
   - List any struct fields or type names that changed and now mismatch the fakeserver.
   - Do NOT auto-edit the fakeserver — report the mismatches and ask before making changes.

6. **Run tests**
   ```
   TF_ACC=1 go test ./internal/provider/... -v -timeout 120s
   ```
   - Report any failures caused by the regen.

7. **Summary**
   Report: regen successful/failed, fakeserver mismatches found (if any), test result.
