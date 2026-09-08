---
status: completed
summary: 'Added DRY_RUN env var / -dry-run CLI flag: DryRun field in main.go, dryRun pass-through in the factory, dry-run branch in the checks runner (Info log + skip Apply), k8s DRY_RUN entry pinned false, CHANGELOG Unreleased bullet, and a Ginkgo test asserting both branches via ApplyCallCount'
execution_id: hue-dry-run-exec-014-add-dry-run-switch
dark-factory-version: dev
created: "2026-09-07T22:00:00Z"
queued: "2026-09-08T05:55:56Z"
started: "2026-09-08T05:55:58Z"
completed: "2026-09-08T05:58:41Z"
---

# feat: add DRY_RUN flag skipping bridge apply calls

<summary>
- Operators can run the hue controller in a mode where it computes the full desired light state but never changes a real light
- A new `DRY_RUN` env var / `-dry-run` CLI flag controls the mode, default off so existing deploys behave identically
- With dry-run on, every scheduled check still evaluates whether it is satisfied, and the controller logs what it would have changed
- With dry-run on, the bridge state-changing call is skipped entirely
- With dry-run off, behavior is exactly what it is today
- This makes a dev deploy against the live bridge safe: the whole reconcile pipeline runs, but the house lights are never touched
- A new test exercises both branches, asserting the state-changing call happens exactly once when off and zero times when on
- The deploy manifest gains a `DRY_RUN` entry pinned to `"false"` so production stays live unless explicitly flipped
- CHANGELOG gets a new `## Unreleased` bullet
</summary>

<objective>
Add a boolean `DRY_RUN` env var / `-dry-run` CLI flag that makes the checks runner log each unsatisfied check's intended action instead of executing it, so the controller can be deployed to a dev environment pointing at the real bridge without ever flipping a real light.
</objective>

<context>
Read first (the prompt depends on these being read before any edit):

- `./CLAUDE.md` — project conventions: `make precommit` is the gate, `docs/dod.md` is the validation prompt, Ginkgo v2 / Gomega test convention, factory functions are pure composition (no conditionals, no I/O).
- `./docs/dod.md` — DoD: Code Quality / Testing / Documentation. Note specifically "Factory functions are pure composition — no conditionals, no I/O" and "Error handling uses `github.com/bborbe/errors` with context wrapping".
- `./.dark-factory.yaml` — `workflow: direct`, `pr: false`, `autoRelease: false`, `validationPrompt: docs/dod.md`. The daemon handles the commit; do NOT commit, do NOT push, do NOT open a PR.
- `./main.go` — the `application` struct shows the libargument tag pattern. `SummerMode bool` (the last field) is the exact precedent for a boolean deploy-time toggle: `required:"false"`, `arg:`, `env:`, `usage:`, `default:"false"`. `Run` passes `a.SummerMode` positionally into `factory.CreateCheckController`.
- `./pkg/factory/factory.go` — `CreateCheckController` currently takes 9 parameters (`url`, `id`, `token`, `inverval`, `summerMode`, `currentDateTimeGetter`, `location`, `sunriseSunsetProvider`, `samplerFactory`). It is pure composition and must stay free of conditionals. It calls `check.NewChecksRunner(currentDateTimeGetter)`.
- `./pkg/check/checks-runner.go` — THE branch point. `RunChecks(ctx, checks CheckList) error` loops the checks; for each it calls `check.Satisfied(ctx)`, skips when satisfied, and otherwise logs `"check not satisfied, apply"` then calls `check.Apply(ctx)` and logs `"check applied"`. The dry-run branch replaces the `Apply` call and its surrounding log lines.
- `./pkg/check/check.go` — the `Check` interface (`Apply` / `Satisfied` / `Name`) and `CheckList`. Unchanged by this prompt. Note the `//counterfeiter:generate` annotation that produces `mocks.Check`.
- `./pkg/check/checks-cron_test.go` — the Ginkgo v2 exemplar to follow for the new test. It demonstrates: external `package check_test`, mock construction from `./mocks`, and — most importantly — the `bytes.Buffer` + `slog.SetDefault` / `AfterEach` restore pattern for asserting on emitted log lines. Copy that logging-capture shape.
- `./pkg/check/check_suite_test.go` — the suite bootstrap; do not modify.
- `./mocks/check.go` — the generated `mocks.Check` fake. It exposes `ApplyCallCount()`, `SatisfiedReturns(bool, error)`, and `NameReturns(string)`. `ApplyCallCount()` is the assertion the new test hangs on.
- `./k8s/hue-deploy.yaml` — the `env:` list. `LISTEN` and the commented `SUMMER_MODE` block are the simple-value entries; the new entry goes alongside them, before the `secretKeyRef` entries.
- `./CHANGELOG.md` — the `## Unreleased` section convention.

Naming: this repo's env vars are unprefixed (`LISTEN`, `URL`, `ID`, `TOKEN`, `SUMMER_MODE`). Use `DRY_RUN`, NOT `HUE_DRY_RUN`, to match.
</context>

<requirements>

### Step 1 — Add the `DryRun` field to the `application` struct (`main.go`)

1. In `./main.go`, add a new field to the `application` struct immediately after the existing `SummerMode` field (keeping it the last entry), following the same tag layout as its neighbours:

   ```go
   DryRun bool `required:"false" arg:"dry-run" env:"DRY_RUN" usage:"Compute and log the desired light state without changing any light" default:"false"`
   ```

   Match the surrounding struct's tag column alignment by hand — `gofmt` does not align struct tags for you.

2. In `Run`, pass `a.DryRun` into `factory.CreateCheckController` as a new argument positioned immediately after `a.SummerMode`, so the two boolean deploy-time toggles stay adjacent.

### Step 2 — Thread the flag through the factory (`pkg/factory/factory.go`)

3. Add a `dryRun bool` parameter to `CreateCheckController`, positioned immediately after `summerMode bool` to match the call site from step 2.

4. Pass `dryRun` into `check.NewChecksRunner(...)` as a new argument. `CreateCheckController` must remain pure composition — do NOT add an `if` in this function; the parameter is passed straight through.

5. Confirm the only call site is `main.go`: `grep -rn 'CreateCheckController' main.go pkg/` should show exactly the definition in `pkg/factory/factory.go` and the one call in `main.go`. If a third match appears, STOP and report.

### Step 3 — Branch in the checks runner (`pkg/check/checks-runner.go`)

6. Add a `dryRun bool` parameter to `NewChecksRunner`, positioned after `currentDateTimeGetter`, and store it as a new `dryRun bool` field on the `checksRunner` struct.

7. In `RunChecks`, in the not-satisfied path (currently: log `"check not satisfied, apply"`, call `check.Apply(ctx)`, log `"check applied"`), branch on `c.dryRun`:

   - When `c.dryRun` is true: emit ONE log line at Info level with the message `dry-run, skip apply` and a `check` attribute carrying `check.Name()`, then continue to the next check. `check.Apply(ctx)` MUST NOT be called.
   - When `c.dryRun` is false: the existing behavior is unchanged — same two Debug lines, same `Apply` call, same `errors.Wrapf` on failure.

   Use Info (not Debug) for the dry-run line: it is the operator-facing evidence that dry-run is working, and it should surface without depending on the deployment's log level.

8. Do NOT change the `ChecksRunner` interface — it stays `RunChecks(ctx context.Context, checks CheckList) error`. The flag is captured in the constructor, not passed per-call.

9. Do NOT change the `Satisfied` evaluation, the `ctx.Done()` select, or the satisfied-skip path. Dry-run gates the write, not the read: a satisfied check still logs `"check satisfied, skip"` exactly as today.

### Step 4 — Add the env var to the deploy manifest (`k8s/hue-deploy.yaml`)

10. In the `env:` list, add a new entry directly after the existing `SUMMER_MODE` block and before the `SENTRY_DSN` entry, with a comment in the same style as the `SUMMER_MODE` comment above it:

    ```yaml
            # DRY_RUN makes the controller compute and log the desired light
            # state without changing any light. Set to "true" for a dev deploy
            # that exercises the full reconcile pipeline against the real
            # bridge without touching the house.
            - name: DRY_RUN
              value: "false"
    ```

    Match the surrounding indentation exactly (entries are nested under `env:` inside the container spec).

### Step 5 — Add a test covering BOTH branches (`pkg/check/checks-runner_test.go`)

11. Create `./pkg/check/checks-runner_test.go` — this file does NOT exist yet; `pkg/check` currently has no test for the runner. Use external package `check_test`, Ginkgo v2 / Gomega, and the same 3-line license header style as `pkg/check/checks-cron_test.go`.

12. Follow the log-capture pattern from `pkg/check/checks-cron_test.go`: a `bytes.Buffer`, `slog.SetDefault` with a `TextHandler` at `slog.LevelDebug` in `BeforeEach`, and restore of the previous default in `AfterEach`.

13. Build the fixture from `mocks.Check`: one fake whose `SatisfiedReturns(false, nil)` and `NameReturns("test-check")`, wrapped in a `check.CheckList`. An unsatisfied check is what drives the code into the branch under test.

14. Write at minimum these assertions, as two `Context` blocks or a `DescribeTable` — both branches are mandatory; a single-branch test does not satisfy this step:

    - **dry-run enabled** (`check.NewChecksRunner(currentDateTimeGetter, true)`): after `RunChecks(ctx, checks)` returns nil, `Expect(fakeCheck.ApplyCallCount()).To(Equal(0))` AND the captured log buffer contains `dry-run, skip apply`.
    - **dry-run disabled** (`check.NewChecksRunner(currentDateTimeGetter, false)`): after `RunChecks(ctx, checks)` returns nil, `Expect(fakeCheck.ApplyCallCount()).To(Equal(1))` AND the captured log buffer does NOT contain `dry-run, skip apply`.

    `ApplyCallCount()` is the load-bearing assertion — it is the only thing that distinguishes a real implementation from one that wires the flag but leaves `Apply` unconditional. Do not replace it with a log-only assertion.

15. For the `currentDateTimeGetter` argument, use whatever the repo's existing tests use for a `libtime.CurrentDateTimeGetter`. Check `pkg/check/` and `pkg/handler/` test files for an existing fixture or fake before inventing one; `libtime.NewCurrentDateTime()` is acceptable if no test fixture exists, since the runner does not read the clock in the path under test.

### Step 6 — CHANGELOG

16. In `./CHANGELOG.md`, add a `## Unreleased` section at the top of the version list (directly above `## v0.6.0`) if one does not already exist, with a single bullet:

    ```
    - feat: Add `DRY_RUN` env var / `-dry-run` CLI flag that makes the checks runner log each unsatisfied check's intended action at Info level instead of calling `Apply`, so the controller can run against a real bridge without changing any light. Defaults to `false`; `k8s/hue-deploy.yaml` pins it to `"false"` so production behavior is unchanged.
    ```

    If a `## Unreleased` section already exists, append the bullet to it and leave existing bullets untouched.

### Step 7 — Verification gate

17. Run `make precommit` from the repo root. It must pass. If it fails for a reason unrelated to this change, STOP and report rather than fixing unrelated breakage.

18. Before you finish, re-run every command in `<verification>` and confirm each one passes, then walk every requirement above against the actual diff.

</requirements>

<constraints>
- Do NOT commit, do NOT push, do NOT open a PR — the dark-factory daemon handles git per `.dark-factory.yaml` (`workflow: direct`, `pr: false`, `autoRelease: false`).
- Do NOT run `git commit --amend` or `git rebase`.
- Do NOT add a conditional to `pkg/factory/factory.go` — factory functions are pure composition per `docs/dod.md`. The flag passes straight through.
- Do NOT change the `ChecksRunner` or `Check` interfaces.
- Do NOT change the satisfied-skip path or the `ctx.Done()` cancellation select in `RunChecks`.
- Use `github.com/bborbe/errors` for any error wrapping — never `fmt.Errorf`, never a bare `return err`.
- Use `log/slog` structured attributes for the new log line — no `fmt.Printf`, no string concatenation into the message.
- Do NOT run `go mod vendor` or `go mod tidy` — no dependency surface changes.
- Do NOT edit the `LICENSE` file or the license header (first 3 lines) of any Go file; new Go files MUST carry the standard 3-line header with the current year.
- Do NOT modify `pkg/check/checks-creator.go`, `pkg/check/checks-cron.go`, or any `pkg/handler/` file — the flag does not reach them.
- Do NOT reformat, reorder, or modernize code outside the insertions described above.
- Do NOT rename `SUMMER_MODE` or touch its wiring.
- `make precommit` MUST pass at the end.
</constraints>

<verification>
1. `make precommit` — must pass (ensure + format + generate + test + check + addlicense).
2. `grep -c 'DryRun' main.go` — must print `2`: the struct field and the `a.DryRun` argument.
3. `grep -c 'dryRun' pkg/factory/factory.go` — must print `2`: the parameter and the pass-through argument.
4. `! grep -qE '^	if |^		if ' pkg/factory/factory.go` — must succeed; the factory stays free of conditionals.
5. `grep -c 'dryRun' pkg/check/checks-runner.go` — must print at least `3`: the constructor parameter, the struct field, and the branch condition.
6. `grep -c 'DRY_RUN' k8s/hue-deploy.yaml` — must print `1`.
7. `ls pkg/check/checks-runner_test.go` — must succeed.
8. `grep -c 'ApplyCallCount' pkg/check/checks-runner_test.go` — must print at least `2` (one assertion per branch).
9. `grep -c 'NewChecksRunner' pkg/check/checks-runner_test.go` — must print at least `2`, one per branch, with different boolean arguments.
10. `grep -c 'Unreleased' CHANGELOG.md` — must print at least `1`.
</verification>
