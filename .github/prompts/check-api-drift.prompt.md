---
name: check-api-drift
description: Download the latest Solace Mission Control API spec and compare it against the version stored in the repo. Identifies what needs changing in the provider.
argument-hint: Leave blank to auto-download the latest spec. Or provide a local path to a spec file to use instead.
agent: provider-maintainer
tools: ["read_file", "grep_search", "file_search", "run_in_terminal", "fetch_webpage"]
---

# Check API Drift

Download the latest Solace Mission Control OpenAPI spec and compare it against the version stored in the repository.
Identify what impact any changes have on this Terraform provider.

## Related prompts

- `/regen-client` — regenerate the REST client after updating the spec
- `/update-dependency` — update oapi-codegen or other dependencies
- `/fix-dependabot-pr` — if drift was discovered while fixing a dependabot PR

## Context

> **Toolchain:** Tool versions (`go`, `gh`, `oapi-codegen`) are managed by mise. Activate before running commands:
> ```powershell
> mise activate pwsh | Out-String | Invoke-Expression
> ```

- Current spec (in repo): `api/missioncontrol_api_v2.json`
- Spec listing page (source of truth for the download link): `https://api.solace.dev/cloud/page/openapi-specifications`
- Generated REST client: `internal/missioncontrol/client.go` (do NOT hand-edit — regenerated via oapi-codegen)
- Fake server (test stub): `internal/fakeserver/fakeserver.go` — hardcoded to support exactly the endpoints the provider uses
- Provider resources/datasources: `internal/provider/`

## Steps

1. **Download the latest spec**
   - If the user provided a local file path, use that as the updated spec and skip to sub-step (c).
   - Otherwise:
     - (a) Fetch the spec listing page to find the current download link:
       ```
       curl -fsSL https://api.solace.dev/cloud/page/openapi-specifications
       ```
       Parse the response for the link labelled **"Mission Control - v2.0 - JSON"**.
       Extract the `href` URL — do not assume it matches any previously known URL.
       > **Fallback:** If the page returns HTML that is hard to parse reliably, stop and ask the user to visit `https://api.solace.dev/cloud/page/openapi-specifications`, copy the URL of the "Mission Control - v2.0 - JSON" link, and paste it here.
     - (b) Download the spec using that URL:
       ```
       curl -fsSL <extracted-url> -o $env:TEMP\missioncontrol_api_v2_latest.json
       ```
       (On Linux/macOS: use `/tmp/missioncontrol_api_v2_latest.json`)
     - (c) Read `api/missioncontrol_api_v2.json` as the baseline (currently stored in repo).
   - Show the `info.version` or similar top-level metadata from both specs so the user can confirm which versions are being compared.

2. **Identify changed endpoints** (only for paths/operations the provider actually uses)
   - Grep `internal/provider/` for API call patterns to find which endpoints are in use.
   - For each used endpoint, compare request/response schemas between old and new spec.
   - Categorize changes as:
     - 🔴 **Breaking** — removed endpoint, removed required field, changed field type
     - 🟡 **Potentially breaking** — renamed field, changed enum values, new required field
     - 🟢 **Additive/safe** — new optional field, new endpoint (not used by provider)

3. **Assess fakeserver impact**
   - For each breaking or potentially breaking change, check if `internal/fakeserver/fakeserver.go` needs updating.
   - List specific structs/handlers that need changes.

4. **Assess client regen need**
   - If any used endpoint changed, recommend running the regen-client prompt after updating the spec.

5. **Output a summary table**

```
| Endpoint | Change Type | Severity | Provider Impact | Fakeserver Impact |
|----------|-------------|----------|-----------------|-------------------|
```

6. **Recommend next steps** in priority order.
