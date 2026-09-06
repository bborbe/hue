---
status: completed
spec: [001-inject-interfaces-pkg-check]
summary: 'Added SunriseSunsetProvider capability interface and NewSunriseSunsetProvider concrete implementation in pkg/ reproducing the pre-refactor sunrisesunset call (lat 50.1, lon 8.1, UTC offset 0, now.UTC()), with a Ginkgo boundary test proving parity against a direct library call at summer/winter instants and wrapped validation errors; CHANGELOG updated under ## Unreleased'
execution_id: hue-inject-interfaces-pkg-check-exec-005-spec-001-inject-interfaces-pkg-check-sunrise-capability
dark-factory-version: dev
created: "2026-09-06T12:00:00Z"
queued: "2026-09-06T10:24:48Z"
started: "2026-09-06T10:24:50Z"
completed: "2026-09-06T10:26:43Z"
---

# Sunrise/sunset capability interface + concrete implementation in pkg/

<summary>
- A new sunrise/sunset capability interface lands in the shared package, so downstream code can receive sunrise/sunset via constructor injection instead of calling the library directly
- The real implementation wraps the `kelvins/sunrisesunset` library and reproduces the pre-refactor inputs exactly: latitude 50.1, longitude 8.1, UTC offset 0, date from the caller-provided instant converted to UTC
- The capability is pure computation: it has no clock of its own and performs no I/O — the instant arrives as a parameter
- A boundary test proves the capability returns values identical to a direct library call at a summer and a winter instant, and that a library validation error is wrapped
- No counterfeiter mock is generated for the new interface — that is explicitly a sibling PR (spec Non-goal)
- The checks layer is untouched by this prompt; it still calls the library directly until the next prompt
- The repo keeps compiling and `make precommit` stays green
</summary>

<objective>
Create the sunrise/sunset capability interface and its real concrete implementation in `pkg/` so that the checks creator (next prompt) can obtain sunrise and sunset through a constructor-injected capability instead of calling the `sunrisesunset` library from business logic. The concrete implementation must produce exactly the values the pre-refactor direct call produced, because a later parity test and a live log-capture diff enforce byte-for-byte equality.
</objective>

<context>
Read first (the prompt depends on these being read before any edit):

- `./CLAUDE.md` — project conventions: `make precommit` is the mandatory gate, `docs/dod.md` is the validation prompt, `errors.Wrap(ctx, err, "...")` never bare `return err`, interface → constructor → struct → method shape, counterfeiter mocks live in `mocks/`.
- `./docs/dod.md` — DoD: exported types need doc comments; no debug output; Ginkgo v2 / Gomega tests.
- `./.dark-factory.yaml` — `worktree: false`, `pr: false`, `autoRelease: false`; the daemon handles git, the agent MUST NOT commit.
- `./pkg/bridges-provider.go` — the interface-plus-impl pattern to copy: exported interface with doc comment, exported constructor returning the interface, unexported struct, BSD license header.
- `./pkg/time-of-day.go` — pkg-level file shape (license header lines 1-3, `package pkg`).
- `./pkg/check/checks-creator.go` — the exact pre-refactor call this capability replaces (see lines 72-81): `sunrisesunset.Parameters{Latitude: 50.1, Longitude: 8.1, UtcOffset: 0, Date: now.UTC()}` then `p.GetSunriseSunset()` and `errors.Wrap(ctx, err, "get sunrise and sunset failed")`. The capability must reproduce this input exactly, including `now.UTC()`.
- `./pkg/pkg_suite_test.go` — package `pkg_test` Ginkgo suite (already bootstrapped; a new test file needs no suite changes).
- `./pkg/time-of-day_test.go` — Ginkgo test shape in `pkg_test`.
- Coding plugin docs (in-container paths): `/home/node/.claude/plugins/marketplaces/coding/docs/go-composition.md` (pure composition: no I/O, no conditionals, no `context.Background()` in a capability) and `/home/node/.claude/plugins/marketplaces/coding/docs/go-error-wrapping-guide.md` (`errors.Wrap(ctx, err, "...")`, never bare `return err`).

Library APIs verified against the module cache:

- `github.com/kelvins/sunrisesunset` — `type Parameters struct { Latitude float64; Longitude float64; UtcOffset float64; Date time.Time }` and `func (p *Parameters) GetSunriseSunset() (time.Time, time.Time, error)`. The library validates latitude (-90..90), longitude (-180..180), offset (-12..14) and date (1900..2200) and returns a plain error on invalid input.
- `github.com/bborbe/errors` — `errors.Wrap(ctx, err, "...")`.
</context>

<requirements>

### Step 1 — Create `pkg/sunrise-sunset.go`

1. Create the new file `pkg/sunrise-sunset.go` (repo-relative path). `package pkg`. Copy the BSD license header verbatim from `pkg/time-of-day.go` lines 1-3.

2. Imports needed: `"context"`, `"time"`, `"github.com/bborbe/errors"`, `"github.com/kelvins/sunrisesunset"`. Both libraries are already direct dependencies of this module (see `go.mod`), so no dependency change is needed and `go.mod` / `go.sum` must NOT be touched.

3. Define the interface (exact shape):

```go
// SunriseSunsetProvider computes the apparent sunrise and sunset times for the
// given instant at the controller's fixed location coordinates.
type SunriseSunsetProvider interface {
	GetSunriseSunset(ctx context.Context, now time.Time) (time.Time, time.Time, error)
}
```

4. Define the constructor and concrete implementation (exact shape — this reproduces the pre-refactor call from `pkg/check/checks-creator.go` line-for-line, including the `now.UTC()` argument and the error message):

```go
// NewSunriseSunsetProvider returns the real sunrise/sunset capability for the
// controller's fixed location coordinates.
func NewSunriseSunsetProvider() SunriseSunsetProvider {
	return &sunriseSunsetProvider{}
}

type sunriseSunsetProvider struct{}

func (s *sunriseSunsetProvider) GetSunriseSunset(ctx context.Context, now time.Time) (time.Time, time.Time, error) {
	p := sunrisesunset.Parameters{
		Latitude:  50.1,
		Longitude: 8.1,
		UtcOffset: 0,
		Date:      now.UTC(),
	}
	sunrise, sunset, err := p.GetSunriseSunset()
	if err != nil {
		return time.Time{}, time.Time{}, errors.Wrap(ctx, err, "get sunrise and sunset failed")
	}
	return sunrise, sunset, nil
}
```

5. Constraints on this file:
   - Do NOT add a `//counterfeiter:generate` directive — counterfeiter mocks for the new interfaces are explicitly sibling PR 2 (spec Non-goal). `make generate` must produce no new mock files.
   - Do NOT take a clock (`libtime.CurrentDateTimeGetter`) or any other dependency — the instant arrives as the `now` parameter (spec Desired Behavior 1: "the date taken from the injected clock's instant"; the checks creator owns the clock and passes the instant).
   - No log output, no I/O, no `context.Background()` — pure computation (spec Constraints + `docs/dod.md` "no debug output").
   - The error message `"get sunrise and sunset failed"` must match the pre-refactor message byte-for-byte so error parity is preserved.

### Step 2 — Create `pkg/sunrise-sunset_test.go`

6. Create `pkg/sunrise-sunset_test.go` (repo-relative). `package pkg_test`. License header as in `pkg/time-of-day_test.go`. Ginkgo v2 / Gomega.

7. Imports: `"context"`, `"time"`, `"github.com/kelvins/sunrisesunset"`, `. "github.com/onsi/ginkgo/v2"`, `. "github.com/onsi/gomega"`, `"github.com/bborbe/hue/pkg"`.

8. This test is the boundary test for the new capability (per the "test the boundaries the new code crosses" rule): it drives the capability through the real `sunrisesunset` library boundary and compares the result against a direct library call with the exact pre-refactor parameters. Shape:

```go
var _ = Describe("SunriseSunsetProvider", func() {
	var (
		ctx      context.Context
		provider pkg.SunriseSunsetProvider
	)

	BeforeEach(func() {
		ctx = context.Background()
		provider = pkg.NewSunriseSunsetProvider()
	})

	DescribeTable("matches the direct sunrisesunset call with pre-refactor parameters",
		func(now time.Time) {
			expected := sunrisesunset.Parameters{
				Latitude:  50.1,
				Longitude: 8.1,
				UtcOffset: 0,
				Date:      now.UTC(),
			}
			expectedSunrise, expectedSunset, err := expected.GetSunriseSunset()
			Expect(err).NotTo(HaveOccurred())

			sunrise, sunset, err := provider.GetSunriseSunset(ctx, now)
			Expect(err).NotTo(HaveOccurred())
			Expect(sunrise).To(Equal(expectedSunrise))
			Expect(sunset).To(Equal(expectedSunset))
		},
		Entry("summer date", time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)),
		Entry("winter date", time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)),
	)

	It("wraps library validation errors", func() {
		// 1800 is outside the library's supported 1900-2200 date range, so
		// GetSunriseSunset returns an error that the capability must wrap.
		_, _, err := provider.GetSunriseSunset(ctx, time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC))
		Expect(err).To(HaveOccurred())
	})
})
```

9. Do NOT touch any existing file in this prompt. In particular `pkg/check/` stays exactly as-is — removing the direct library call from the checks creator is the next prompt's job.

### Step 3 — Self-check

10. Before finishing, re-run the `<verification>` block below and confirm every check passes; walk each requirement against the change. Confirm the change touched ONLY the two new files (`pkg/sunrise-sunset.go`, `pkg/sunrise-sunset_test.go`) via verification step 10 — git is not usable inside the container for this worktree repo, so the scope check is filesystem-based; if any pre-existing file's content changed, STOP and report.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git. No `git commit`, `git push`, `git add`, no PR.
- Do NOT touch `pkg/check/`, `pkg/factory/`, or `main.go` — the checks layer keeps calling `sunrisesunset` directly until the next prompt in this sequence.
- Do NOT add a counterfeiter `//go:generate` / `//counterfeiter:generate` directive — mocks for the new interfaces are sibling PR 2 (spec Non-goal).
- `github.com/bborbe/errors` wrapping via `errors.Wrap(ctx, err, "...")`; never bare `return err` (spec Constraints).
- Sunrise library injected via interface — frozen design rule (spec Constraints). The capability must produce the same values as the pre-refactor direct call (parity test + live capture enforce this).
- Factory functions stay pure composition: no conditionals, no I/O, no `context.Background()` (spec Constraints) — applies to the new capability too.
- `github.com/bborbe/time` is injected via `libtime.CurrentDateTimeGetter` in constructors and created once in `main.go` — the new capability takes the instant as a parameter instead, it does NOT take a clock.
- The time-zone location stays `Europe/Berlin`; no config surface is added (spec Constraints).
- No new log output in the checks-cron path (spec Constraints) — the new file must not log.
- Existing tests must keep passing; tests may be edited only for the new constructor parameters, and the parity test uses the real concrete capability (counterfeiter mocks are PR 2) (spec Constraints).
- `make precommit` mandatory before finishing (spec Constraints).
- Never number the prompt filename — dark-factory assigns numbers on approve.
</constraints>

<verification>
All commands run from the repo root. Zero-match greps are written in a form whose exit code is truthful (see prompt-writing guide "exit codes lie in two directions").

1. `[ -f pkg/sunrise-sunset.go ]` — file exists. Must exit 0.
2. `grep -n "type SunriseSunsetProvider interface" pkg/sunrise-sunset.go` — must print the interface declaration.
3. `grep -n "func NewSunriseSunsetProvider() SunriseSunsetProvider" pkg/sunrise-sunset.go` — must print the constructor.
4. `grep -n "Latitude:  50.1\|Longitude: 8.1\|UtcOffset: 0\|Date:      now.UTC()" pkg/sunrise-sunset.go` — must print all four pre-refactor parameter lines.
5. `grep -n "errors.Wrap(ctx, err, \"get sunrise and sunset failed\")" pkg/sunrise-sunset.go` — must print the wrapping line.
6. `! grep -q "counterfeiter" pkg/sunrise-sunset.go` — must exit 0 (no mock directive; PR 2 owns mocks).
7. `grep -rn "NewSunriseSunsetProvider" pkg/sunrise-sunset_test.go` — must print at least one test usage.
8. `make precommit` — must exit 0 (DoD gate per `docs/dod.md`).
9. `make test` — must exit 0.
10. `find pkg -maxdepth 1 -type f -name '*.go'` — print the listing and confirm it contains exactly the pre-change files (bridges-provider-cache.go, bridges-provider-fallback.go, bridges-provider-func.go, bridges-provider.go, light.go, pkg_suite_test.go, time-of-day.go, time-of-day_test.go, token.go) plus `sunrise-sunset.go` and `sunrise-sunset_test.go`. Do NOT use `git status` — git is unavailable in the container for this worktree repo. Any file that is not in the pre-change listing means scope was exceeded: STOP and report.
</verification>
