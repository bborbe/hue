---
status: completed
summary: Added counterfeiter generate directives for CheckCreator, ChecksRunner, and BridgesProvider, regenerated the three fakes into mocks/, and updated CHANGELOG.md
execution_id: hue-counterfeiter-mocks-exec-009-counterfeiter-mocks
dark-factory-version: dev
created: "2026-09-06T16:22:01Z"
queued: "2026-09-06T16:22:01Z"
started: "2026-09-06T16:24:12Z"
completed: "2026-09-06T16:25:46Z"
---

# Counterfeiter mocks for check interfaces

<summary>
- Test doubles generated for the three interfaces introduced by PR 1
- Generation wired via counterfeiter `//go:generate` directives matching the repo's existing pattern
- Fakes land in the existing mocks package alongside the current check and trigger fakes
- No production behavior change — directives plus regenerated fakes only
- Counterfeiter version verified aligned between the tool-version file and the module file; no version edit expected
- All existing tests and lint checks continue to pass
- Closes the open `counterfeiter-mocks-required` code-review finding
</summary>

<objective>
Give the `CheckCreator`, `ChecksRunner`, and `BridgesProvider` interfaces from PR 1 their own counterfeiter fakes, so unit tests can mock them and the interface-definition → test-coverage gap flagged by the review finding is closed.
</objective>

<context>
Read CLAUDE.md for project conventions (`make precommit` is mandatory before commit; dark-factory handles git).

The repo generates mocks with counterfeiter v6.12.2. `make generate` wipes and regenerates `mocks/` by running `go generate -mod=mod ./...`, which triggers `//go:generate go run -mod=mod github.com/maxbrunsfeld/counterfeiter/v6 -generate` in each package's `*_suite_test.go`. Counterfeiter's `-generate` mode emits fakes ONLY for interfaces carrying an explicit `//counterfeiter:generate` directive.

Existing directive pattern to follow (already in the repo):
- `pkg/check/check.go`: `//counterfeiter:generate -o ../../mocks/check.go --fake-name Check . Check`
- `pkg/trigger/trigger.go`: `//counterfeiter:generate -o ../../mocks/trigger.go --fake-name Trigger . Trigger`

The `-o` output path is relative to the package directory: packages under `pkg/check` use `../../mocks/`, the package `pkg` itself uses `../mocks/`.

The three interfaces that still lack directives and fakes:
- `CheckCreator` in `pkg/check/checks-creator.go` (package `check`)
- `ChecksRunner` in `pkg/check/checks-runner.go` (package `check`) — the name is `ChecksRunner`, plural, NOT `CheckRunner`
- `BridgesProvider` in `pkg/bridges-provider.go` (package `pkg`)

tools.env pins `COUNTERFEITER_VERSION ?= v6.12.2`; go.mod resolves `github.com/maxbrunsfeld/counterfeiter/v6` to v6.12.2 (indirect). The suite_test.go `//go:generate` lines pin no `@version` — they resolve via go.mod, so no version alignment change is needed.
</context>

<requirements>
1. Add a `//counterfeiter:generate` directive on a new line immediately above the `type CheckCreator interface {` line in `pkg/check/checks-creator.go`:
   `//counterfeiter:generate -o ../../mocks/checks-creator.go --fake-name CheckCreator . CheckCreator`
2. Add the same directive above the `type ChecksRunner interface {` line in `pkg/check/checks-runner.go`:
   `//counterfeiter:generate -o ../../mocks/checks-runner.go --fake-name ChecksRunner . ChecksRunner`
3. Add the same directive above the `type BridgesProvider interface {` line in `pkg/bridges-provider.go` (this package is one level shallower, so the output path is `../mocks/`):
   `//counterfeiter:generate -o ../mocks/bridges-provider.go --fake-name BridgesProvider . BridgesProvider`
4. Run `make generate`. Confirm the three fake files now exist — `mocks/checks-creator.go`, `mocks/checks-runner.go`, `mocks/bridges-provider.go` — and declare types `CheckCreator`, `ChecksRunner`, `BridgesProvider` in `package mocks`.
5. Do NOT modify the interface method signatures, the existing fakes (`mocks/check.go`, `mocks/trigger.go`), or `mocks/mocks.go`'s package declaration. Do NOT touch `main_test.go` files or any `cmd/` code — out of scope.
6. Confirm `COUNTERFEITER_VERSION` in `tools.env` (v6.12.2) still matches the counterfeiter version resolved in go.mod — no version edit expected.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git
- Follow the existing directive format exactly (see `pkg/check/check.go` and `pkg/trigger/trigger.go`) — do not invent a different shape
- Existing tests must still pass
- Do NOT hand-add license headers to generated files — `make precommit`'s addlicense step handles headers; counterfeiter's `Code generated ... DO NOT EDIT.` files are exempt
- Do NOT run `go mod vendor`; never use `-mod=vendor`
</constraints>

<verification>
Run `make precommit` — must pass.

Presence check: `ls mocks/checks-creator.go mocks/checks-runner.go mocks/bridges-provider.go` — all three must exist.
</verification>

Before you finish, re-run `<verification>` and confirm it passes; walk each requirement (1–6) against the actual change.
