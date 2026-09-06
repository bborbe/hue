---
status: completed
summary: Pinned all five counterfeiter go:generate directives to v6.12.2 and added -buildvcs=false to the five gexec.Build smoke tests; mocks regenerated identically and make precommit passes
execution_id: hue-counterfeiter-mocks-exec-010-pin-counterfeiter-version-and-buildvcs
dark-factory-version: dev
created: "2026-09-06T16:50:48Z"
queued: "2026-09-06T16:50:48Z"
started: "2026-09-06T16:52:39Z"
completed: "2026-09-06T16:54:40Z"
---

# Pin counterfeiter version and disable build VCS stamping in main_test.go

<summary>
- All `go:generate` directives that invoke counterfeiter now pin an explicit version, matching the version already declared in the project's tooling config
- This removes drift risk where counterfeiter could silently resolve to a different version than the one the project standardizes on
- All binary-compile smoke tests now build with VCS stamping disabled
- This avoids `go build` picking up local git metadata (dirty flags, commit hashes) during test runs, which is irrelevant for a compile-only check and can vary machine to machine
- No production behavior changes — these are build/tooling invocation tweaks only
- No interface or mock content changes — the counterfeiter directives generating fakes are untouched
- Mocks are regenerated as part of this change and are expected to be identical, since only the invocation version string changed, not the interfaces
</summary>

<objective>
Pin every counterfeiter `//go:generate` invocation to the exact version the project already standardizes on (`tools.env`'s `COUNTERFEITER_VERSION`), and disable VCS build-info stamping in the binary-compile smoke tests, so mock generation and compile checks are fully reproducible and independent of local git/tool-resolution state.
</objective>

<context>
Read CLAUDE.md for project conventions (`make precommit` is mandatory before finishing; dark-factory handles git — do not commit).

`tools.env` pins the counterfeiter version:
```
COUNTERFEITER_VERSION      ?= v6.12.2
```

Five `*_suite_test.go` files each carry an unversioned counterfeiter `go:generate` invocation at line 16:
```go
//go:generate go run -mod=mod github.com/maxbrunsfeld/counterfeiter/v6 -generate
```
in:
- `pkg/pkg_suite_test.go`
- `pkg/check/check_suite_test.go`
- `pkg/trigger/trigger_suite_test.go`
- `pkg/factory/factory_suite_test.go`
- `pkg/handler/handler_suite_test.go`

Five `main_test.go` files each call `gexec.Build` with only `-mod=mod` as an extra arg, at line 18:
- `main_test.go`: `gexec.Build("github.com/bborbe/hue/", "-mod=mod")`
- `cmd/create-token/main_test.go`: `gexec.Build("github.com/bborbe/hue/cmd/create-token", "-mod=mod")`
- `cmd/list-lights/main_test.go`: `gexec.Build("github.com/bborbe/hue/cmd/list-lights", "-mod=mod")`
- `cmd/turnon-light/main_test.go`: `gexec.Build("github.com/bborbe/hue/cmd/turnon-light", "-mod=mod")`
- `cmd/turnoff-light/main_test.go`: `gexec.Build("github.com/bborbe/hue/cmd/turnoff-light", "-mod=mod")`

`gexec.Build` (package `github.com/onsi/gomega/gexec`) takes the package path as the first argument and any number of additional `go build` flag strings as variadic trailing arguments — adding `-buildvcs=false` is a matter of appending one more string argument to the existing call.

`make generate` runs `go generate -mod=mod ./...`, which re-triggers every `go:generate` line above and rewrites `mocks/`. Since the interfaces annotated with `//counterfeiter:generate` are unchanged (already added and committed in a prior prompt), regeneration is expected to produce byte-identical output to what's currently in `mocks/` — only the invocation's version pin changes, not what gets generated.
</context>

<requirements>
1. In each of the 5 `*_suite_test.go` files listed above, change the `go:generate` line from:
   `//go:generate go run -mod=mod github.com/maxbrunsfeld/counterfeiter/v6 -generate`
   to:
   `//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6@v6.12.2 -generate`
   (drop `-mod=mod`: it is a no-op once the version is pinned, and this is the canonical form used by all sibling bborbe repos and documented in `tools.env`'s header comment)
   Match the version string exactly to `tools.env`'s `COUNTERFEITER_VERSION` value (`v6.12.2`). Change only this line in each file — no other content.
2. In each of the 5 `main_test.go` files listed above, add `-buildvcs=false` as an additional argument to the existing `gexec.Build(...)` call, alongside the existing `-mod=mod` argument. Do not remove or reorder the existing arguments; do not rewrite any other part of these files (they already contain a working `TestSuite` + `gexec.Build` compile check from a prior prompt).
   Example transform:
   ```go
   // before
   _, err = gexec.Build("github.com/bborbe/hue/", "-mod=mod")
   // after
   _, err = gexec.Build("github.com/bborbe/hue/", "-mod=mod", "-buildvcs=false")
   ```
3. Run `make generate` to regenerate `mocks/`. Confirm it completes without error and the diff (if any) is limited to incidental regeneration noise, not interface/behavior changes — do not hand-edit anything under `mocks/`. Note: `make generate` rewrites `mocks/mocks.go` to a bare `package mocks`, temporarily removing its license header — `make precommit`'s final `addlicense` step restores it. Do not restore it by hand.
4. Run `make precommit` and ensure it passes.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git
- Do NOT touch `mocks/` content beyond what `make generate` produces automatically
- Do NOT modify interface method signatures anywhere
- Do NOT re-add or alter the `//counterfeiter:generate` directives on `CheckCreator`, `ChecksRunner`, `BridgesProvider`, `Check`, or `Trigger` — those are already correct from a prior prompt
- Do NOT change any other line in the 5 `*_suite_test.go` files or the 5 `main_test.go` files besides the two edits described above
- Do NOT run `go mod vendor`; never use `-mod=vendor`
- `make precommit`'s `ensure` step (`go mod tidy -e`) will drop `github.com/maxbrunsfeld/counterfeiter/v6 v6.12.2 // indirect` from `go.mod` and its two lines from `go.sum` — this is expected and correct once the directive is version-pinned. Do NOT re-add them.
- Existing tests must still pass
</constraints>

<verification>
```bash
set -e
for f in pkg/pkg_suite_test.go pkg/check/check_suite_test.go pkg/trigger/trigger_suite_test.go pkg/factory/factory_suite_test.go pkg/handler/handler_suite_test.go; do
  grep -q 'counterfeiter/v6@v6.12.2 -generate' "$f" || { echo "MISSING version pin: $f"; exit 1; }
done
for f in main_test.go cmd/create-token/main_test.go cmd/list-lights/main_test.go cmd/turnon-light/main_test.go cmd/turnoff-light/main_test.go; do
  grep -q -- '-buildvcs=false' "$f" || { echo "MISSING -buildvcs=false: $f"; exit 1; }
done
make precommit
```
The two loops must print nothing (any `MISSING …` line means an edit was skipped). `make precommit` must pass.
</verification>

Before you finish, re-run `<verification>` and confirm it passes; walk each requirement (1–4) against the actual change.
