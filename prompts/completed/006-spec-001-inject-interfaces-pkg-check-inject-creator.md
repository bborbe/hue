---
status: completed
spec: [001-inject-interfaces-pkg-check]
summary: Rewired NewCheckCreator to receive the pre-resolved Europe/Berlin *time.Location and pkg.SunriseSunsetProvider via its constructor, removing all direct time.LoadLocation and sunrisesunset library calls from pkg/check business logic; main.go now resolves the location once at the composition root and the factory/creator call sites were updated in the same change
execution_id: hue-inject-interfaces-pkg-check-exec-006-spec-001-inject-interfaces-pkg-check-inject-creator
dark-factory-version: dev
created: "2026-09-06T12:00:00Z"
queued: "2026-09-06T10:24:48Z"
started: "2026-09-06T10:26:44Z"
completed: "2026-09-06T10:29:20Z"
---

# Inject sunrise capability + location into NewCheckCreator, strip pkg/check of direct calls

<summary>
- The checks creator stops calling `time.LoadLocation` and the `sunrisesunset` library — both arrive via constructor parameters instead
- The creator gains a pre-resolved time-zone location and a sunrise/sunset capability through its constructor; its business logic becomes free of package-function calls
- The checks package's non-test code now contains zero references to the sunrise library and zero direct time-zone lookups, closing the code-review findings
- The two V(2) log lines in the schedule path keep their exact wording, so the live pre-/post-merge capture stays diff-comparable
- The factory and the composition root call sites are updated in the same change because Go has no default parameters — the creator's signature change would otherwise leave the repo uncompilable; the location is resolved once at the composition root
- Existing creator tests are updated for the new constructor parameters and keep proving the schedule is unchanged
- The repo compiles and `make precommit` stays green at the end of this prompt
</summary>

<objective>
Rewire the checks creator to consume the sunrise/sunset capability and the pre-resolved `Europe/Berlin` location through its constructor, so `pkg/check` business logic contains zero direct `sunrisesunset` and zero `time.LoadLocation` calls (closing the `no-package-function-calls-in-business-logic` review finding). The schedule the creator renders must be unchanged: same instants, same hours, same V(2) log wording.
</objective>

<context>
Read first (the prompt depends on these being read before any edit):

- `./CLAUDE.md` — project conventions: `make precommit` gate, `errors.Wrap(ctx, err, "...")`, factory functions are pure composition, Ginkgo v2 / Gomega.
- `./docs/dod.md` — DoD.
- `./.dark-factory.yaml` — `worktree: false`, `pr: false`, `autoRelease: false`; daemon handles git, agent MUST NOT commit.
- `./pkg/sunrise-sunset.go` — the capability created by the previous prompt: `type SunriseSunsetProvider interface { GetSunriseSunset(ctx context.Context, now time.Time) (time.Time, time.Time, error) }` and `func NewSunriseSunsetProvider() SunriseSunsetProvider`.
- `./pkg/check/checks-creator.go` — the file being changed. Current state (verified):
  - `func NewCheckCreator(provider pkg.BridgesProvider, summerMode bool, currentDateTimeGetter libtime.CurrentDateTimeGetter) CheckCreator`
  - `type checkCreator struct { provider pkg.BridgesProvider; summerMode bool; location string; currentDateTimeGetter libtime.CurrentDateTimeGetter }`
  - `CreateChecks` starts with `loc, err := time.LoadLocation(c.location)` (the block to remove), uses `loc` throughout, and computes sunrise/sunset via `sunrisesunset.Parameters{Latitude: 50.1, Longitude: 8.1, UtcOffset: 0, Date: now.UTC()}` + `p.GetSunriseSunset()` (the block to replace).
- `./pkg/check/checks-creator_test.go` — the test file with 5 `check.NewCheckCreator(...)` call sites (the "aquarium window" table, the "aquarium light at 21:00 Berlin" table, the "Jana Aqua Light window boundary" table's constructor site, and the "decouples Jana Aqua Light" It block's two sites). Two Describes already load `berlin, err = stdtime.LoadLocation("Europe/Berlin")` in `BeforeEach`; the first Describe ("aquarium window") does not.
- `./pkg/factory/factory.go` — `func CreateCheckController(url string, id string, token pkg.Token, inverval time.Duration, summerMode bool, currentDateTimeGetter libtime.CurrentDateTimeGetter) run.Func`; its only call site is `main.go`. This signature must grow the two new creator parameters so the repo compiles.
- `./main.go` — `application.Run` calls `factory.CreateCheckController(a.Url, a.ID, a.Token, a.Inverval, a.SummerMode, libtime.NewCurrentDateTime())`. This call site must pass the resolved location and the capability.
- Coding plugin docs (in-container paths): `/home/node/.claude/plugins/marketplaces/coding/docs/go-error-wrapping-guide.md` (wrap, never bare `return err`), `/home/node/.claude/plugins/marketplaces/coding/docs/go-composition.md` and `/home/node/.claude/plugins/marketplaces/coding/docs/go-factory-pattern.md` (factories are pure composition; error-prone construction like `time.LoadLocation` belongs in `main.go Run`, not in a factory — see go-factory-pattern section on "Errors are runtime concerns"), `/home/node/.claude/plugins/marketplaces/coding/docs/go-time-injection.md` (constructor accepts `libtime.CurrentDateTimeGetter`).

Library APIs verified against the module cache:

- `github.com/kelvins/sunrisesunset` — must be removed from this file's imports (the capability in `pkg/` owns it from now on).
- `github.com/bborbe/errors` — `errors.Wrap(ctx, err, "...")`.
- `github.com/bborbe/time` v1.27.11 — `libtime.CurrentDateTimeGetter` with `Now() libtime.DateTime`; `libtime.DateTime.Time()` converts to `stdtime.Time` (already used by the creator).
</context>

<requirements>

### Step 1 — Change `NewCheckCreator` and the `checkCreator` struct (`pkg/check/checks-creator.go`)

1. Change the constructor signature from the current 3-parameter form to this 5-parameter form (new parameters appended last; types verified):

```go
func NewCheckCreator(
	provider pkg.BridgesProvider,
	summerMode bool,
	currentDateTimeGetter libtime.CurrentDateTimeGetter,
	location *time.Location,
	sunriseSunsetProvider pkg.SunriseSunsetProvider,
) CheckCreator {
	return &checkCreator{
		provider:              provider,
		summerMode:            summerMode,
		location:              location,
		currentDateTimeGetter: currentDateTimeGetter,
		sunriseSunsetProvider: sunriseSunsetProvider,
	}
}
```

2. Change the struct: `location` becomes `*time.Location` (was `string`), and add the capability field:

```go
type checkCreator struct {
	provider              pkg.BridgesProvider
	summerMode            bool
	location              *time.Location
	currentDateTimeGetter libtime.CurrentDateTimeGetter
	sunriseSunsetProvider pkg.SunriseSunsetProvider
}
```

3. In `CreateChecks`, delete the opening block `loc, err := time.LoadLocation(c.location)` with its error wrap (`return nil, errors.Wrap(ctx, err, "load location failed")`) — the location now arrives pre-resolved. Every later use of the local `loc` variable becomes `c.location` (the `pkg.TimeOfDay{... Location: ...}` literals, the glog lines, and the `.In(loc)` / `.String()` calls).

4. In `CreateChecks`, replace the direct library call block

```go
p := sunrisesunset.Parameters{
	Latitude:  50.1,
	Longitude: 8.1,
	UtcOffset: 0,
	Date:      now.UTC(),
}
sunrise, sunset, err := p.GetSunriseSunset()
if err != nil {
	return nil, errors.Wrap(ctx, err, "get sunrise and sunset failed")
}
```

with the capability call — identical error message:

```go
sunrise, sunset, err := c.sunriseSunsetProvider.GetSunriseSunset(ctx, now)
if err != nil {
	return nil, errors.Wrap(ctx, err, "get sunrise and sunset failed")
}
```

5. Remove the import `"github.com/kelvins/sunrisesunset"` from this file. Keep `"time"` (still used for `5*time.Minute`, `25*time.Minute`, `time.RFC3339`), `libtime "github.com/bborbe/time"`, `"github.com/bborbe/errors"`, `"github.com/golang/glog"`, and `"github.com/bborbe/hue/pkg"`.

6. Log parity (spec Desired Behavior 6 + Constraints "no new log lines"): the two `glog.V(2)` lines in `CreateChecks` must keep their exact format strings and wording — only the receiver changes from the removed `loc` local to `c.location`:

   - `glog.V(2).Infof("current time %s in %s", now.In(c.location).Format(time.RFC3339), c.location.String())`
   - `glog.V(2).Infof("now %s sunrise %s sunset %s", now.In(c.location).Format("15:04:05"), sunrise.In(c.location).Format("15:04:05"), sunset.In(c.location).Format("15:04:05"))`

   Do NOT add, remove, reword, or reorder any log line in `pkg/check/`.

7. Do NOT change the `CheckCreator` interface in this file. Do NOT change anything else in `CreateChecks` — the hour constants, the summerMode branching, and the 9 switch constructions stay byte-identical.

### Step 2 — Update `pkg/factory/factory.go` (compilation ripple, not new wiring)

8. The creator signature change cannot compile without its production callers — Go has no default parameters and `CreateCheckController` is the only production caller. Extend it (new parameters appended last, types verified):

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
		check.NewChecksRunner(),
		inverval,
	)
}
```

9. The factory stays pure composition: no conditionals, no I/O, no `context.Background()`, no `time.LoadLocation` here. `check.NewChecksRunner()` stays untouched in this prompt — it gains the clock in the next prompt.

### Step 3 — Update `main.go` (composition root resolves the location once)

10. In `application.Run`, before `service.Run`, resolve the location once and own the load error (spec Desired Behavior 2: "The location is resolved exactly once at the composition root (main.go), which owns the load error"). Insert:

```go
location, err := time.LoadLocation("Europe/Berlin")
if err != nil {
	return errors.Wrap(ctx, err, "load location failed")
}
```

This requires adding the import `"github.com/bborbe/errors"` to `main.go` (stdlib `"time"` and `"github.com/bborbe/hue/pkg"` are already imported).

11. Extend the `factory.CreateCheckController(...)` call in `Run` to pass the new parameters as the last two arguments:

```go
factory.CreateCheckController(
	a.Url,
	a.ID,
	a.Token,
	a.Inverval,
	a.SummerMode,
	libtime.NewCurrentDateTime(),
	location,
	pkg.NewSunriseSunsetProvider(),
)
```

The clock stays constructed exactly once via `libtime.NewCurrentDateTime()` and the capability exactly once via `pkg.NewSunriseSunsetProvider()` (spec Desired Behavior 4).

### Step 4 — Update `pkg/check/checks-creator_test.go`

12. Update all five `check.NewCheckCreator(...)` call sites to the new signature — each gains `berlin` (the `*stdtime.Location` for `Europe/Berlin`, already loaded in two of the Describes) and `pkg.NewSunriseSunsetProvider()` as the last two arguments. Call sites (verify against the file, don't trust line numbers):
   - the "aquarium window" `DescribeTable` body,
   - the "aquarium light at 21:00 Berlin" `DescribeTable` body,
   - the "Jana Aqua Light window boundary" `DescribeTable` body,
   - the two constructions in the "decouples Jana Aqua Light from the shared aquarium window" `It` block.

13. The first `Describe` ("aquarium window", which currently constructs with `libtime.NewCurrentDateTime()` and has no `berlin` in scope) needs a `berlin` in scope: declare `berlin *stdtime.Location` in that Describe's `var` block and `var err error` inside the `BeforeEach`, then add `berlin, err = stdtime.LoadLocation("Europe/Berlin")` and `Expect(err).NotTo(HaveOccurred())`, mirroring the `BeforeEach` blocks of the other two Describes (existing test code is allowed to keep resolving `Europe/Berlin` via `LoadLocation` — the spec's negative grep excludes `_test.go`).

14. Tests may be edited ONLY for the new constructor parameters (spec Constraints). Do NOT change any assertion, table entry, or fixture. Do NOT convert the `fixedClock` helper. The tests continue to exercise `CreateChecks` end-to-end with the real concrete capability (mocks are PR 2), which is the boundary test for the new `pkg.SunriseSunsetProvider` and `*time.Location` constructor parameters.

### Step 5 — Self-check

15. Before finishing, re-run the `<verification>` block below and confirm every check passes; walk each acceptance criterion against the change. Run `make precommit` and confirm it exits 0.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git. No `git commit`, `git push`, `git add`, no PR.
- Do NOT touch `pkg/check/checks-runner.go` — it gains the clock in the next prompt (spec DB 3); `check.NewChecksRunner()` in the factory stays argument-less in this prompt.
- Do NOT touch `pkg/sunrise-sunset.go` or `pkg/sunrise-sunset_test.go` from the previous prompt.
- Do NOT add a counterfeiter directive — mocks for the new interfaces are sibling PR 2 (spec Non-goal).
- `github.com/bborbe/errors` wrapping via `errors.Wrap(ctx, err, "...")`; never bare `return err` (spec Constraints). The removed `LoadLocation` error wrap moves to `main.go` with the same message.
- Sunrise library injected via interface (frozen design rule, spec Constraints). After this prompt `pkg/check` non-test code contains zero references to `sunrisesunset` and zero `time.LoadLocation` calls.
- `BridgesProvider` stays composed (`func` → `cache` → `fallback`); never collapsed to a single struct (spec Constraints). `CreateBridgesProvider` is not touched by this prompt.
- The time-zone location stays `Europe/Berlin`; no config surface is added (spec Constraints).
- No new log lines in the checks-cron path (spec Constraints) — the two V(2) lines keep their exact wording; the log line contents and order must not change.
- Existing tests keep passing; tests may be edited only for the new constructor parameters (spec Constraints). The parity test (next-next prompt) uses the real concrete capability (mocks are PR 2).
- `make precommit` mandatory before finishing (spec Constraints).
- Never number the prompt filename — dark-factory assigns numbers on approve.
</constraints>

<verification>
All commands run from the repo root. Zero-match greps use a form with a truthful exit code (see prompt-writing guide "exit codes lie in two directions").

1. `[ -z "$(grep -rn 'time.Now()' pkg/check/)" ]` — must exit 0 (spec AC 1).
2. `[ -z "$(grep -rn 'sunrisesunset' pkg/check/ --include='*.go' | grep -v '_test.go')" ]` — must exit 0 (spec AC 2).
3. `[ -z "$(grep -rn 'LoadLocation' pkg/check/ --include='*.go' | grep -v '_test.go')" ]` — must exit 0 (spec AC 3).
4. `grep -n "sunriseSunsetProvider pkg.SunriseSunsetProvider" pkg/check/checks-creator.go` — must print the struct field (or the constructor param line; at least one match).
5. `grep -n "location \*time.Location" pkg/check/checks-creator.go` — must print the struct field.
6. `grep -n "c.sunriseSunsetProvider.GetSunriseSunset(ctx, now)" pkg/check/checks-creator.go` — must print the capability call.
7. `[ -z "$(grep -n 'sunrisesunset' pkg/check/checks-creator.go)" ]` — must exit 0 (import removed from the creator).
8. `grep -n "time.LoadLocation(\"Europe/Berlin\")" main.go` — must print the composition-root resolution.
9. `grep -rn "pkg.NewSunriseSunsetProvider()" main.go` — must print exactly one line (capability constructed once at the composition root).
10. `grep -n "libtime.NewCurrentDateTime()" main.go` — must print exactly one line (clock constructed once).
11. `make precommit` — must exit 0 (DoD gate per `docs/dod.md`).
12. `make test` — must exit 0.
</verification>
