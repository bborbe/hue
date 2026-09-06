---
status: completed
approved: "2026-09-06T18:44:29Z"
generating: "2026-09-06T18:55:14Z"
prompted: "2026-09-06T18:55:14Z"
verifying: "2026-09-06T20:19:39Z"
completed: "2026-09-06T20:19:47Z"
branch: dark-factory/migrate-glog-slog
---

## Summary

- Replace the deprecated `github.com/golang/glog` logger with the org-standard `log/slog` across all of hue's own Go code (10 files, 20 call-sites); the direct dependency drops out of `go.mod`.
- Fix the three behavioral logging findings the migration surfaces — they are fixed as part of the migration, not as an import swap: `use-v-for-debug-not-info` (4 sites), `no-tight-loop-without-sampler` (2 sites), `skip-empty-v2-heartbeats` (3 lines).
- Every converted log line carries an explicit level decided by operator-visibility (info for operator-facing output, debug for state/heartbeat detail), and structured key-value attributes instead of interpolated strings.
- The controller's default log level keeps emitting the same lines the deployment's `LOGLEVEL=2` emits today (output parity); the `/setloglevel` endpoint is converted to control a slog level instead of becoming a dead glog knob.
- `github.com/bborbe/log` provides the sampler (logger-agnostic `Sampler.IsSample()`); dependency libraries that still emit glog internally are out of scope.

## Problem

`glog` is deprecated; the org standard is `log/slog`, and the fleet task [[Migrate All Go Projects from glog to slog]] tracks hue among the remaining repos. Beyond the logger swap, the migration surfaces three review findings that only make sense to fix while the lines are being rewritten: four bare `glog.Info`/`Infof` calls in the CLIs whose level is decided by glog's always-on V(0) semantics instead of by message intent; two `for`-loop bodies with unsampled log calls (the checks-cron cycle loop and the list-lights listing loop); and three unconditional per-cycle heartbeat lines that emit once per loop iteration no matter whether anything observable happened. Leaving these open keeps hue on a deprecated logger whose verbosity semantics no longer match how the code logs. A slog migration that merely swaps the import would silently drop operator-visible heartbeat output (level mismatch) or keep an always-on V(0)-style default (noise), so the level mapping and the finding fixes are part of the migration itself.

## Goal

After this work, hue's own Go code contains zero `glog` references. Every log line is a structured `slog` call (key-value attributes, no `fmt.Sprintf`-interpolated message strings) with an explicit, intent-based level: operator-facing output at `slog.Info`, debug/heartbeat/state detail at `slog.Debug`, failures at `slog.Warn`. The controller emits the same lines at its default level that the `LOGLEVEL=2` deployment emits today. No log line fires from a loop merely because the loop ran: the checks-cron cycle loop logs only through a sampler, and the list-lights CLI emits its full listing without a per-iteration log call. The `/setloglevel/{level}` endpoint keeps its operational purpose (bump verbosity at runtime, auto-reset) but drives a slog level instead of a glog verbosity. The four review finding classes (`slog-not-glog-in-new-projects`, `use-v-for-debug-not-info`, `no-tight-loop-without-sampler`, `skip-empty-v2-heartbeats`) report zero findings on the funnel.

## Non-goals

- `pkg/check` interface injection — sibling PR 1
- Counterfeiter mocks — sibling PR 2; the sampler is tested with `github.com/bborbe/log`'s `DefaultSamplerFactory`, no new mock
- Validation / `Lights` rename — sibling PR 4
- Any repo other than hue — covered by [[Migrate All Go Projects from glog to slog]]
- Fixing `pkg/trigger/time-of-day.go`'s `time.Now()` (`go-time/no-time-now-direct`) — tracked separately in the parent task; only its `glog` import is touched here
- Re-plumbing the k8s `LOGLEVEL` env / `-v` flag as a slog knob — the `-v={LOGLEVEL}` line in `k8s/hue-deploy.yaml` stays untouched (it is glog's flag, inert after the migration); no new env/flag log-level config surface is added
- Introducing a logging wrapper package — use stdlib `slog` package-level calls directly
- Keeping `/setloglevel` wired to `github.com/bborbe/log`'s glog-typed `NewLogLevelSetter` — that would be a dead knob after the migration; the endpoint is converted to a slog level setter instead
- Log output of `github.com/bborbe/service` / `github.com/bborbe/log` themselves (they still emit glog lines internally) — the fleet task migrates the libraries; hue's own files carry zero glog

## Acceptance Criteria

- [ ] **Zero glog in hue's Go code** — evidence (negative): `grep -r "glog" --include="*.go" .` returns 0 lines; the direct dependency is gone — `sed -n '/^require (/,/^)/p' go.mod | grep "github.com/golang/glog"` returns 0 lines (an indirect requirement via `bborbe/log` / `bborbe/service` may remain).
- [ ] **Funnel cleared** — evidence: `bash ~/.claude/plugins/marketplaces/coding/scripts/ast-grep-runner.sh /Users/bborbe/Documents/workspaces/hue-migrate-glog-slog` JSON shows `slog-not-glog-in-new-projects`, `use-v-for-debug-not-info`, `skip-empty-v2-heartbeats` at 0 findings, and `no-tight-loop-without-sampler` at 0 genuine findings after adjudication (the only remaining mechanical hits, if any, are adjudicated-exempt per DB 3).
- [ ] **`use-v-for-debug-not-info` resolved** — the 4 formerly-bare CLI `glog.Info(f)` sites land at `slog.Info`, as do the two `V(2)` operator-feedback lines — evidence: `grep -rn "slog.Debug" cmd/ --include="main.go"` returns 0 lines AND `grep -rn "slog.Info" cmd/ --include="main.go"` returns ≥6 lines AND `grep -rn "slog.Debug" pkg/check/ pkg/bridges-provider.go --include="*.go"` returns ≥6 lines (DB 2 pkg/ rows stay at debug, not flattened to Info).
- [ ] **`no-tight-loop-without-sampler` resolved** — checks-cron's cycle loop logs through a sampler, list-lights' loop contains no per-iteration log call — evidence: `grep -rn "IsSample\|NewSampleTime\|SamplerFactory" pkg/check/ pkg/factory/ main.go --include="*.go"` returns ≥1 line; the funnel AC above passes.
- [ ] **`skip-empty-v2-heartbeats` resolved** — the three unconditional per-cycle heartbeat lines are gone — evidence (negative): `grep -rn "all checks applied\|sleep for\|next trigger in" pkg/ --include="*.go"` returns 0 lines.
- [ ] **`/setloglevel` converted to slog** — evidence: `grep -rn "NewLogLevelSetter\|NewSetLoglevelHandler" . --include="*.go"` returns 0 lines AND `grep -n "setloglevel" main.go` returns ≥1 line (endpoint still registered) AND `grep -n "LevelVar" main.go` returns ≥1 line (the endpoint drives a slog level).
- [ ] `make precommit` exits 0 — evidence: exit code.
- [ ] **Post-Deploy (Rung-3):** the merged controller on nuke-prod logs structured slog lines (no glog-header-prefixed hue lines) and shows no panic / repeated errors — evidence: `kubectlnukeprod -n hue logs -l app=hue --tail=50` shows `slog`-shaped key-value lines and returns no `panic:` match; `kubectlnukeprod -n hue get pods -l app=hue` shows `Running 1/1`.
  - `deploy_check:` `kubectlnukeprod -n hue get deploy/hue -o jsonpath='{.spec.template.spec.containers[0].image}' | awk -F: '{print $NF}'`
  - `deploy_target:` `$(git rev-parse --short HEAD)`

**Scenario coverage — default: NO new scenario.** Unit and integration tests fully reach this change (logging levels, sampler gating, structured output are all testable in-process; the CLI and controller are already covered by `gexec.Build` compile checks in the existing `main_test.go` files). No real cluster or external dependency is required to observe the behavior.

## Verification

### Container-executable (runs inside the YOLO container at prompt time)

```
cd /Users/bborbe/Documents/workspaces/hue-migrate-glog-slog && make precommit   # exits 0
make test                                                                       # exits 0
grep -r "glog" --include="*.go" .                                               # 0 lines
sed -n '/^require (/,/^)/p' go.mod | grep "github.com/golang/glog"              # 0 lines
grep -rn "slog.Debug" cmd/ --include="main.go"                                  # 0 lines
grep -rn "slog.Info" cmd/ --include="main.go"                                   # ≥6 lines
grep -rn "slog.Debug" pkg/check/ pkg/bridges-provider.go --include="*.go"       # ≥6 lines
grep -rn "IsSample\|NewSampleTime\|SamplerFactory" pkg/check/ pkg/factory/ main.go --include="*.go"   # ≥1 line
grep -rn "all checks applied\|sleep for\|next trigger in" pkg/ --include="*.go" # 0 lines
grep -rn "NewLogLevelSetter\|NewSetLoglevelHandler" . --include="*.go"          # 0 lines
grep -n "setloglevel" main.go                                                   # ≥1 line
grep -n "LevelVar" main.go                                                      # ≥1 line
```

### Operator-executable (runs on the host after PR merge, spec verification ladder)

```
bash ~/.claude/plugins/marketplaces/coding/scripts/ast-grep-runner.sh /Users/bborbe/Documents/workspaces/hue-migrate-glog-slog   # 3 rules at 0; no-tight-loop at 0 genuine
kubectlnukeprod -n hue get pods -l app=hue      # Running 1/1 (post-deploy, per task Definition of Done)
kubectlnukeprod -n hue logs -l app=hue --tail=50   # structured slog lines, no panic / repeated errors
```

## Desired Behavior

1. **Repo-wide import swap.** All 10 Go files that import `github.com/golang/glog` switch to `log/slog`; all 20 call-sites become package-level `slog.Info` / `slog.Debug` / `slog.Warn` / `slog.Error` calls with key-value attribute pairs. Messages are never built with `fmt.Sprintf` inside a `slog` call — values become attributes. `go mod tidy` removes the direct `github.com/golang/glog` requirement.

2. **Explicit level classification** (this is what resolves `use-v-for-debug-not-info` and defines the level mapping for every converted line):

   | Current line | Target |
   |---|---|
   | `main.go` "starting http server listen on %s" (V(2)) | `slog.Info` — startup |
   | `cmd/turnon-light` "light already on" (V(2)) | `slog.Info` — operator action feedback |
   | `cmd/turnoff-light` "light already off" (V(2)) | `slog.Info` — operator action feedback |
   | `cmd/list-lights` "found %d lights" (bare Info) | `slog.Info` — CLI result |
   | `cmd/list-lights` "'%s' on: %v" per light (bare Info) | `slog.Info` — CLI output |
   | `cmd/turnon-light` "light turned on" (bare Info) | `slog.Info` — CLI result |
   | `cmd/turnoff-light` "light turned off" (bare Info) | `slog.Info` — CLI result |
   | `pkg/check/checks-cron` "run checks failed: %v" (Warning) | `slog.Warn` — sampler-gated (see DB 4) |
   | `pkg/check/checks-creator` "current time … in …" / "now … sunrise … sunset …" (V(2)) | `slog.Debug` |
   | `pkg/check/checks-runner` per-check "satisfied => skip" / "not satisfied => apply" / "applied" (V(2)) | `slog.Debug` — kept (per-item result lines, exempt from `skip-empty-v2-heartbeats`) |
   | `pkg/check/between-time-switch` "now is (not) between …" (V(2)) | `slog.Debug` |
   | `pkg/bridges-provider` discovery lines (V(2)) | `slog.Debug` — count / ID / Host only, never `User` or `%+v` (see DB 6) |

3. **Empty heartbeats removed** (resolves `skip-empty-v2-heartbeats`, 3 lines). The three unconditional per-cycle lines — `pkg/check/checks-cron.go` "all checks applied" and "sleep for %v", and `pkg/trigger/time-of-day.go` "next trigger in %v" — are removed; no line fires from these loops solely because a cycle ran. The two mechanically-flagged loops in `checks-runner.go` (per-item result lines) and `bridges-provider.go` (the `found:` line, guarded by the bridge-ID match) are adjudicated exempt and stay.

4. **Tight-loop logging is sampled or aggregated** (resolves `no-tight-loop-without-sampler`, 2 sites). The checks-cron cycle loop (runs every 60s) logs only through a `github.com/bborbe/log` sampler: a `log.SamplerFactory` is injected at construction per the repo's DI convention (`NewCheckCron` gains the parameter; `pkg/factory/factory.go` threads it; `main.go` constructs the time sampler at the composition root (`log.NewSampleTime(10 * time.Minute)`); tests use `log.DefaultSamplerFactory`, no new counterfeiter mock), and the remaining `slog.Warn` failure line is emitted only when `sampler.IsSample()` is true, capped at most once per 10 minutes (`log.NewSampleTime(10 * time.Minute)`). The list-lights listing loop emits its per-light detail as a single aggregated output (the full listing in one emission after the loop) — a sampler gate is NOT acceptable there because it would truncate the CLI's output mid-listing.

5. **Logger setup per binary + `/setloglevel` conversion.** Each binary installs a default slog handler at its composition root: the controller (`main.go`) uses a `slog.LevelVar` initialized to `slog.LevelDebug` — this is what makes the deployment's current `LOGLEVEL=2` output parity hold — and the `cmd/*` CLIs default to `slog.LevelInfo`. Handler output is text to stderr (matches today's `logtostderr` behavior). The `/setloglevel/{level}` route keeps its current shape but is converted from `log.NewSetLoglevelHandler(ctx, log.NewLogLevelSetter(...))` to a setter that maps the level string onto the `LevelVar` with the existing TTL-reset semantics (reset to the debug default after 5 minutes). No new env/flag log-level knob is added.

6. **Bridge-discovery lines carry no credential-shaped data.** The converted `pkg/bridges-provider.go` lines log the discovered bridge count and per-bridge `ID` + `Host` only. The `%+v` whole-list dump and the `User` field (the huego bridge API key used in `/api/{user}/...`) are not logged.

## Constraints

- Never code directly — all changes go through the dark-factory pipeline (repo CLAUDE.md).
- Repo conventions stay frozen: `github.com/bborbe/time` not stdlib `time`; `github.com/bborbe/errors` with `errors.Wrap(ctx, err, "...")`, never bare `return err`; `github.com/bborbe/collection.Ptr()`; Ginkgo v2 / Gomega; external test packages (`pkg_test`).
- Factory functions stay pure composition — the sampler factory is created at the composition root (`main.go`) and threaded through `pkg/factory/factory.go`; no conditionals, no I/O, no `context.Background()` in factories.
- No new counterfeiter mock for the sampler (sibling PR 2 owns mocks); tests use `github.com/bborbe/log`'s `DefaultSamplerFactory`.
- The controller's default slog level must emit the same lines the `LOGLEVEL=2` deployment emits today (output parity).
- The `/setloglevel` endpoint path stays `/setloglevel/{level}`.
- `k8s/hue-deploy.yaml` and `k8s/k8s.env` stay untouched (the `-v={LOGLEVEL}` line is glog's flag and inert after the migration).
- `make precommit` mandatory before commit (ensure + format + generate + test + check + addlicense).
- Existing tests keep passing; `main_test.go` / `cmd/*/main_test.go` `gexec.Build` compile checks must stay green.
- No scenario is added (see the Scenario coverage note above).

## Assumptions

- **Finding-site inventory** (derived from `ast-grep-runner.sh` on this worktree, branch `feature/migrate-glog-slog`): `use-v-for-debug-not-info` = 4 mechanical sites (the bare `glog.Info(f)` calls in `cmd/list-lights`, `cmd/turnon-light`, `cmd/turnoff-light`); `no-tight-loop-without-sampler` = 2 mechanical sites (`pkg/check/checks-cron.go` loop, `cmd/list-lights/main.go` loop); `skip-empty-v2-heartbeats` = 4 mechanical loop findings, of which 3 are genuine empty heartbeats (the 3 unconditional per-cycle lines in DB 3) and 2 loops adjudicate exempt (checks-runner's per-item result lines, bridges-provider's ID-guarded `found:` line).
- **Level mapping** preserves today's deployed output: `slog.LevelDebug` as the controller default mirrors `LOGLEVEL=2`; the `-v` flag and `LOG_LEVEL` env handling inside `bborbe/service` remain and are inert for hue's own lines.
- **`/setloglevel` conversion** is in scope as part of the repo-wide migration: leaving it wired to the glog-typed `log.NewLogLevelSetter` would make it a dead knob (it would set glog verbosity that no hue line reads). The mapping of the numeric level path segment to `slog.Level` (0 → `Info`, ≥1 → `Debug`) is an implementation decision the prompt makes; the endpoint must keep working and must not panic on invalid input.
- **List-lights aggregation** is the sanctioned fix for its loop finding because a sampler would randomly truncate the CLI's output; aggregation is the documented "log once after the loop" pattern from the go-logging guide.
- **`github.com/bborbe/service` and `github.com/bborbe/log` continue to emit glog lines internally** (e.g. `application started`) until the fleet task migrates the libraries — that output is out of hue's control and out of scope; hue's own files carry zero glog.
- huego's `Bridge.User` is the bridge API key (it is used verbatim in the API path `/api/{user}/...`) and must not appear in logs.

## Failure Modes

| Trigger | Expected behavior | Recovery |
|---|---|---|
| Controller default level set to `Info` instead of `Debug` (level-parity regression) | Deployment logs lose the per-check / sunrise-sunset debug lines operators see today | Set the `LevelVar` default to `slog.LevelDebug`; verify by diffing a pre-/post-merge log capture |
| Sampler gate applied to the list-lights per-light output | `hue list-lights` prints only a subset of lights (mid-listing truncation) | Revert to the single aggregated emission; re-run the CLI |
| `/setloglevel` conversion mis-wired (route missing, setter not driving the `LevelVar`) | `/setloglevel/{n}` returns an error or changes nothing | Verify the route is still registered and the setter writes the `LevelVar`; AC 6 greps catch it |
| Invalid level value sent to `/setloglevel` (attacker or operator typo) | Handler returns an error response, logs nothing, and the `LevelVar` is unchanged; no panic | Validate the path segment before parsing (see Security); revert via TTL reset |
| A converted line accidentally logs the bridge `User` / `%+v` dump | Bridge API key appears in logs | Drop the field at the call-site; the `no-sensitive-data-in-logs` mechanical rule flags it on review |
| `go mod tidy` removes a still-needed direct dependency | Build fails | Restore the requirement; `make precommit` gates it |
| Duplicate deploys (controller vs master overlap) during verification | Two controllers drive the physical lights concurrently | Sequence the deploys; `replicas` stays 1 |

## Security / Abuse Cases

- **`/setloglevel/{level}` input** (HTTP user input): the path segment is attacker- or operator-controlled. It must be validated (numeric, bounded) before it touches the `LevelVar`; an invalid value yields an error response and leaves the level unchanged — no panic, no raw-input echo into logs. The endpoint is unauthenticated today (same as pre-migration) — no new trust boundary is introduced.
- **Log content**: the migration touches lines that previously interpolated `%+v` whole structs and the bridge API key. The converted lines must not log credential-shaped data (`User`, token values, whole discovery structs) — only count / ID / Host. This is DB 6 and is enforced by the `no-sensitive-data-in-logs` review rule.
- **What can hang/retry/race**: the sampler is a plain `IsSample()` call (no I/O); the `LevelVar` write is atomic. No new concurrency. The checks-cron loop already selects on `ctx.Done()` and keeps that behavior.

## Suggested Decomposition

| # | Prompt focus | Covers DBs | Covers ACs | Depends on |
|---|---|---|---|---|
| 1 | Repo-wide glog→slog conversion: all 10 files / 20 call-sites, per-message level classification per the DB 2 table, drop the 3 empty-heartbeat lines (DB 3), sensitive-data-safe bridges-provider lines (DB 6), `go mod tidy` | 1, 2, 3, 6 | 1, 2, 3, 5 | — |
| 2 | Logger setup per binary + `/setloglevel` conversion: default slog handler with `LevelVar` (controller `Debug` for `LOGLEVEL=2` parity, CLIs `Info`), replace the glog-typed setter with a slog `LevelVar` setter keeping the TTL reset, text output to stderr | 5 | 6 | prompt 1 |
| 3 | Sampler wiring: `NewCheckCron` gains a `log.SamplerFactory` parameter (10-minute time sampler gating the failure `Warn`), threaded through `pkg/factory/factory.go` from `main.go`; list-lights per-light listing aggregated into a single emission | 4 | 4 | prompts 1, 2 |

Rationale: prompt 1 is the bulk mechanical sweep and must land first so the repo compiles on slog alone; prompt 2 then establishes the level/logger plumbing the migration's output parity depends on; prompt 3 wires the sampler through the same `main.go` / factory layer last; the main.go-touching prompts are serialized by their dependencies. Each prompt leaves the repo compiling and `make test` green.

## Do-Nothing Option

Hue stays on a deprecated logger with glog's V(0)/V(2) semantics that no longer match the code's actual logging intent; the four review finding classes stay red on the funnel, the parent task's checklist stays open, and PR 4 (which builds on the merged state) stays blocked. The `/setloglevel` endpoint would keep controlling glog verbosity that hue's own lines no longer use once any partial slog migration lands, so the do-nothing option is only coherent as "do nothing at all" — which leaves the deprecated-logger debt and the findings in place.
