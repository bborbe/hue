---
status: approved
spec: [001-inject-interfaces-pkg-check]
created: "2026-09-06T12:00:00Z"
queued: "2026-09-06T10:24:48Z"
---

# Parity test: injected path vs pre-refactor reference computation

<summary>
- A new parity test in `pkg/check` renders the full 9-check light-flip schedule through two independent paths and asserts byte-for-byte equality
- The reference arm reproduces the pre-refactor computation: `Europe/Berlin` via `time.LoadLocation` and sunrise/sunset via the direct `sunrisesunset` library call with the exact pre-refactor parameters (lat 50.1, lon 8.1, offset 0)
- The injected arm drives the real `NewCheckCreator` with the injected location, clock, and real concrete capability — exactly as the factory wires it
- Both arms are swept at two fixed instants (a summer date and a winter date) × both `summerMode` values = four table entries
- The rendered output (sunrise/sunset line plus each of the 9 checks' names, which encode their on/off state at the instant) is compared as strings
- This is the in-repo proof that the refactor changed nothing: the schedule is byte-identical before and after injection
- The reference arm's direct library call is the single allowed `sunrisesunset` occurrence in `pkg/check`, and it lives in a test file exactly as the spec's AC 2 permits
- The repo compiles and `make precommit` stays green
</summary>

<objective>
Prove the injected path produces a byte-for-byte identical schedule to the pre-refactor computation, in-repo, at fixed instants. The parity test renders the full 9-check schedule (each check's name and on/off state plus sunrise and sunset times) through the direct pre-refactor reference computation and through the injected constructor path, at ≥2 fixed instants (one summer-date, one winter-date) × both `summerMode` values, and asserts string equality.
</objective>

<context>
Read first (the prompt depends on these being read before any edit):

- `./CLAUDE.md` — project conventions: `make precommit` gate, Ginkgo v2 / Gomega, `errors.Wrap(ctx, err, "...")`.
- `./docs/dod.md` — DoD.
- `./.dark-factory.yaml` — `worktree: false`, `pr: false`, `autoRelease: false`; daemon handles git, agent MUST NOT commit.
- `./pkg/check/checks-creator_test.go` — the test file that already defines `type fixedClock struct{ t stdtime.Time }` with `func (f fixedClock) Now() libtime.DateTime` and loads `berlin` via `stdtime.LoadLocation("Europe/Berlin")` in `BeforeEach`. The new parity test lives in the same external package `check_test`, so it reuses `fixedClock` and the `berlin` BeforeEach pattern.
- `./pkg/check/checks-creator.go` — the refactored creator (all constructor params injected). The 9-check construction order and the hour constants in `CreateChecks` are the contract the reference arm must replicate exactly: Artemia 8-23; aquarium cluster `aquariumLightOnHour`/`aquariumLightOffhour` (summerMode: 20/+3, else 10/+10); CO2 = on-2 / off-2; Skimmer alternate 5min/25min; Jana Aqua Light 12:30-22:30; Jana Aqua Skimmer alternate 5min/25min.
- `./pkg/check/checks-runner.go` / `./pkg/factory/factory.go` / `./main.go` — wiring reference: the injected path in the test mirrors `CreateCheckController` (creator built with provider, summerMode, clock, berlin, `pkg.NewSunriseSunsetProvider()`).
- `./pkg/check/between-time-switch.go`, `./pkg/check/alternate-switch.go`, `./pkg/check/light-is-on.go`, `./pkg/check/light-is-off.go`, `./pkg/check/func.go` — the public switch constructors the reference arm uses: `NewBetweenTimeSwitch(now time.Time, from, until pkg.TimeOfDay, main, fallback Check) Check`, `NewAlternateSwitch(now time.Time, mainDuration, secondDuration time.Duration, main, fallback Check) Check`, `NewLightIsOn(bridge *huego.Bridge, lightName pkg.LightName) Check`, `NewLightIsOff(bridge *huego.Bridge, lightName pkg.LightName) Check`. Note `NewSwitch` evaluates its predicate at construction, so each returned `Check`'s `Name()` already encodes the on/off state at the instant.
- `./pkg/sunrise-sunset.go` — the injected capability: `pkg.NewSunriseSunsetProvider()` / `GetSunriseSunset(ctx, now)`.
- `./pkg/time-of-day.go` — `type TimeOfDay struct { Hour, Minute, Second int; Location *time.Location }`.
- Coding plugin docs (in-container paths): `/home/node/.claude/plugins/marketplaces/coding/docs/go-testing-guide.md` (Ginkgo tables), `/home/node/.claude/plugins/marketplaces/coding/docs/go-composition.md`.

Library API verified against the module cache: `github.com/kelvins/sunrisesunset` — `type Parameters struct { Latitude float64; Longitude float64; UtcOffset float64; Date time.Time }` and `func (p *Parameters) GetSunriseSunset() (time.Time, time.Time, error)`.

IMPORTANT: the pre-refactor creator source no longer exists in the repo when this prompt runs — prompts 2 and 3 already rewired it. The reference arm below is therefore specified IN FULL in this prompt: copy it verbatim, do not re-derive it from `checks-creator.go`. This is a frozen snapshot of the pre-refactor behavior, which is exactly what a parity test needs.
</context>

<requirements>

### Step 1 — Create `pkg/check/checks-parity_test.go`

1. Create the new file `pkg/check/checks-parity_test.go` (repo-relative path). `package check_test` (same external package as `checks-creator_test.go`, so `fixedClock` from that file is in scope and reusable). BSD license header as in `checks-creator_test.go` lines 1-3.

2. Imports:

```go
import (
	"context"
	"fmt"
	"strings"
	stdtime "time"

	"github.com/amimof/huego"
	"github.com/kelvins/sunrisesunset"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/hue/pkg"
	"github.com/bborbe/hue/pkg/check"
)
```

(`libtime` is NOT needed — `fixedClock` already lives in `checks-creator_test.go`.)

### Step 2 — The shared renderer

3. Add the renderer that both arms use, so the compared strings share one format:

```go
// renderSchedule renders the schedule the checks-cron path logs: the sunrise
// and sunset times in the given location plus each check's name (which encodes
// its on/off state at the instant).
func renderSchedule(loc *stdtime.Location, sunrise, sunset stdtime.Time, checks check.Checks) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "sunrise %s sunset %s\n", sunrise.In(loc).Format("15:04:05"), sunset.In(loc).Format("15:04:05"))
	for _, c := range checks {
		fmt.Fprintf(&sb, "%s\n", c.Name())
	}
	return sb.String()
}
```

### Step 3 — The reference arm (frozen pre-refactor computation)

4. Add the reference computation. Copy it VERBATIM — it reproduces the pre-refactor `CreateChecks`: `Europe/Berlin` resolved via `time.LoadLocation`, sunrise/sunset via the direct `sunrisesunset` call with the exact pre-refactor parameters (latitude 50.1, longitude 8.1, UTC offset 0, `Date: now.UTC()`), and the same 9 checks in the same order with the same hours. `panic(err)` is acceptable here — a failure in the reference arm means the test cannot establish a baseline and must fail loudly.

```go
// referenceSchedule reproduces the pre-refactor CreateChecks computation:
// Europe/Berlin via time.LoadLocation and sunrise/sunset via the direct
// sunrisesunset library call with the exact pre-refactor parameters
// (latitude 50.1, longitude 8.1, UTC offset 0, date from the instant in UTC).
func referenceSchedule(now stdtime.Time, summerMode bool) string {
	loc, err := stdtime.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}
	p := sunrisesunset.Parameters{
		Latitude:  50.1,
		Longitude: 8.1,
		UtcOffset: 0,
		Date:      now.UTC(),
	}
	sunrise, sunset, err := p.GetSunriseSunset()
	if err != nil {
		panic(err)
	}

	var aquariumLightOnHour int
	var aquariumLightOffhour int
	if summerMode {
		aquariumLightOnHour = 20
		aquariumLightOffhour = aquariumLightOnHour + 3
	} else {
		aquariumLightOnHour = 10
		aquariumLightOffhour = aquariumLightOnHour + 10
	}
	co2OnHour := aquariumLightOnHour - 2
	co2OffHour := aquariumLightOffhour - 2

	checks := check.Checks{
		check.NewBetweenTimeSwitch(now, pkg.TimeOfDay{Hour: 8, Location: loc}, pkg.TimeOfDay{Hour: 23, Location: loc}, check.NewLightIsOn(nil, "Artemia Licht"), check.NewLightIsOff(nil, "Artemia Licht")),
		check.NewBetweenTimeSwitch(now, pkg.TimeOfDay{Hour: aquariumLightOnHour, Location: loc}, pkg.TimeOfDay{Hour: aquariumLightOffhour, Location: loc}, check.NewLightIsOn(nil, "Aquarium Licht"), check.NewLightIsOff(nil, "Aquarium Licht")),
		check.NewBetweenTimeSwitch(now, pkg.TimeOfDay{Hour: aquariumLightOnHour, Location: loc}, pkg.TimeOfDay{Hour: aquariumLightOffhour, Location: loc}, check.NewLightIsOn(nil, "Aquarium Rack"), check.NewLightIsOff(nil, "Aquarium Rack")),
		check.NewBetweenTimeSwitch(now, pkg.TimeOfDay{Hour: co2OnHour, Location: loc}, pkg.TimeOfDay{Hour: co2OffHour, Location: loc}, check.NewLightIsOn(nil, "Aquarium CO2"), check.NewLightIsOff(nil, "Aquarium CO2")),
		check.NewBetweenTimeSwitch(now, pkg.TimeOfDay{Hour: aquariumLightOnHour, Location: loc}, pkg.TimeOfDay{Hour: aquariumLightOffhour, Location: loc}, check.NewLightIsOn(nil, "Garnelen Licht 1"), check.NewLightIsOff(nil, "Garnelen Licht 1")),
		check.NewBetweenTimeSwitch(now, pkg.TimeOfDay{Hour: aquariumLightOnHour, Location: loc}, pkg.TimeOfDay{Hour: aquariumLightOffhour, Location: loc}, check.NewLightIsOn(nil, "Garnelen Licht 2"), check.NewLightIsOff(nil, "Garnelen Licht 2")),
		check.NewAlternateSwitch(now, 5*stdtime.Minute, 25*stdtime.Minute, check.NewLightIsOn(nil, "Aquarium Skimmer"), check.NewLightIsOff(nil, "Aquarium Skimmer")),
		check.NewBetweenTimeSwitch(now, pkg.TimeOfDay{Hour: 12, Minute: 30, Location: loc}, pkg.TimeOfDay{Hour: 22, Minute: 30, Location: loc}, check.NewLightIsOn(nil, "Jana Aqua Light"), check.NewLightIsOff(nil, "Jana Aqua Light")),
		check.NewAlternateSwitch(now, 5*stdtime.Minute, 25*stdtime.Minute, check.NewLightIsOn(nil, "Jana Aqua Skimmer"), check.NewLightIsOff(nil, "Jana Aqua Skimmer")),
	}
	return renderSchedule(loc, sunrise, sunset, checks)
}
```

5. Notes on the reference arm:
   - Passing `nil` as the bridge to `NewLightIsOn`/`NewLightIsOff` is safe: the rendered output only calls `Name()`, which never dereferences the bridge (verify in `light-is-on.go` / `light-is-off.go` / `func.go`).
   - The `Date: now.UTC()` conversion must be reproduced exactly — it matches the capability's internal conversion and the pre-refactor input.

### Step 4 — The injected arm and the parity assertion

6. Add the test body. The injected arm mirrors the factory wiring: a `pkg.BridgesProviderFunc` returning `[]*huego.Bridge{nil}`, the shared `berlin` location, a `fixedClock` pinned to the instant, and one real `pkg.NewSunriseSunsetProvider()` instance that is both passed to the creator and queried for the injected arm's sunrise/sunset. Load `berlin` in `BeforeEach` exactly like `checks-creator_test.go` does:

```go
var _ = Describe("Parity: injected path vs pre-refactor reference", func() {
	var (
		ctx    context.Context
		berlin *stdtime.Location
	)

	BeforeEach(func() {
		var err error
		ctx = context.Background()
		berlin, err = stdtime.LoadLocation("Europe/Berlin")
		Expect(err).NotTo(HaveOccurred())
	})

	DescribeTable("renders the same schedule through the injected path and the reference computation",
		func(instant stdtime.Time, summerMode bool) {
			provider := pkg.BridgesProviderFunc(func(_ context.Context) ([]*huego.Bridge, error) {
				return []*huego.Bridge{nil}, nil
			})

			sunriseSunsetProvider := pkg.NewSunriseSunsetProvider()
			clock := fixedClock{t: instant}
			creator := check.NewCheckCreator(provider, summerMode, clock, berlin, sunriseSunsetProvider)

			checks, err := creator.CreateChecks(ctx)
			Expect(err).NotTo(HaveOccurred())
			sunrise, sunset, err := sunriseSunsetProvider.GetSunriseSunset(ctx, instant)
			Expect(err).NotTo(HaveOccurred())

			injected := renderSchedule(berlin, sunrise, sunset, checks)
			reference := referenceSchedule(instant, summerMode)

			Expect(injected).To(Equal(reference))
		},
		// The instant is pinned in stdtime.UTC, NOT berlin: Entry arguments
		// are evaluated at spec-tree construction, before BeforeEach runs, so
		// berlin is still nil there and stdtime.Date panics on a nil location
		// ("time: missing Location in call to Date"). The instant's location
		// does not affect the parity assertion — both arms consume the
		// identical instant value, and the suite sets time.Local = UTC.
		Entry("summer date, summer mode", stdtime.Date(2026, 6, 15, 12, 0, 0, 0, stdtime.UTC), true),
		Entry("summer date, winter mode", stdtime.Date(2026, 6, 15, 12, 0, 0, 0, stdtime.UTC), false),
		Entry("winter date, summer mode", stdtime.Date(2026, 1, 15, 12, 0, 0, 0, stdtime.UTC), true),
		Entry("winter date, winter mode", stdtime.Date(2026, 1, 15, 12, 0, 0, 0, stdtime.UTC), false),
	)
})
```

7. The four entries satisfy the spec's sweep requirement: ≥2 fixed instants (one summer-date, one winter-date) × both `summerMode` values.

### Step 5 — Self-check

8. Before finishing, re-run the `<verification>` block below and confirm every check passes; walk each requirement against the change. Run `make precommit` and confirm it exits 0.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git. No `git commit`, `git push`, `git add`, no PR.
- This prompt adds ONLY `pkg/check/checks-parity_test.go`. Do NOT modify any existing file — the prior prompts left everything wired and green.
- The parity test uses the real concrete capability (`pkg.NewSunriseSunsetProvider()`) — counterfeiter mocks are sibling PR 2 (spec Constraints).
- The direct `sunrisesunset` reference call is allowed ONLY inside a `_test.go` file (spec AC 2: "the parity test's reference computation is the single allowed occurrence and lives in a test file").
- No new log output in the checks-cron path (spec Constraints) — the test must not log.
- `github.com/bborbe/errors` wrapping via `errors.Wrap(ctx, err, "...")` applies to production code; the test reference arm uses `panic(err)` on baseline failure.
- Sunrise library injected via interface (frozen design rule, spec Constraints) — production code keeps zero direct references; only the test reference arm calls it.
- The time-zone location stays `Europe/Berlin`; no config surface is added (spec Constraints).
- Existing tests keep passing (spec Constraints).
- `make precommit` mandatory before finishing (spec Constraints).
- Never number the prompt filename — dark-factory assigns numbers on approve.
</constraints>

<verification>
All commands run from the repo root.

1. `[ -f pkg/check/checks-parity_test.go ]` — file exists. Must exit 0.
2. `grep -rn "sunrisesunset" pkg/check/ --include='*_test.go'` — must print ≥1 line (spec AC 5 evidence: the parity test's independent reference computation).
3. `grep -n "referenceSchedule" pkg/check/checks-parity_test.go` — must print the reference-arm function.
4. `grep -n "GetSunriseSunset" pkg/check/checks-parity_test.go` — must print ≥1 direct `sunrisesunset.Parameters{...}.GetSunriseSunset()` reference call.
5. `grep -n "Latitude:  50.1\|Longitude: 8.1\|UtcOffset: 0" pkg/check/checks-parity_test.go` — must print the three fixed-coordinate lines of the reference computation.
6. `grep -n "stdtime.UTC" pkg/check/checks-parity_test.go` — must print the four Entry instants pinned in UTC (never a nil location passed to `stdtime.Date`).
7. `[ -z "$(grep -rn 'sunrisesunset' pkg/check/ --include='*.go' | grep -v '_test.go')" ]` — must exit 0 (spec AC 2 still holds; production code untouched).
8. `[ -z "$(grep -rn 'time.Now()' pkg/check/)" ]` — must exit 0 (spec AC 1).
9. `[ -z "$(grep -rn 'LoadLocation' pkg/check/ --include='*.go' | grep -v '_test.go')" ]` — must exit 0 (spec AC 3).
10. `make precommit` — must exit 0 (DoD gate per `docs/dod.md`).
11. `make test` — must exit 0.
</verification>
