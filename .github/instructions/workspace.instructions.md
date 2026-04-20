---
name: Workspace Context & Standards
description: Project context, architecture, and coding standards for the Solace Cluster Manager Terraform provider.
applyTo: "**"
---

## Project Context

This is a **Terraform provider** (`terraform-provider-gsolaceclustermgr`) that manages Solace event brokers running in **Solace Cloud** via the Solace Mission Control API (v2). The provider allows users to create, read, update, and delete Solace broker instances in the cloud using Terraform or OpenTofu.

- The upstream REST API is large; only the subset required by this provider is used.
- The REST client code (`internal/missioncontrol/`) is **generated** using [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) from the OpenAPI spec in `api/missioncontrol_api_v2.json`. Do **not** hand-edit generated files.
- A **mock server** (`internal/fakeserver/`) is a hardcoded stub implementation of the relevant Solace Cloud API endpoints. It exists so that acceptance tests can run without real cloud credentials or broker quota. Tests must target the fake server, not the live API.

---

## Compatibility Requirements

> This is critical — do not break compatibility with any of these targets.

- **Terraform 1.5.x** — minimum supported version; avoid features that require newer Terraform core.
- **OpenTofu** — must be compatible with recent OpenTofu releases (which tracks Terraform 1.x compatibility).
- **Go** — use the version declared in `go.mod`. Do not introduce language features beyond that version.
- **terraform-plugin-framework** — use the version declared in `go.mod`. Follow the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) API, not the older `terraform-plugin-sdk`.

- The cloud service API  may silently evolve, so that calling the real service may fail / behave differently compared to calling the local fakeserver. We regularly need to check for api changes.


---

## Architecture

```
internal/
  provider/          # Terraform provider, resources, data sources, and tests
  missioncontrol/    # Generated REST client (do not hand-edit)
  fakeserver/        # Mock Solace Cloud API for testing
  fakeservercli/     # CLI entrypoint to run the fake server standalone
api/                 # OpenAPI spec (source of truth for the REST client)
examples/            # Example Terraform configurations
docs/                # Generated provider documentation
```

---

## Go Coding Standards

- Follow standard Go conventions: [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
- Use `context.Context` as the first argument to functions that perform I/O or long-running work.
- Prefer explicit error returns over panics.
- Error messages should be lowercase and not end with punctuation (Go convention).
- Keep functions small and focused; avoid deeply nested logic.
- Use `tflog` (from `github.com/hashicorp/terraform-plugin-log/tflog`) for logging — not `fmt.Println` or the standard `log` package.

---

## Terraform Plugin Framework Standards

- Follow the patterns established in the [Terraform Plugin Framework sample provider](https://github.com/hashicorp/terraform-provider-scaffolding-framework).
- Every resource must implement `resource.Resource` and `resource.ResourceWithConfigure`. Import state support (`resource.ResourceWithImportState`) should be added where it makes sense.
- Schema attribute names use `snake_case`.
- Use `types.String`, `types.Int32`, `types.Bool`, `types.List`, etc. — not raw Go primitives — for schema model fields.
- Use `plan modifiers` (`stringplanmodifier`, `int32planmodifier`, etc.) appropriately to signal computed vs. required fields.
- Provider data is passed via `CMProviderData` (defined in `provider.go`).
- For long-running operations (broker creation/deletion), poll with a timeout using the `PollingTimeoutDuration` and `PollingIntervalDuration` values from `CMProviderData`.

---

## Testing Standards

- Acceptance tests live in `internal/provider/` with the `_test.go` suffix.
- Tests **must** use the fake server (`internal/fakeserver/`) — do not require live Solace Cloud credentials for CI.
- Use `github.com/hashicorp/terraform-plugin-testing` for acceptance test helpers.
- Unit utility tests (e.g., `broker_util_test.go`) test pure functions without infrastructure.
- Test function names follow the pattern `TestAcc<Resource><Scenario>` for acceptance tests.

---

## What Not to Do

- Do **not** edit `internal/missioncontrol/client.go` by hand — regenerate it from the OpenAPI spec using `oapi-codegen`.
- Do **not** use `terraform-plugin-sdk` — this provider uses `terraform-plugin-framework` exclusively.
- Do **not** add Terraform features that are unavailable in Terraform 1.5.x or current OpenTofu.
- Do **not** call the real Solace Cloud API in tests unless explicitly asked for.
- Do **not** delete a solace live broker, really never.

---

## Local Development Toolchain

Tool versions (Go, Terraform, `gh` CLI, etc.) are managed via **[mise](https://mise.jdx.dev/)** and declared in `.tool-versions` at the repo root.

`mise` is **not** automatically activated in non-login shells on this machine. Before running any terminal commands that require managed tools (`go`, `terraform`, `gh`, `oapi-codegen`, etc.), activate mise first:

```bash
eval "$(mise activate bash)"
```

On PowerShell / Windows:
```powershell
mise activate pwsh | Out-String | Invoke-Expression
```

> If a tool is not found on PATH, this activation is almost certainly the cause. Do not try to install tools globally — activate mise instead.

### GitHub CLI (`gh`) Authentication

The repo belongs to the **GEBIT** GitHub organisation. After running `gh auth login`, the token may need to be explicitly authorized for the org (required if GEBIT enforces SAML SSO):

```bash
# Check auth status and whether the token is authorized for the org
gh auth status

# If the token is not authorized for GEBIT, re-authorize:
gh auth refresh --hostname github.com --scopes read:org
# Then visit the printed URL and authorize the token for GEBIT in the browser
```

For PR and workflow commands, always run `gh` from within the repo directory — it uses the `origin` remote from `.git/config` to target the correct repo automatically.

To list org repos explicitly:
```bash
gh repo list GEBIT
```
