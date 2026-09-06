---
status: approved
spec: [002-migrate-glog-slog]
created: "2026-09-06T20:55:00Z"
queued: "2026-09-06T19:18:14Z"
---

# Sampler wiring for checks-cron + list-lights aggregation

<summary>
- The checks-cron cycle loop's failure warning is now sampler-gated: `NewCheckCron` gains a `log.SamplerFactory` parameter, the sampler is created once at construction, and the `slog.Warn("run checks failed", ...)` line fires only when `sampler.IsSample()` is true
- The composition root passes a time sampler capped at most once per 10 minutes (`log.NewSampleTime(10 * time.Minute)`), threaded through the factory from `main.go` per the repo's DI convention
- Tests use `log.DefaultSamplerFactory` (no new counterfeiter mock, per the spec constraint) and prove the gate behavior: a failing runner produces exactly one warning line across many loop iterations, and a healthy runner produces none
- The list-lights CLI aggregates its per-light detail into a single post-loop emission instead of one log call per light — no sampler there, because sampling would truncate the CLI's output mid-listing
- The `github.com/bborbe/log` import returns to `main.go` (removed by the previous prompt) for the sampler factory
- `make precommit` + `make test` stay green
</summary>

<objective>
Resolve the two `no-tight-loop-without-sampler` findings: gate the checks-cron failure log through a `github.com/bborbe/log` sampler so no log line fires from the 60-second cycle loop unless something genuinely happened, and replace the list-lights per-iteration log call with one aggregated emission after the loop. This is spec 002's Desired Behavior 4 and completes the migration chain (depends on prompts 1 and 2 having landed).
</objective>

<context>
Read first (the prompt depends on these being read before any edit):

- `./CLAUDE.md` — project conventions: `make precommit` gate, `errors.Wrap(ctx, err, "...")`, factory functions are pure composition, `github.com/bborbe/time` for injected clocks (stdlib `time` for `time.Duration` / `time.NewTimer`, which checks-cron already uses).
- `./docs/dod.md` — DoD: exported symbols need doc comments; Ginkgo v2 / Gomega tests; CHANGELOG entry under `## Unreleased`.
- `./.dark-factory.yaml` — `worktree: false`, `pr: false`; the daemon handles git, the agent MUST NOT commit.
- `./specs/in-progress/002-migrate-glog-slog.md` — Desired Behavior 4 (sampler wiring + list-lights aggregation), Constraints (pure composition; tests use `log.DefaultSamplerFactory`, no new mock), Failure Modes (sampler gate on list-lights must not truncate the listing — revert to single aggregated emission), Assumptions (list-lights aggregation is the sanctioned fix; `NewCheckCron` gains the parameter, `pkg/factory` threads it, `main.go` passes it at the composition root, tests use `log.DefaultSamplerFactory`).
- `./pkg/check/checks-cron.go` — current state after the previous prompts of this spec: `func NewCheckCron(creator CheckCreator, runner ChecksRunner, interval time.Duration) run.Func` with the ungated `slog.Warn("run checks failed", "error", err)` inside the `for` loop; the two heartbeat lines are already gone. This prompt adds the `samplerFactory` parameter and gates the Warn.
- `./pkg/factory/factory.go` — `func CreateCheckController(url string, id string, token pkg.Token, inverval time.Duration, summerMode bool, currentDateTimeGetter libtime.CurrentDateTimeGetter, location *time.Location, sunriseSunsetProvider pkg.SunriseSunsetProvider) run.Func` calls `check.NewCheckCron(...)` with three arguments; this prompt adds the `samplerFactory log.SamplerFactory` parameter and threads it.
- `./main.go` — current state after the previous prompts: `Run` installs the slog default handler with the `LevelVar` (Debug), calls `factory.CreateCheckController(...)` then `a.createHttpServer(&logLevel)`; the `github.com/bborbe/log` import is currently absent. This prompt constructs the sampler factory in `Run` and re-adds the import.
- `./cmd/list-lights/main.go` — current state: `slog.Info("light state", "name", light.Name, "on", light.IsOn())` inside the `for _, light := range lights` loop (converted by the first prompt); this prompt aggregates it into one emission after the loop.
- `./mocks/checks-creator.go` and `./mocks/checks-runner.go` — counterfeiter fakes already present (sibling PR 2); the new test uses them (no new mock).
- `./pkg/check/check_suite_test.go` — package `check_test` Ginkgo suite (already bootstrapped; a new test file needs no suite changes).
- `./CHANGELOG.md` — append under the existing `## Unreleased` section.
- Coding plugin docs (in-container paths): `/home/node/.claude/plugins/marketplaces/coding/docs/go-logging-guide.md` (the `no-tight-loop-without-sampler` rule and its two sanctioned fixes — sampler gate AND "aggregate and log once after the loop"; the canonical constructor shape `log.SamplerFactory -> Sampler stored on struct`; the sampler table `NewSampleTime(d)` / `NewSampleMod(n)` / `NewSamplerGlogLevel(n)` / `SamplerList`), `/home/node/.claude/plugins/marketplaces/coding/docs/go-context-cancellation-in-loops.md` (the checks-cron loop already selects on `ctx.Done()` — keep that), `/home/node/.claude/plugins/marketplaces/coding/docs/go-composition.md` and `/home/node/.claude/plugins/marketplaces/coding/docs/go-factory-pattern.md` (pure composition), `/home/node/.claude/plugins/marketplaces/coding/docs/go-testing-guide.md`, `/home/node/.claude/plugins/marketplaces/coding/docs/changelog-guide.md`, `/home/node/.claude/plugins/marketplaces/coding/docs/go-precommit.md`.

Library APIs verified against the module cache (`github.com/bborbe/log` v1.6.25):

- `type Sampler interface { IsSample() bool }`
- `type SamplerFactory interface { Sampler() Sampler }`
- `type SamplerFactoryFunc func() Sampler` — implements `SamplerFactory` by calling itself; `type SamplerFunc func() bool` — implements `Sampler` by calling itself
- `var DefaultSamplerFactory SamplerFactory = SamplerFactoryFunc(func() Sampler { return SamplerList{NewSampleTime(10 * time.Second), NewSamplerGlogLevel(4)} })`
- `func NewSampleTime(duration stdtime.Duration) Sampler` — returns true at most once per duration; thread-safe; uses `github.com/bborbe/time` internally
</context>

<requirements>

### Step 1 — Sampler parameter on `NewCheckCron` (`./pkg/check/checks-cron.go`)

1. Change the constructor signature and create the sampler once at construction (the go-logging-guide's canonical shape — `SamplerFactory` in, `Sampler` retained):

```go
import (
	"context"
	"log/slog"
	"time"

	"github.com/bborbe/errors"
	"github.com/bborbe/log"
	"github.com/bborbe/run"
)

func NewCheckCron(
	creator CheckCreator,
	runner ChecksRunner,
	interval time.Duration,
	samplerFactory log.SamplerFactory,
) run.Func {
	sampler := samplerFactory.Sampler()
	return func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				checks, err := creator.CreateChecks(ctx)
				if err != nil {
					return errors.Wrapf(ctx, err, "create checks failed")
				}
				if err := runner.RunChecks(ctx, checks); err != nil {
					if sampler.IsSample() {
						slog.Warn("run checks failed", "error", err)
					}
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.NewTimer(interval).C:
				}
			}
		}
	}
}
```

2. Add the `"github.com/bborbe/log"` import. The `select` on `ctx.Done()` and the `time.NewTimer(interval)` structure stay exactly as-is (spec Security: the checks-cron loop already selects on `ctx.Done()` and keeps that behavior). The error path behavior stays: `CreateChecks` errors still abort via `errors.Wrapf(ctx, err, "create checks failed")`; `RunChecks` errors are logged (now sampler-gated) and the loop continues.

### Step 2 — Thread through the factory (`./pkg/factory/factory.go`)

Add the `"github.com/bborbe/log"` import and extend `CreateCheckController` by one parameter, threading it to `check.NewCheckCron` (pure composition — no conditionals, no I/O, no `context.Background()`):

```go
func CreateCheckController(
	url string,
	id string,
	token pkg.Token,
	inverval time.Duration,
	summerMode bool,
	currentDateTimeGetter libtime.CurrentDateTimeGetter,
	location *time.Location,
	sunriseSunsetProvider pkg.SunriseSunsetProvider,
	samplerFactory log.SamplerFactory,
) run.Func {
	return check.NewCheckCron(
		check.NewCheckCreator(
			CreateBridgesProvider(
				url,
				id,
				token,
			),
			summerMode,
			currentDateTimeGetter,
			location,
			sunriseSunsetProvider,
		),
		check.NewChecksRunner(currentDateTimeGetter),
		inverval,
		samplerFactory,
	)
}
```

### Step 3 — Composition root (`./main.go`)

1. Re-add the `"github.com/bborbe/log"` import (the previous prompt removed it — it is required again here).
2. In `func (a *application) Run(...)`, construct the sampler factory at the composition root and pass it to `factory.CreateCheckController(...)`:

```go
	samplerFactory := log.SamplerFactoryFunc(func() log.Sampler {
		return log.NewSampleTime(10 * time.Minute)
	})
```

and add `samplerFactory,` as the last argument to the `factory.CreateCheckController(...)` call (after `pkg.NewSunriseSunsetProvider()`).

AUDIT NOTE for the human reviewer (spec ambiguity): spec DB 4 says both "main.go passes `log.DefaultSamplerFactory` at the composition root" AND "capped at most once per 10 minutes (`log.NewSampleTime(10 * time.Minute)`)". These conflict — `DefaultSamplerFactory` samples at 10 seconds (plus a glog-verbosity gate that is inert after the migration), not 10 minutes. This prompt implements the 10-minute cap as stated twice in DB 4 (`NewSampleTime(10 * time.Minute)`), and reserves `DefaultSamplerFactory` for tests per the spec Constraint ("tests use `github.com/bborbe/log`'s `DefaultSamplerFactory`, no new counterfeiter mock"). If you prefer `main.go` to pass `DefaultSamplerFactory` verbatim, the production cap becomes 10 seconds — flag the choice.

### Step 4 — List-lights aggregation (`./cmd/list-lights/main.go`)

Replace the per-iteration log call with a single aggregated emission after the loop (spec DB 4: a sampler gate is NOT acceptable here because it would truncate the CLI's output mid-listing). Add `"fmt"` and `"strings"` to the imports, then:

```go
	slog.Info("found lights", "count", len(lights))
	var listing strings.Builder
	for _, light := range lights {
		fmt.Fprintf(&listing, "'%s' on: %v\n", light.Name, light.IsOn())
	}
	slog.Info("lights listing", "listing", listing.String())
	return nil
```

The loop body contains no log call — `no-tight-loop-without-sampler` on the listing loop is resolved by aggregation, not by a sampler. The `slog.Info` line count in `cmd/` stays ≥ 6 (list-lights keeps its two Info lines: "found lights" and "lights listing").

### Step 5 — Sampler-gate test (`./pkg/check/checks-cron_test.go`, new file)

Create `./pkg/check/checks-cron_test.go`, `package check_test`, Ginkgo v2 / Gomega, using the existing `mocks.CheckCreator` / `mocks.ChecksRunner` fakes and `log.DefaultSamplerFactory` (no new mock, per spec Constraint). This test traverses the real boundary the new code crosses: checks-cron dispatch → `sampler.IsSample()` → `slog.Warn` emission.

1. In `BeforeEach`: `creator := &mocks.CheckCreator{}` with `creator.CreateChecksReturns(check.Checks{}, nil)`; `runner := &mocks.ChecksRunner{}`; capture slog output — `buf := &bytes.Buffer{}`, `old := slog.Default()`, `slog.SetDefault(slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})))`. Restore with `slog.SetDefault(old)` in `AfterEach`.
2. It "logs the run-failed warning at most once per sampling window": `runner.RunChecksReturns(errors.New(ctx, "boom"))`; run `check.NewCheckCron(creator, runner, time.Millisecond, log.DefaultSamplerFactory)` in a goroutine with a cancellable context; `time.Sleep(150 * time.Millisecond)`; `cancel()`; `time.Sleep(20 * time.Millisecond)`; then `Expect(strings.Count(buf.String(), "run checks failed")).To(Equal(1))` — `DefaultSamplerFactory`'s 10-second `NewSampleTime` window makes the first iteration the only one that samples, so ~150 iterations yield exactly one warning line (deterministic; the test finishes in ~170ms).
3. It "emits no warning when checks apply successfully": `runner.RunChecksReturns(nil)`; run the cron for `time.Sleep(50 * time.Millisecond)`, cancel, small settle sleep, then `Expect(buf.String()).NotTo(ContainSubstring("run checks failed"))`.

Use `github.com/bborbe/errors` (`errors.New(ctx, ...)`) for the stubbed failure in the test. The suite file `./pkg/check/check_suite_test.go` needs no changes.

### Step 6 — CHANGELOG

Append bullets under the existing `## Unreleased` section in `./CHANGELOG.md` (conventional prefixes), e.g.:

```markdown
- refactor: Gate the checks-cron failure warning behind a github.com/bborbe/log time sampler (at most once per 10 minutes), threaded through the factory from main.go; tests use DefaultSamplerFactory
- refactor: Aggregate the list-lights per-light detail into a single post-loop emission instead of one log call per light
```

### Step 7 — Self-check

Walk spec Desired Behavior 4 and Acceptance Criterion 4 against the change. Re-run the `<verification>` block below and confirm every check passes, then run `make precommit` and `make test` and confirm both exit 0.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git. No `git commit`, `git push`, `git add`, no PR (spec repo CLAUDE.md + `.dark-factory.yaml`).
- Spec Constraints (copied verbatim): never code directly (dark-factory pipeline); repo conventions frozen — `github.com/bborbe/time` not stdlib `time`, `github.com/bborbe/errors` with `errors.Wrap(ctx, err, "...")` never bare `return err`, `github.com/bborbe/collection.Ptr()`, Ginkgo v2 / Gomega, external test packages (`pkg_test`); factory functions stay pure composition — no conditionals, no I/O, no `context.Background()`; no new counterfeiter mock for the sampler (sibling PR 2 owns mocks) — tests use `github.com/bborbe/log`'s `DefaultSamplerFactory`; the controller's default slog level must emit the same lines the `LOGLEVEL=2` deployment emits today (output parity — already established by the previous prompt, keep it); the `/setloglevel` endpoint path stays `/setloglevel/{level}`; `k8s/hue-deploy.yaml` and `k8s/k8s.env` stay untouched; `make precommit` mandatory before commit; existing tests keep passing; `main_test.go` / `cmd/*/main_test.go` `gexec.Build` compile checks must stay green; no scenario is added.
- Do NOT apply a sampler gate to the list-lights listing loop — the sanctioned fix is the single aggregated emission (spec DB 4 + Failure Modes row: a sampler would truncate the CLI's output mid-listing).
- Do NOT change the checks-cron `ctx.Done()` / timer loop structure, the `CreateChecks` error abort, or the error-wrapping in this prompt — only the Warn gating and the constructor/factory/main wiring change.
- Do NOT touch `k8s/`, `go.mod`, `tools.env`, `mocks/`, `pkg/handler/`, or the `/setloglevel` handler from the previous prompt.
- Do NOT renumber the prompt filename — dark-factory assigns numbers on approve.
</constraints>

<verification>
All commands run from the repo root. Zero-match greps use a form with a truthful exit code.

1. `grep -rn 'IsSample\|NewSampleTime\|SamplerFactory' pkg/check/ pkg/factory/ main.go --include='*.go'` — must print ≥1 line (spec AC 4: sampler wired end-to-end).
2. `grep -n 'SamplerFactory' pkg/check/checks-cron.go` — must print ≥1 line (the constructor parameter).
3. `grep -n 'SamplerFactory' pkg/factory/factory.go` — must print ≥1 line (factory threads it).
4. `grep -n 'SamplerFactory\|NewSampleTime' main.go` — must print ≥1 line (composition root passes the 10-minute sampler).
5. `[ -z "$(grep -rn 'slog.Info\|slog.Debug\|slog.Warn' cmd/list-lights/main.go | grep -n 'light state')" ]` — must exit 0 (the per-iteration log call is gone from the listing loop; the two aggregated Info lines remain).
6. `grep -rn 'slog.Info' cmd/ --include='main.go'` — must still print ≥6 lines (spec AC 3 holds).
7. `[ -z "$(grep -r 'glog' --include='*.go' .)" ]` — must still exit 0 (spec AC 1 holds).
8. `[ -z "$(grep -rn 'all checks applied\|sleep for\|next trigger in' pkg/ --include='*.go')" ]` — must still exit 0 (spec AC 5 holds).
9. `make precommit` — must exit 0 (DoD gate per `docs/dod.md`).
10. `make test` — must exit 0.
</verification>
