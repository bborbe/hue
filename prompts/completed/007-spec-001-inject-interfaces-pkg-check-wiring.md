---
status: completed
spec: [001-inject-interfaces-pkg-check]
summary: Threaded the libtime.CurrentDateTimeGetter clock into check.NewChecksRunner via its constructor for DI parity with NewCheckCreator, wired the same clock instance through factory.CreateCheckController, and confirmed main.go constructs clock, Europe/Berlin location, and sunrise/sunset capability exactly once each
execution_id: hue-inject-interfaces-pkg-check-exec-007-spec-001-inject-interfaces-pkg-check-wiring
dark-factory-version: dev
created: "2026-09-06T12:00:00Z"
queued: "2026-09-06T10:24:48Z"
started: "2026-09-06T10:29:21Z"
completed: "2026-09-06T10:30:26Z"
---

# Runner clock constructor + end-to-end composition wiring

<summary>
- The checks runner becomes the last check-layer constructor to take the clock, completing the layer's DI pattern — it receives `libtime.CurrentDateTimeGetter` and retains it with no behavior change
- The factory passes the same clock instance to both the creator and the runner, so the whole checks path shares one clock
- The composition root is verified to construct the clock, the `Europe/Berlin` location, and the sunrise/sunset capability exactly once each
- All five constructor-DI acceptance-criteria greps pass end-to-end: runner takes the clock, factory threads it ≥3 times, and `main.go` owns clock, location, and capability construction
- The factory stays pure composition — no conditionals, no I/O, no `context.Background()` anywhere in it
- The runner's `RunChecks` method is untouched: no timing behavior, no new log output, so the cron-loop capture remains diff-comparable
- The repo compiles and `make precommit` stays green
</summary>

<objective>
Complete the constructor-injection wiring for the checks layer: the runner accepts the clock like the creator already does, the factory threads the clock to both, and the composition root in `main.go` is confirmed to construct the clock, the time-zone location, and the sunrise/sunset capability exactly once each. This satisfies the spec's AC 4 (constructor DI wired end-to-end) and removes the last DI inconsistency in `pkg/check`.
</objective>

<context>
Read first (the prompt depends on these being read before any edit):

- `./CLAUDE.md` — project conventions: `make precommit` gate, `errors.Wrap(ctx, err, "...")`, factory functions are pure composition, `github.com/bborbe/time` injected via `libtime.CurrentDateTimeGetter`.
- `./docs/dod.md` — DoD.
- `./.dark-factory.yaml` — `worktree: false`, `pr: false`, `autoRelease: false`; daemon handles git, agent MUST NOT commit.
- `./pkg/check/checks-runner.go` — the file being changed. Current state (verified):
  - `func NewChecksRunner() ChecksRunner` — no parameters today; this prompt adds the clock.
  - `type checksRunner struct { }` — gains the clock field.
  - `RunChecks(ctx context.Context, checks Checks) error` — MUST NOT change (no timing behavior, no new log output; spec DB 3).
- `./pkg/factory/factory.go` — `CreateCheckController` (signature extended by the previous prompt) calls `check.NewCheckCreator(...)` with five arguments and `check.NewChecksRunner()` with none; this prompt threads the clock into the runner.
- `./main.go` — `application.Run` (composition root) constructs `libtime.NewCurrentDateTime()`, resolves `time.LoadLocation("Europe/Berlin")`, and constructs `pkg.NewSunriseSunsetProvider()` — each exactly once, per the previous prompt. Verify only; no edits expected unless a check below fails.
- `./pkg/check/checks-creator.go` — reference for how the creator takes the clock: constructor parameter `currentDateTimeGetter libtime.CurrentDateTimeGetter` stored on the struct.
- Coding plugin docs (in-container paths): `/home/node/.claude/plugins/marketplaces/coding/docs/go-time-injection.md` (constructor accepts `CurrentDateTimeGetter`; `NewCurrentDateTime()` in factory → receive from caller), `/home/node/.claude/plugins/marketplaces/coding/docs/go-composition.md` and `/home/node/.claude/plugins/marketplaces/coding/docs/go-factory-pattern.md` (pure composition).

Library API verified against the module cache: `github.com/bborbe/time` v1.27.11 — `type CurrentDateTimeGetter interface { Now() DateTime }`, `func NewCurrentDateTime() CurrentDateTime`, `type DateTime stdtime.Time`.
</context>

<requirements>

### Step 1 — Runner constructor takes the clock (`pkg/check/checks-runner.go`)

1. Change the constructor from `func NewChecksRunner() ChecksRunner` to:

```go
func NewChecksRunner(currentDateTimeGetter libtime.CurrentDateTimeGetter) ChecksRunner {
	return &checksRunner{
		currentDateTimeGetter: currentDateTimeGetter,
	}
}
```

2. Change the struct from `type checksRunner struct {}` to:

```go
type checksRunner struct {
	currentDateTimeGetter libtime.CurrentDateTimeGetter
}
```

3. Add the import `libtime "github.com/bborbe/time"` to the file (the file currently imports only `"context"` and `"github.com/golang/glog"`).

4. Do NOT change `RunChecks` in any way — no timing behavior, no new log lines, no reordered lines (spec DB 3 + Constraints: the cron-loop capture must stay diff-comparable). The clock field is retained for structural DI parity with the creator; it is not consumed in this PR.

### Step 2 — Factory threads the clock to the runner (`pkg/factory/factory.go`)

5. In `CreateCheckController`, change the runner construction from `check.NewChecksRunner()` to `check.NewChecksRunner(currentDateTimeGetter)` — the same `currentDateTimeGetter` parameter that already flows into `check.NewCheckCreator(...)`. The factory gains no new parameters in this prompt; nothing else in `factory.go` changes.

6. Verify the factory is pure composition: read `factory.go` end-to-end and confirm `CreateCheckController`, `CreateBridgesProvider`, `CreateListLightsHandler`, and `CreateStatusHandler` contain no conditionals, no I/O, and no `context.Background()` (spec Constraints + `docs/dod.md`). Do not refactor anything — report only.

### Step 3 — Verify the composition root (`main.go`)

7. Read `application.Run` in `main.go` and confirm the three composition-root constructions are each present exactly once: `libtime.NewCurrentDateTime()` (clock), `time.LoadLocation("Europe/Berlin")` (location, error wrapped via `errors.Wrap(ctx, err, "load location failed")`), and `pkg.NewSunriseSunsetProvider()` (capability) — all passed into `factory.CreateCheckController(...)` (spec Desired Behavior 4). If any is missing or duplicated, STOP and report — main.go wiring belongs to the previous prompt; do not edit it.

### Step 4 — Self-check

8. Before finishing, re-run the `<verification>` block below and confirm every check passes; walk each acceptance criterion against the change. Run `make precommit` and confirm it exits 0.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git. No `git commit`, `git push`, `git add`, no PR.
- Do NOT touch `pkg/check/checks-creator.go`, `pkg/check/checks-creator_test.go`, or `pkg/sunrise-sunset.go` — done in prior prompts.
- Do NOT add a counterfeiter directive — mocks for the new interfaces are sibling PR 2 (spec Non-goal).
- `github.com/bborbe/errors` wrapping via `errors.Wrap(ctx, err, "...")`; never bare `return err` (spec Constraints).
- `github.com/bborbe/time` is injected via `libtime.CurrentDateTimeGetter` in constructors and created once in `main.go` (spec Constraints) — the runner constructor takes the getter, it must NOT call `libtime.NewCurrentDateTime()` or `time.Now()` itself.
- Factory functions stay pure composition: no conditionals, no I/O, no `context.Background()` (spec Constraints).
- Sunrise library injected via interface (frozen design rule, spec Constraints); the capability stays constructed exactly once in `main.go`.
- The time-zone location stays `Europe/Berlin`; no config surface is added (spec Constraints).
- No new log output in the checks-cron path (spec Constraints) — `RunChecks` must remain byte-identical in behavior and log lines.
- Existing tests keep passing; tests may be edited only for the new constructor parameters (spec Constraints). No test edits are expected in this prompt — if `make test` fails, STOP and report rather than editing tests.
- `make precommit` mandatory before finishing (spec Constraints).
- Never number the prompt filename — dark-factory assigns numbers on approve.
</constraints>

<verification>
All commands run from the repo root. Zero-match greps use a form with a truthful exit code.

1. `grep -n "libtime.CurrentDateTimeGetter" pkg/check/checks-runner.go` — must print ≥1 line (spec AC 4 evidence).
2. `grep -c "currentDateTimeGetter" pkg/factory/factory.go` — must print `3` (signature + creator call + runner call; spec AC 4 evidence).
3. `grep -n "libtime.NewCurrentDateTime()" main.go` — must print exactly 1 line (spec AC 4 evidence).
4. `grep -n "time.LoadLocation" main.go` — must print ≥1 line (spec AC 4 evidence).
5. `grep -rni "sunrisesunset" main.go pkg/factory/` — must print ≥1 line (spec AC 4 evidence; matched by `pkg.NewSunriseSunsetProvider()` in main.go).
6. `[ -z "$(grep -rn 'context.Background()' pkg/factory/)" ]` — must exit 0 (pure composition).
7. `[ -z "$(grep -rn 'time.Now()' pkg/check/)" ]` — must exit 0 (spec AC 1, still holds).
8. `[ -z "$(grep -rn 'sunrisesunset' pkg/check/ --include='*.go' | grep -v '_test.go')" ]` — must exit 0 (spec AC 2, still holds).
9. `[ -z "$(grep -rn 'LoadLocation' pkg/check/ --include='*.go' | grep -v '_test.go')" ]` — must exit 0 (spec AC 3, still holds).
10. `make precommit` — must exit 0 (DoD gate per `docs/dod.md`).
11. `make test` — must exit 0.
</verification>
