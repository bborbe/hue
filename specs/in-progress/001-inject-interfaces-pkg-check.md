---
status: verifying
tags:
    - dark-factory
    - spec
approved: "2026-09-06T10:10:27Z"
generating: "2026-09-06T10:21:24Z"
prompted: "2026-09-06T10:21:24Z"
verifying: "2026-09-06T10:32:21Z"
branch: dark-factory/inject-interfaces-pkg-check
---

## Summary

- PR 1 of the `pkg/check` refactor split: close the last two direct-package-dependency findings in the checks layer so its business logic calls no `time.Now()`, no `time.LoadLocation`, and no sunrise-library functions.
- The checks creator gains a sunrise/sunset capability interface and a pre-resolved time-zone location as constructor parameters; both are constructed once at the composition root (`main.go`) and threaded through the factory.
- The checks runner gains the clock as a constructor parameter (structural DI, no behavior change).
- A parity test renders the 9-check light-flip schedule through the pre-refactor computation and through the injected path at fixed instants and asserts byte-for-byte equality; a live pre-/post-merge cron-tick capture on quant confirms the same at runtime.
- The clock and `BridgesProvider` injection already landed on master v0.3.2 — this spec closes the remaining findings, it does not redo the finished parts.

## Problem

`pkg/check/checks-creator.go` computes sunrise and sunset by calling the `sunrisesunset` library directly inside `CreateChecks` and resolves the `Europe/Berlin` time zone via `time.LoadLocation`. Both are package-level calls in business logic: they match the `no-package-function-calls-in-business-logic` mechanical finding, keep the code-review gate red, and make the light-flip schedule verifiable only by live observation. `pkg/check/checks-runner.go` is the only check-layer constructor that still takes no clock, so the layer's DI pattern is inconsistent. Every future schedule change stays unprovable unless these dependencies move behind constructor-injected parameters.

## Goal

After this work, `pkg/check` business logic contains zero direct clock calls and zero package-function calls: sunrise/sunset arrives via a constructor-injected capability, the time-zone location is resolved once outside the checks layer, and the checks runner takes its clock via constructor exactly like the creator already does. The schedule the controller applies is byte-for-byte identical before and after the change, proven by an in-repo parity test and by a live pre-/post-merge capture diff on quant.

## Non-goals

- Counterfeiter mocks for the new interfaces — sibling PR 2
- glog → slog — sibling PR 3
- Validation / `Lights` rename — sibling PR 4
- `pkg/trigger` stays untouched (legacy, not wired into `main.go`; its `time.Now()` is outside this PR's verification scope)
- `pkg/time-of-day.go` needs no change — its `Duration(now)` already receives the instant as a parameter
- No new log output in the checks-cron path (any added line would break the parity capture diff)
- No new feature work

## Acceptance Criteria

- [ ] `pkg/check` contains zero direct `time.Now()` calls — evidence (negative): `grep -rn "time.Now()" pkg/check/` returns 0 lines.
- [ ] `pkg/check` contains zero references to the sunrise library outside test files — evidence (negative): `grep -rn "sunrisesunset" pkg/check/ --include='*.go' | grep -v '_test.go'` returns 0 lines (the parity test's reference computation is the single allowed occurrence and lives in a test file).
- [ ] `pkg/check` contains zero `time.LoadLocation` calls outside test files — evidence (negative): `grep -rn "LoadLocation" pkg/check/ --include='*.go' | grep -v '_test.go'` returns 0 lines (existing test code resolves `Europe/Berlin` via `LoadLocation` to build the injected location and stays).
- [ ] Constructor DI wired end-to-end — evidence: `grep -n "libtime.CurrentDateTimeGetter" pkg/check/checks-runner.go` returns ≥1 line; `grep -c "currentDateTimeGetter" pkg/factory/factory.go` returns ≥3 (signature + creator call + runner call); `grep -n "libtime.NewCurrentDateTime()" main.go` returns ≥1 line (clock created once); `grep -n "time.LoadLocation" main.go` returns ≥1 line (location resolved once at the composition root); `grep -rni "sunrisesunset" main.go pkg/factory/` returns ≥1 line (capability constructed once and threaded through).
- [ ] In-repo parity test proves the schedule is byte-for-byte identical before/after — evidence: `pkg/check/` contains a parity test that renders the 9-check schedule (each check's name and on/off state at the instant, plus sunrise and sunset times) through the pre-refactor reference computation and through the injected path at ≥2 fixed instants (one summer-date, one winter-date, both `summerMode` values) and asserts byte equality; `grep -rn "sunrisesunset" pkg/check/ --include='*_test.go'` returns ≥1 line (the parity test's independent reference computation); `make test` exits 0.
- [ ] Code-review gate: mechanical ast-grep run plus LLM adjudication on the changed `pkg/check` files reports zero genuine findings for `go-time/no-time-now-direct` and `go-composition/no-package-function-calls-in-business-logic` (the only mechanically-flagged calls remaining are `errors.Wrap` — the mandated wrapping lib — and same-package constructors; glog is PR 3) — evidence: `bash ~/.claude/plugins/marketplaces/coding/scripts/ast-grep-runner.sh /Users/bborbe/Documents/workspaces/hue-inject-interfaces-pkg-check/pkg/check` JSON shows both rule ids at 0 genuine findings; `/coding:code-review` output confirms.
- [ ] `make precommit` exits 0 — evidence: exit code.
- [ ] **Post-Deploy (Rung-3):** live parity — the post-merge master cron-tick capture shows byte-for-byte identical schedule lines (the sunrise/sunset line plus the per-check satisfied/apply lines, after stripping glog headers and normalizing the per-cycle `now` field) against the pre-merge feature-branch capture taken on the same UTC date at the same time-of-day; `kubectlnukeprod -n hue get pods -l app=hue` shows `Running 1/1` and logs show no panic / repeated error pattern. (Note: hue runs on nuke-prod, not quant — migrated 2026-08-21 `a862f4a`.)
  - `deploy_check:` `kubectlnukeprod -n hue get deploy/hue -o jsonpath='{.spec.template.spec.containers[0].image}' | awk -F: '{print $NF}'`
  - `deploy_target:` `master`

## Verification

## Container-executable (runs inside the YOLO container at prompt time)

```
cd /Users/bborbe/Documents/workspaces/hue-inject-interfaces-pkg-check && make precommit   # exits 0
make test                                                                                  # exits 0
make check                                                                                 # exits 0
grep -rn "time.Now()" pkg/check/                                                           # 0 lines
grep -rn "sunrisesunset" pkg/check/ --include='*.go' | grep -v '_test.go'                  # 0 lines
grep -rn "LoadLocation" pkg/check/ --include='*.go' | grep -v '_test.go'                   # 0 lines
grep -n "libtime.CurrentDateTimeGetter" pkg/check/checks-runner.go                          # ≥1 line
grep -c "currentDateTimeGetter" pkg/factory/factory.go                                     # ≥3
grep -n "libtime.NewCurrentDateTime()" main.go                                             # ≥1 line
grep -n "time.LoadLocation" main.go                                                        # ≥1 line
grep -rni "sunrisesunset" main.go pkg/factory/                                             # ≥1 line
```

## Operator-executable (runs on the host after PR merge, spec verification ladder)

```
# code-review gate
bash ~/.claude/plugins/marketplaces/coding/scripts/ast-grep-runner.sh /Users/bborbe/Documents/workspaces/hue-inject-interfaces-pkg-check/pkg/check   # both rules: 0 genuine findings
/coding:code-review on pkg/check                                                            # both rules at zero

# pre-merge capture (feature branch deployed to nuke-prod)
kubectlnukeprod -n hue get pods -l app=hue                                                   # Running 1/1
kubectlnukeprod -n hue logs -l app=hue --tail=500 --since=3m > /tmp/pre-merge.raw
# extract one full 60s cycle's schedule lines (strip glog header prefix; drop `current time` and `sleep for` lines):
grep -E "sunrise|satisfied|applied" /tmp/pre-merge.raw | sed -E 's/^[IWEF][0-9]{4} [0-9:.]+ +[0-9]+ [^]]+\] //; s/now [0-9]{2}:[0-9]{2}:[0-9]{2}/now <TIME>/' | sort -u > /tmp/pre-merge.schedule

# post-merge capture (master deployed, SAME UTC date, same time-of-day)
kubectlnukeprod -n hue logs -l app=hue --tail=500 --since=3m > /tmp/post-merge.raw
grep -E "sunrise|satisfied|applied" /tmp/post-merge.raw | sed -E 's/^[IWEF][0-9]{4} [0-9:.]+ +[0-9]+ [^]]+\] //; s/now [0-9]{2}:[0-9]{2}:[0-9]{2}/now <TIME>/' | sort -u > /tmp/post-merge.schedule

diff /tmp/pre-merge.schedule /tmp/post-merge.schedule                                      # empty
```

## Desired Behavior

1. The checks creator obtains sunrise and sunset times through a constructor-injected sunrise/sunset capability instead of calling the sunrise library directly. The real implementation reproduces the pre-refactor inputs exactly: fixed coordinates latitude 50.1, longitude 8.1, UTC offset 0, and the date taken from the injected clock's instant.
2. The checks creator obtains the `Europe/Berlin` time zone through a constructor-injected `*time.Location` instead of calling `time.LoadLocation`. The location is resolved exactly once at the composition root (`main.go`), which owns the load error.
3. The checks runner accepts the clock via its constructor and retains it; it adds no timing behavior and no new log output, so the cron-loop capture remains diff-comparable.
4. Composition wiring: `main.go` constructs the clock (once), the time-zone location (once), and the sunrise/sunset capability (once); factory functions thread them through constructor calls only, adding no conditionals, no I/O, and no `context.Background()`.
5. A parity test renders the full 9-check schedule through both the pre-refactor reference computation (the direct `sunrisesunset` call with the exact pre-refactor parameters) and the injected path at fixed instants, and asserts the two rendered outputs are byte-for-byte identical. Both `summerMode` values are swept at each instant.
6. The checks-cron path emits exactly the same V(2) log lines as before the change (no new timestamps, no new messages, same wording), so the pre-/post-merge live capture is diff-comparable.

## Constraints

- `BridgesProvider` stays composed (`func` → `cache` → `fallback`); never collapsed to a single struct.
- Factory functions stay pure composition: no conditionals, no I/O, no `context.Background()`.
- `github.com/bborbe/time` is injected via `libtime.CurrentDateTimeGetter` in constructors and created once in `main.go` — already true at master v0.3.2 and must remain.
- `github.com/bborbe/errors` wrapping via `errors.Wrap(ctx, err, "...")`; never bare `return err`.
- Sunrise library injected via interface (frozen design rule).
- The sunrise/sunset capability must produce the same values as the pre-refactor direct call — the parity test and live capture are the enforcement.
- The time-zone location stays `Europe/Berlin`; no config surface is added.
- No new log lines in the checks-cron path.
- Existing tests keep passing; tests may be edited only for the new constructor parameters, and the parity test uses the real concrete capability (counterfeiter mocks are PR 2).
- `make precommit` mandatory before commit.

## Assumptions

- Clock and `BridgesProvider` injection already landed on master v0.3.2 (commit 94634f9); the delta of this PR is the sunrise/sunset capability, the time-zone location injection, the runner constructor, the composition wiring, and the parity proof.
- The task's "time-of-day.go" is `pkg/time-of-day.go`; it already receives the instant as a parameter and requires no change.
- `pkg/trigger` is legacy code not wired into `main.go`; its `time.Now()` is outside this PR's verification scope.
- The quant deployment runs at `LOGLEVEL=2`, so the V(2) cron-tick lines are emitted and capturable.
- The two live captures happen on the same UTC date at the same time-of-day; the deployed config uses `SUMMER_MODE=false`. Avoid capturing across schedule transition boundaries (08:00/10:00/18:00/20:00/21:00/23:00 Berlin), where a single 60s cycle can legitimately straddle an on/off transition and produce extra lines.

## Failure Modes

| Trigger | Expected behavior | Recovery | Detection | Reversibility |
|---|---|---|---|---|
| Hue bridges unreachable during a live capture | Runner logs "run checks failed"; capture lacks the per-check satisfied/apply lines | Redo the capture when bridges are reachable | Capture file is missing check lines | Reversible |
| Pre- and post-merge captures on different UTC dates | Sunrise/sunset line differs; `diff` non-empty | Redo the post-merge capture on the same UTC date at the same time-of-day | `diff` shows only the sunrise/sunset line differing | Reversible |
| Feature-branch deploy and master deploy overlap | Two controllers drive the physical lights concurrently | Sequence the deploys: wait for the `Recreate` rollout of one to finish before starting the other; `replicas` stays 1 | Two pods with different image tags observed in the same window | Reversible |
| A future dependency bump changes sunrise values | Parity test fails | Handle in a separate spec; do not silently update the reference | Parity test red | Reversible |

## Security / Abuse Cases

Not applicable: the change is constructor plumbing with no HTTP, file, or user-input surface. The sunrise/sunset capability computes from fixed compile-time constants (coordinates, offset); the time-zone location is a fixed string resolved once in `main.go`; nothing attacker-controlled crosses a trust boundary, and nothing added can hang, retry forever, or race.

## Suggested Decomposition

| # | Prompt focus | Covers DBs | Covers ACs | Depends on |
|---|---|---|---|---|
| 1 | Sunrise/sunset capability: interface plus real concrete implementation in `pkg/` wrapping the `sunrisesunset` library with the exact pre-refactor parameters (lat 50.1, lon 8.1, offset 0, date from the clock instant) | 1 | 2 | — |
| 2 | Inject the sunrise capability and the time-zone location into `NewCheckCreator`; drop `time.LoadLocation` and all `sunrisesunset` usage from `pkg/check`; update the existing `checks-creator_test.go` for the new constructor parameters | 1, 2, 6 | 1, 2, 3 | prompt 1 |
| 3 | `checks-runner.go` constructor takes the clock; `pkg/factory/factory.go` threads location + sunrise + clock into the creator and clock into the runner (pure composition); `main.go` resolves the location once and constructs the clock and the capability once | 3, 4 | 4 | prompts 1, 2 |
| 4 | Parity test: reference computation (direct `sunrisesunset` call, exact pre-refactor parameters) vs injected path, rendered byte-for-byte at ≥2 fixed instants × both `summerMode` values | 5 | 5 | prompts 2, 3 |

Rationale: prompt 1 establishes the capability contract first; prompt 2 rewires the creator to consume it (the code-review findings close here); prompt 3 completes the DI end-to-end; prompt 4 adds the proof on top of settled code. Each prompt leaves the repo compiling and `make test` green.

## Do-Nothing Option

`pkg/check` keeps the direct `sunrisesunset` and `time.LoadLocation` calls: the `no-package-function-calls-in-business-logic` finding stays red, the light-flip schedule stays verifiable only by live observation, and every future schedule change carries unnoticed-drift risk. The refactor is already split and PR 1 is the agreed first step; skipping it leaves the parent checklist open and blocks PRs 2-4 that build on the injected interfaces.
