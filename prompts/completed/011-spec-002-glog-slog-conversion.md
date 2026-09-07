---
status: completed
spec: [002-migrate-glog-slog]
summary: Converted all 10 hue Go files from glog to stdlib log/slog (20 call-sites) with explicit intent-based levels, deleted the three unconditional per-cycle heartbeat lines, made bridge-discovery lines log count/ID/Host only, moved glog to an indirect go.mod requirement, and added the CHANGELOG Unreleased entry
execution_id: hue-migrate-glog-slog-exec-011-spec-002-glog-slog-conversion
dark-factory-version: dev
created: "2026-09-06T20:55:00Z"
queued: "2026-09-06T19:18:14Z"
started: "2026-09-06T19:18:16Z"
completed: "2026-09-06T19:20:50Z"
---

# Repo-wide glog → slog conversion (level classification, empty-heartbeat removal, sensitive-data-safe bridges-provider)

<summary>
- Every one of the 10 Go files that imports `github.com/golang/glog` switches to stdlib `log/slog`; all 20 call-sites become package-level `slog.Info` / `slog.Debug` / `slog.Warn` calls with key-value attribute pairs — never `fmt.Sprintf`-interpolated message strings
- Each converted line gets an explicit intent-based level per the spec's mapping table: operator-facing output at `Info`, per-check / sunrise-sunset / bridge-discovery state at `Debug`, failures at `Warn`
- The three unconditional per-cycle heartbeat lines are deleted (checks-cron "all checks applied" and "sleep for", time-of-day "next trigger in") — no line fires from a loop merely because a cycle ran
- The bridges-provider discovery lines log only the bridge count and per-bridge `ID` + `Host` — the `%+v` whole-list dump and the `User` API key never appear in a log call
- `go mod tidy` drops `github.com/golang/glog` from the direct `require` block (it remains only as an indirect requirement of `bborbe/log` / `bborbe/service`)
- The checks-cron failure line and the list-lights per-light line are converted in place here but intentionally NOT yet sampler-gated/aggregated — that is the next prompt in this spec's chain, which depends on this prompt landing first
- The repo compiles and `make precommit` + `make test` stay green
</summary>

<objective>
Replace the deprecated glog logger with stdlib `log/slog` across all of hue's own Go code, assigning each converted line an explicit intent-based level, deleting the three empty per-cycle heartbeat lines, and making the bridge-discovery lines safe for logs (count / ID / Host only). This is the bulk mechanical sweep of spec 002: it must land first so the repo compiles on slog alone and the next prompts (logger setup, sampler wiring) can build on it.
</objective>

<context>
Read first (the prompt depends on these being read before any edit):

- `./CLAUDE.md` — project conventions: `make precommit` is the mandatory gate, `github.com/bborbe/errors` with `errors.Wrap(ctx, err, "...")` never bare `return err`, factory functions are pure composition, `github.com/bborbe/time` injected (stdlib `time` stays only for `time.Duration` / `time.Location` in existing signatures).
- `./docs/dod.md` — DoD: exported symbols need doc comments, no debug output (use structured logging), CHANGELOG entry under `## Unreleased`.
- `./.dark-factory.yaml` — `worktree: false`, `pr: false`, `autoRelease: false`; the daemon handles git, the agent MUST NOT commit.
- `./specs/in-progress/002-migrate-glog-slog.md` — the authoritative spec: Desired Behavior 1 (import swap + structured key-values, no `fmt.Sprintf`), 2 (level classification table — every row is copied into the requirements below), 3 (empty-heartbeat removal), 6 (bridge-discovery lines log count / ID / Host only), and the Failure Modes / Security rows the conversion must honor.
- `./main.go`, `./cmd/list-lights/main.go`, `./cmd/turnon-light/main.go`, `./cmd/turnoff-light/main.go`, `./pkg/bridges-provider.go`, `./pkg/trigger/time-of-day.go`, `./pkg/check/between-time-switch.go`, `./pkg/check/checks-runner.go`, `./pkg/check/checks-cron.go`, `./pkg/check/checks-creator.go` — the 10 files being changed; each current glog call-site is quoted verbatim in the requirements below.
- `./go.mod` — `github.com/golang/glog v1.2.5` is a direct requirement (first `require (...)` block); after the sweep + `go mod tidy` it must move to the indirect block only.
- `./CHANGELOG.md` — currently ends at `## v0.4.2` with no `## Unreleased` section; add the section above `## v0.4.2` (see the changelog-guide reference below).
- Coding plugin docs (in-container paths): `/home/node/.claude/plugins/marketplaces/coding/docs/go-logging-guide.md` (slog key-value style, never `slog.Info(fmt.Sprintf(...))`; `no-sensitive-data-in-logs`; `skip-empty-v2-heartbeats`; `lowercase-log-messages`), `/home/node/.claude/plugins/marketplaces/coding/docs/go-glog-guide.md` (why V(2) content is debug-shaped and V0 content is operator Info), `/home/node/.claude/plugins/marketplaces/coding/docs/go-error-wrapping-guide.md`, `/home/node/.claude/plugins/marketplaces/coding/docs/changelog-guide.md` (`## Unreleased` goes directly above the highest `## vX.Y.Z`; bullets use conventional prefixes `feat:` / `fix:` / `refactor:` / `test:` / `chore:`), `/home/node/.claude/plugins/marketplaces/coding/docs/go-precommit.md`.

Verified facts the requirements rely on (read from source):

- `pkg.TimeOfDay` implements `func (t TimeOfDay) String() string` returning `"HH:MM:SS"` — use `.String()` explicitly when logging `from` / `until`.
- The checks-cron loop keeps its existing `select` on `ctx.Done()` plus `time.NewTimer(interval)` structure; only the log lines change in this prompt.
- The `huego.Bridge` fields used below are `ID` and `Host` (strings); `discover.User` is the bridge API key and must never be logged (spec DB 6 / Security).
- `make precommit` = ensure + format + generate + test + check + addlicense (Makefile).
</context>

<requirements>

### Step 1 — Convert all 10 files, one call-site at a time

For each file below: remove the `"github.com/golang/glog"` import, add `"log/slog"` to the stdlib import group (alphabetical, first import block — unless a step below says otherwise), and apply the exact old → new conversion. Messages are lowercase, plain, and never contain interpolated values — values become attributes. Never build a message with `fmt.Sprintf` inside a slog call (spec DB 1).

**1. `./main.go`** (line 90)

| Old | New |
|---|---|
| `glog.V(2).Infof("starting http server listen on %s", a.Listen)` | `slog.Info("starting http server", "listen", a.Listen)` |

**2. `./cmd/list-lights/main.go`** (lines 50, 52)

| Old | New |
|---|---|
| `glog.Infof("found %d lights", len(lights))` | `slog.Info("found lights", "count", len(lights))` |
| `glog.Infof("'%s' on: %v", light.Name, light.IsOn())` | `slog.Info("light state", "name", light.Name, "on", light.IsOn())` |

The per-light `slog.Info("light state", ...)` line stays INSIDE the `for _, light := range lights` loop in this prompt — it is converted in place but NOT yet aggregated (spec decomposition: list-lights aggregation is the last prompt's job, which depends on this prompt). Do not restructure the loop here.

**3. `./cmd/turnon-light/main.go`** (lines 47, 53)

| Old | New |
|---|---|
| `glog.V(2).Info("light already on")` | `slog.Info("light already on")` |
| `glog.Infof("light turned on")` | `slog.Info("light turned on")` |

**4. `./cmd/turnoff-light/main.go`** (lines 47, 53)

| Old | New |
|---|---|
| `glog.V(2).Info("light already off")` | `slog.Info("light already off")` |
| `glog.Infof("light turned off")` | `slog.Info("light turned off")` |

**5. `./pkg/bridges-provider.go`** (lines 27, 39)

| Old | New |
|---|---|
| `glog.V(2).Infof("list %+v", list)` | `slog.Debug("discovered bridges", "count", len(list))` |
| `glog.V(2).Infof("found: %s %s %s", discover.ID, discover.Host, discover.User)` | `slog.Debug("found bridge", "id", discover.ID, "host", discover.Host)` |

The `%+v` whole-discovery dump is gone (count only). The `User` field is NOT logged — it must not appear in any log call in this file. The struct construction `User: token.String()` (line 43) is a value assignment, not a log, and stays untouched. The `found:` Debug line stays guarded by the `discover.ID != id` match (it is adjudicated exempt from `skip-empty-v2-heartbeats` per spec DB 3) — keep the guard exactly as-is.

**6. `./pkg/trigger/time-of-day.go`** (line 20)

- DELETE the line `glog.V(2).Infof("next trigger in %v", duration)` entirely (spec DB 3 — empty per-cycle heartbeat).
- Remove the `"github.com/golang/glog"` import. Do NOT add a `log/slog` import — this file has no remaining log lines.
- The local `duration := timeOfDay.Duration(time.Now())` stays (it feeds `time.NewTimer(duration)`). Do NOT touch the `time.Now()` call — fixing it is explicitly out of scope (spec Non-goal; only this file's glog import is touched).

**7. `./pkg/check/between-time-switch.go`** (lines 42, 45)

| Old | New |
|---|---|
| `glog.V(2).Infof("now is not between %s and %s => use fallback", from, until)` | `slog.Debug("now is not between, use fallback", "from", from.String(), "until", until.String())` |
| `glog.V(2).Infof("now is between %s and %s => use main", from, until)` | `slog.Debug("now is between, use main", "from", from.String(), "until", until.String())` |

**8. `./pkg/check/checks-runner.go`** (lines 40, 43, 47)

| Old | New |
|---|---|
| `glog.V(2).Infof("%s is satisfied => skip", check.Name())` | `slog.Debug("check satisfied, skip", "check", check.Name())` |
| `glog.V(2).Infof("%s is not satisfied => apply", check.Name())` | `slog.Debug("check not satisfied, apply", "check", check.Name())` |
| `glog.V(2).Infof("%s applied", check.Name())` | `slog.Debug("check applied", "check", check.Name())` |

These three per-item result lines stay inside the loop and stay at Debug — they are adjudicated exempt from `skip-empty-v2-heartbeats` (spec DB 3). Do not remove or re-level them.

**9. `./pkg/check/checks-cron.go`** (lines 32, 34, 36)

- `glog.Warningf("run checks failed: %v", err)` → `slog.Warn("run checks failed", "error", err)`. This line stays ungated by a sampler in this prompt (the sampler is the last prompt's job); do not add sampler logic here.
- DELETE the `else { glog.V(2).Infof("all checks applied") }` branch entirely (spec DB 3). The `if err := runner.RunChecks(ctx, checks); err != nil { ... }` becomes a bare `if` with no `else`.
- DELETE the line `glog.V(2).Infof("sleep for %v", interval)` entirely (spec DB 3). The `interval` parameter is still used by `time.NewTimer(interval)` below it — the import and signature stay.

**10. `./pkg/check/checks-creator.go`** (lines 55-56, 77-78)

| Old | New |
|---|---|
| `glog.V(2).\n\tInfof("current time %s in %s", now.In(c.location).Format(time.RFC3339), c.location.String())` | `slog.Debug("current time", "time", now.In(c.location).Format(time.RFC3339), "location", c.location.String())` |
| `glog.V(2).\n\tInfof("now %s sunrise %s sunset %s", now.In(c.location).Format("15:04:05"), sunrise.In(c.location).Format("15:04:05"), sunset.In(c.location).Format("15:04:05"))` | `slog.Debug("sunrise sunset", "now", now.In(c.location).Format("15:04:05"), "sunrise", sunrise.In(c.location).Format("15:04:05"), "sunset", sunset.In(c.location).Format("15:04:05"))` |

Collapse the two-line `glog.V(2).` + `Infof(...)` forms into single-line `slog.Debug(...)` calls.

### Step 2 — Drop the direct glog requirement

Run `go mod tidy`. `github.com/golang/glog` must leave the first (direct) `require (...)` block; it may remain in the second (indirect) block because `github.com/bborbe/log` and `github.com/bborbe/service` still import it internally (spec AC 1 explicitly allows the indirect requirement). If `go mod tidy` removes anything else, restore it — the dependency set is frozen.

### Step 3 — CHANGELOG

Add a `## Unreleased` section directly above `## v0.4.2` in `./CHANGELOG.md` (never above the `# Changelog` preamble) with conventional-prefix bullets covering this prompt, e.g.:

```markdown
## Unreleased

- refactor: Migrate all hue Go code from the deprecated `github.com/golang/glog` logger to stdlib `log/slog` with explicit intent-based levels (Info for operator-facing output, Debug for state/heartbeat detail, Warn for failures) and structured key-value attributes
- refactor: Remove the three unconditional per-cycle heartbeat log lines (checks-cron "all checks applied" / "sleep for", time-of-day "next trigger in")
- fix: Stop logging the bridge `User` API key and the `%+v` whole-discovery dump; bridge-discovery lines now log count / ID / Host only
```

### Step 4 — Self-check

Walk the spec's Desired Behavior 1, 2, 3, 6 and Acceptance Criteria 1, 2, 3, 5 against the change. Re-run the `<verification>` block below and confirm every check passes, then run `make precommit` and `make test` and confirm both exit 0.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git. No `git commit`, `git push`, `git add`, no PR (spec repo CLAUDE.md + `.dark-factory.yaml`).
- Never code directly — all changes go through the dark-factory pipeline (spec Constraints, repo CLAUDE.md).
- Spec Constraints (copied verbatim): repo conventions stay frozen — `github.com/bborbe/time` not stdlib `time`; `github.com/bborbe/errors` with `errors.Wrap(ctx, err, "...")`, never bare `return err`; `github.com/bborbe/collection.Ptr()`; Ginkgo v2 / Gomega; external test packages (`pkg_test`). Factory functions stay pure composition. No new counterfeiter mock for the sampler (sibling PR 2 owns mocks). The controller's default slog level must emit the same lines the `LOGLEVEL=2` deployment emits today (output parity). The `/setloglevel` endpoint path stays `/setloglevel/{level}`. `k8s/hue-deploy.yaml` and `k8s/k8s.env` stay untouched (the `-v={LOGLEVEL}` line is glog's flag and inert after the migration). `make precommit` mandatory before commit. Existing tests keep passing; `main_test.go` / `cmd/*/main_test.go` `gexec.Build` compile checks must stay green. No scenario is added.
- Do NOT do the next prompt's work: no sampler gating in `checks-cron.go`, no aggregation of the list-lights per-light line, no slog handler / `LevelVar` / `/setloglevel` conversion, no changes to `pkg/factory/factory.go` or `pkg/handler/` — those land in the subsequent prompts of this spec.
- Do NOT touch `pkg/check/checks-parity_test.go`, `pkg/check/checks-creator_test.go`, `pkg/check/between-time-switch_test.go`, or any `mocks/` file — the conversion is log-line-only; no test or mock changes are expected. If `make test` fails for a reason unrelated to the import swap, STOP and report rather than editing tests.
- Do NOT touch `k8s/` files, `go.mod` beyond `go mod tidy`, or `tools.env`.
- Do NOT renumber the prompt filename — dark-factory assigns numbers on approve.
- Spec Non-goals that remain out of scope here: `pkg/check` interface injection (sibling PR 1), counterfeiter mocks (sibling PR 2), validation / `Lights` rename (sibling PR 4), fixing `pkg/trigger/time-of-day.go`'s `time.Now()` (tracked separately), adding any env/flag log-level config surface, introducing a logging wrapper package (use stdlib slog package-level calls directly).
</constraints>

<verification>
All commands run from the repo root. Zero-match greps use a form with a truthful exit code.

1. `[ -z "$(grep -r 'glog' --include='*.go' .)" ]` — must exit 0 (spec AC 1: zero glog in hue's Go code, including comments).
2. `[ -z "$(sed -n '/^require (/,/^)/p' go.mod | grep 'github.com/golang/glog')" ]` — must exit 0 (spec AC 1: direct dependency dropped; indirect may remain).
3. `[ -z "$(grep -rn 'slog.Debug' cmd/ --include='main.go')" ]` — must exit 0 (spec AC 3).
4. `grep -rn 'slog.Info' cmd/ --include='main.go'` — must print ≥6 lines (spec AC 3: 2 per CLI × list-lights/turnon-light/turnoff-light).
5. `grep -rn 'slog.Debug' pkg/check/ pkg/bridges-provider.go --include='*.go'` — must print ≥6 lines (spec AC 3: 2 creator + 3 runner + 2 between-time-switch + 2 bridges-provider = 9).
6. `[ -z "$(grep -rn 'all checks applied\|sleep for\|next trigger in' pkg/ --include='*.go')" ]` — must exit 0 (spec AC 5: empty heartbeats gone).
7. `[ -z "$(grep -rn 'fmt.Sprintf' pkg/ cmd/ --include='*.go' | grep -i slog)" ]` — must exit 0 (spec DB 1: no `fmt.Sprintf` inside slog calls).
8. `grep -n 'User\|%+v' pkg/bridges-provider.go` — must print ONLY the `User: token.String()` struct-construction line (no `User` or `%+v` in any log call; spec DB 6).
9. `make precommit` — must exit 0 (DoD gate per `docs/dod.md`).
10. `make test` — must exit 0.
</verification>
