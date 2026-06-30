---
status: completed
summary: Added HUE_SUMMER_MODE boolean flag that toggles aquarium light schedule between daytime (10:00-20:00) and evening-only (20:00-23:00) windows
execution_id: hue-revert-aquarium-heat-wave-override-exec-002-add-summer-mode-boolean-flag
dark-factory-version: v0.188.1
created: "2026-06-29T18:55:00Z"
queued: "2026-06-29T19:46:10Z"
started: "2026-06-29T19:46:26Z"
completed: "2026-06-29T19:48:29Z"
---

<summary>
- Operators can toggle the aquarium/CO2/Garnelen/Jana light schedule between daytime (10:00-20:00) and evening-only (20:00-23:00) at deploy time
- A new `HUE_SUMMER_MODE` env var / `-summer-mode` CLI flag controls which window is active
- Default is off (daytime) so existing deploys behave identically to today
- Cascading via the shared `aquariumLightOnHour` / `aquariumLightOffhour` variables means no per-light branching is needed
- Artemia, Skimmer, CO2 base, and Jana Aqua Skimmer schedules are unaffected — only the aquarium cluster toggles
- Existing tests stay green; a new test exercises BOTH branches of the toggle
- `k8s/hue-deploy.yaml` gets a `SUMMER_MODE` env var pinned to `"false"` so production deploys stay on daytime unless explicitly flipped
- CHANGELOG gets a new `## Unreleased` bullet describing the feature
- All edits land as a NEW commit on top of the existing revert commit (`028bb9b`) — never amend, never rebase

</summary>

<objective>
Add a boolean `HUE_SUMMER_MODE` env var / `-summer-mode` CLI flag that selects between the daytime (`10:00-20:00`) and evening-only (`20:00-23:00`) aquarium light window. The plumbing extends three signatures by one parameter each (`application` field → `factory.CreateCheckController` → `check.NewCheckCreator`), the `CreateChecks` body branches on the flag, a new test covers both windows, and the deploy manifest gains a `SUMMER_MODE` env var pinned to `"false"` so existing prod behavior is preserved.

</objective>

<context>
Read first (the prompt depends on these being read before any edit):

- `./CLAUDE.md` — project conventions: `make precommit` is the gate, `docs/dod.md` is the validation prompt, Ginkgo v2 / Gomega test convention, factory functions are pure composition.
- `./docs/dod.md` — DoD: Code Quality / Testing / Documentation; agent self-checks after `make precommit`.
- `./.dark-factory.yaml` — `worktree: false`, `pr: false`, `autoRelease: false`, `validationPrompt: docs/dod.md`. Daemon handles commit; agent MUST NOT run `git commit --amend`, MUST NOT push, MUST NOT open a PR.
- `./main.go` — current `application` struct (lines 32-43) shows the libargument tag pattern with `default:"..."` and `arg:` / `env:` / `usage:`; `Run` (lines 45-61) shows how `a.Inverval` is passed to `factory.CreateCheckController`. Insertion point for the new `SummerMode bool` field: end of the `application` struct (after `BuildDate *libtime.DateTime`, line 42), and pass it as a new LAST argument to `factory.CreateCheckController` at line 53.
- `./pkg/factory/factory.go` — current `CreateCheckController(url, id, token, inverval time.Duration) run.Func` (lines 20-37); must gain a 5th `summerMode bool` parameter that flows into `check.NewCheckCreator(provider, summerMode)`. Note that `factory_suite_test.go` does NOT exercise `CreateCheckController` (only TestSuite bootstrap), so no test signature ripple there.
- `./pkg/check/checks-creator.go` — current day-time version (lines 48-53 already restored to `aquariumLightOnHour := 10` / `+10` by the prior revert prompt). Defines `CheckCreator` interface (line 18), `NewCheckCreator` constructor (line 22), `checkCreator` struct (line 29), `CreateChecks` method (line 34). The branching happens here: replace the two-line day-time block with a 4-line `if c.summerMode { ... } else { ... }` that picks `20/+3` vs `10/+10`. The 6 downstream `NewBetweenTimeSwitch` blocks (Aquarium Licht / Rack / CO2 / Garnelen 1 / Garnelen 2 / Jana Aqua Light) follow the shared variable unchanged. Do NOT touch the `NewAlternateSwitch` blocks (Skimmer + Jana Aqua Skimmer) or the Artemia + CO2 base lines.
- `./pkg/check/checks-cron.go` — `NewCheckCron(creator, runner, interval)` signature is UNCHANGED (the flag is captured inside the creator, not passed to the cron).
- `./pkg/check/checks-runner.go` — `ChecksRunner` consumer signature is UNCHANGED. No signature ripple on consumers.
- `./pkg/check/check.go` — `Checks` is `[]Check` (a flat slice, no exported accessor for the hours). The hours passed to `NewBetweenTimeSwitch` are captured inside an unexported closure and NOT exposed on the returned `Check` interface. Test assertions CANNOT read `pkg.TimeOfDay.Hour` from the constructed `Checks` directly — see Step 6 for the feasible substitute test pattern.
- `./pkg/check/alternate-switch_test.go` — Ginkgo v2 / Gomega exemplar for the new test. `Describe` → nested `Context` blocks; `BeforeEach` sets up fixtures; `It` assertions use `Expect(...).To(Equal(...))`. External test package (`package check_test`).
- `./pkg/handler/status_test.go` (lines 30-65) — shows how to inject a fake `pkg.BridgesProviderFunc` for tests that need a `BridgesProvider`. The new `pkg/check/checks-creator_test.go` will use the same pattern: `pkg.BridgesProviderFunc(func(ctx) ([]*huego.Bridge, error) { return []*huego.Bridge{nil}, nil })` so the bridge-skip logic doesn't panic.
- `./k8s/hue-deploy.yaml` — current `env:` list (lines 50-72). Add a `SUMMER_MODE` entry to the existing list with `value: "false"`, place it alphabetically next to the other simple-value env vars (after `LISTEN`).

Base commit on branch `feature/revert-aquarium-heat-wave-override`: `028bb9b` (the revert commit, already on top of `origin/master` `7791b54`). The new commit lands on top of `028bb9b` — DO NOT amend, DO NOT rebase, DO NOT push. Final state must show exactly 2 commits on the branch via `git log --oneline -2`.

</context>

<requirements>

### Step 0 — Confirm worktree state

1. Run `git log --oneline -3` from the repo root. Expected output:
   ```
   028bb9b fix: revert aquarium heat-wave override, restore daytime schedule
   7791b54 release v0.1.0
   9b9f85d Merge pull request #6 from bborbe/chore/build-info-pattern
   ```
   If the worktree is not on `028bb9b`, STOP and report.

2. Run `git status --short` — must be clean. If dirty, STOP and report.

### Step 1 — Add `SummerMode` field on the `application` struct (`main.go`)

3. In `./main.go`, add a new field as the LAST entry of the `application` struct (after `BuildDate *libtime.DateTime` on line 42):
   ```go
   SummerMode      bool              `required:"false" arg:"summer-mode"       env:"SUMMER_MODE"       usage:"Use the summer (evening-only) aquarium light window"  default:"false"`
   ```
   Note: `default:"false"` is the canonical libargument tag for a boolean default. The `required:"false"` matches every other non-required field in the struct.

4. In `Run` (line 45), pass the new flag as the LAST argument to `factory.CreateCheckController` so existing parameter order is preserved:
   ```go
   factory.CreateCheckController(
       a.Url,
       a.ID,
       a.Token,
       a.Inverval,
       a.SummerMode,
   ),
   ```

### Step 2 — Extend `CreateCheckController` signature (`pkg/factory/factory.go`)

5. In `./pkg/factory/factory.go`, change `CreateCheckController` (line 20) to accept a 5th parameter `summerMode bool` as the LAST positional argument. After the change:
   ```go
   func CreateCheckController(
       url string,
       id string,
       token pkg.Token,
       inverval time.Duration,
       summerMode bool,
   ) run.Func {
       return check.NewCheckCron(
           check.NewCheckCreator(
               CreateBridgesProvider(
                   url,
                   id,
                   token,
               ),
               summerMode,
           ),
           check.NewChecksRunner(),
           inverval,
       )
   }
   ```
   The only call site is `main.go:53` (already updated in step 1.4). No other callers — `grep -rn 'factory.CreateCheckController' pkg/ main.go` should show ONLY `main.go` and the factory definition.

### Step 3 — Extend `NewCheckCreator` and branch in `CreateChecks` (`pkg/check/checks-creator.go`)

6. In `./pkg/check/checks-creator.go`:
   - Change `NewCheckCreator` (line 22) to accept a 2nd parameter `summerMode bool`:
     ```go
     func NewCheckCreator(provider pkg.BridgesProvider, summerMode bool) CheckCreator {
         return &checkCreator{
             provider:    provider,
             summerMode:  summerMode,
             location:    "Europe/Berlin",
         }
     }
     ```
   - Add `summerMode bool` as a new field on the `checkCreator` struct (line 29), placed next to `provider` for clarity:
     ```go
     type checkCreator struct {
         provider   pkg.BridgesProvider
         summerMode bool
         location   string
     }
     ```
   - In `CreateChecks` (line 34), replace the existing 4-line block at lines 48-53:
     ```go
     // Temporary heat-wave evening-only window — operator will revert to the
     // daytime values below by hand (~1-2 weeks) once the extreme heat is gone.
     //aquariumLightOnHour := 10
     //aquariumLightOffhour := aquariumLightOnHour + 10
     aquariumLightOnHour := 20
     aquariumLightOffhour := aquariumLightOnHour + 3
     ```
     with:
     ```go
     var aquariumLightOnHour int
     var aquariumLightOffhour int
     if c.summerMode {
         // Evening-only window for sustained heat.
         aquariumLightOnHour = 20
         aquariumLightOffhour = aquariumLightOnHour + 3
     } else {
         // Standard daytime window.
         aquariumLightOnHour = 10
         aquariumLightOffhour = aquariumLightOnHour + 10
     }
     ```
     The block must be a clean swap — no leftover commented-out lines, no comment header explaining the toggle (the usage string on the flag already documents it).

7. Do NOT touch:
   - `co2OnHour := aquariumLightOnHour - 2` / `co2OffHour := aquariumLightOffhour - 2` — their evaluated values shift with the flag, the line is unchanged.
   - `artemiaLightOnHour := 8` / `artemiaLightOffhour := 23` — unchanged.
   - The 6 `NewBetweenTimeSwitch` blocks (Aquarium Licht / Rack / CO2 / Garnelen 1 / Garnelen 2 / Jana Aqua Light) — they consume the shared variable unchanged.
   - The 2 `NewAlternateSwitch` blocks (Skimmer + Jana Aqua Skimmer) — unchanged.

8. Do NOT change the `CheckCreator` interface (line 18). It stays `CreateChecks(ctx) (Checks, error)`.

### Step 4 — Add `SUMMER_MODE` env var to `k8s/hue-deploy.yaml`

9. In `./k8s/hue-deploy.yaml`, in the `env:` list (lines 50-72), add a new entry directly after the `LISTEN` block (which is the only other simple-value env var):
   ```yaml
               - name: SUMMER_MODE
                 value: "false"
   ```
   Pinned to `"false"` so production deploys stay on the daytime window unless an operator explicitly flips it. Add a comment above the entry explaining the toggle:
   ```yaml
               # SUMMER_MODE toggles the aquarium light schedule between the
               # standard daytime window (10:00-20:00, default) and the
               # evening-only window (20:00-23:00, set to "true" for heat waves).
               - name: SUMMER_MODE
                 value: "false"
   ```

### Step 5 — Add CHANGELOG entry under `## Unreleased`

10. In `./CHANGELOG.md`, the file currently has a single bullet under `## Unreleased` (added by the prior revert prompt `001-`). Add a new bullet below it — do NOT modify the existing one. The new bullet should describe the feature:
    ```
    - feat: Add `HUE_SUMMER_MODE` env var / `-summer-mode` CLI flag toggling the aquarium light schedule between the standard daytime window (10:00-20:00, default) and the evening-only window (20:00-23:00) for sustained-heat conditions. Defaults to `false` so existing deploys behave identically; flip to `"true"` in `k8s/hue-deploy.yaml` to activate the summer window. Cascades through the shared `aquariumLightOnHour` / `aquariumLightOffhour` variables — covers Aquarium Licht, Aquarium Rack, Aquarium CO2, Garnelen Licht 1/2, Jana Aqua Light. Artemia, Skimmer, CO2 base, Jana Aqua Skimmer schedules unaffected.
    ```

### Step 6 — Add a test exercising BOTH branches

11. Create a new file `./pkg/check/checks-creator_test.go` (does NOT exist yet — this is greenfield in this package). Use `package check_test` (external package), Ginkgo v2 / Gomega, with the same import + suite pattern as `pkg/check/alternate-switch_test.go`. Pattern reference: `pkg/check/handler/status_test.go:30-65` shows how to inject a fake `pkg.BridgesProviderFunc` returning one bridge (the test does NOT care about actual bridge state — only that the day/night branching produced the expected hours).

12. Structure of the new test:
    - One top-level `Describe("CheckCreator", func() { ... })`.
    - `BeforeEach` sets up `ctx := context.Background()` and a `provider` returning `[]*huego.Bridge{nil}, nil` (a single-element slice, matching how `CreateChecks` accesses `bridges[0]`).
    - Two `Context` blocks: `Context("summer mode disabled", func() { ... })` and `Context("summer mode enabled", func() { ... })`.
    - Each `Context` builds the creator via `check.NewCheckCreator(provider, <flag>)`, calls `creator.CreateChecks(ctx)`, and asserts that the returned `Checks` slice contains exactly the expected number of entries (9 — 6 between-time-switch for the aquarium cluster + 1 alternate-switch for Skimmer + 1 between-time-switch for Artemia + 1 alternate-switch for Jana Aqua Skimmer; count is stable across both branches).

13. The CANONICAL test pattern (per "Test the boundaries the new code crosses" rule): exercise the actual branching logic by calling `CreateChecks` directly on a `checkCreator` constructed with each flag value. Because `Checks` is `[]Check` with no exported hour accessor, the **test must use a boundary-crossing assertion rather than introspection**. Recommended shape:
    - One top-level `Describe("CheckCreator", func() { ... })`.
    - A `DescribeTable("aquarium window", func(summerMode bool, expectedEntryCount int) { ... }, Entry("summer mode disabled", false, 9), Entry("summer mode enabled", true, 9))` that calls `check.NewCheckCreator(provider, summerMode)`, then `creator.CreateChecks(ctx)`, and asserts `Expect(checks).To(HaveLen(expectedEntryCount))` AND `Expect(err).NotTo(HaveOccurred())`.
    - The 9-entry count is stable across both branches because the toggle only swaps the shared `aquariumLightOnHour` / `aquariumLightOffhour` values — it does not add or remove switches. The test still crosses the branching boundary (calls the code path with each flag value) and verifies the function returns successfully without error.
    - A separate `It` block asserting that `c.summerMode` round-trips through `NewCheckCreator`: construct a `checkCreator` directly (via the public constructor and reading the unexported field via a tiny accessor in `_test.go` if needed, OR by checking that the resulting `Checks` entries differ — if `TimeOfDay.Hour` could be exposed, this is the second-best check).

14. DO NOT attempt to introspect `pkg.TimeOfDay.Hour` from the constructed `Checks` slice — `BetweenTimeSwitch` captures the values in an unexported closure (`pkg/check/between-time-switch.go:17`) and the returned `Check` interface does NOT expose them. The `mocks.Check` only records call counts to `Apply(ctx)` / `Satisfied(ctx)` / `Name()`, not the `NewBetweenTimeSwitch` constructor args. There is no `NewBetweenTimeSwitchArgsForCall()` to inspect. Refactoring `BetweenTimeSwitch` to expose `From()` / `Until()` accessors is OUT OF SCOPE for this prompt.

15. Do NOT modify the existing `pkg/check/func.go` test surface — that file is for the `Check`-interface helpers and is unrelated.

### Step 7 — `pkg/factory/factory_suite_test.go`

16. Leave `pkg/factory/factory_suite_test.go` UNCHANGED — it does not call `CreateCheckController` (only TestSuite bootstrap). Confirm via `grep -n 'CreateCheckController' pkg/factory/factory_suite_test.go` — no output expected.

### Step 8 — Verification gate

17. Run `make precommit` from the repo root. Must pass per DoD. If it fails for a reason unrelated to this change, STOP and report.

### Step 9 — NEW commit (not amend)

18. After `make precommit` passes, `git add` the changed files:
    - `main.go`
    - `pkg/factory/factory.go`
    - `pkg/check/checks-creator.go`
    - `pkg/check/checks-creator_test.go` (new)
    - `k8s/hue-deploy.yaml`
    - `CHANGELOG.md`
    - DO NOT `git add` `.gitignore`, `prompts/completed/001-revert-aquarium-heat-wave-override.md`, or any file from the prior revert commit (these already exist in `028bb9b` and must not be re-staged).

19. `git commit -m "feat: add SUMMER_MODE flag toggling aquarium schedule"` (one-liner conventional prefix; the dark-factory daemon will append any standard trailers).

20. DO NOT run `git commit --amend`. DO NOT run `git push`. DO NOT open a PR. The daemon will handle these per `.dark-factory.yaml` (`pr: false`, `autoRelease: false` — a human pushes).

21. Final verification commands:
    - `git log --oneline -2` must show exactly 2 commits:
      ```
      <new-hash> feat: add SUMMER_MODE flag toggling aquarium schedule
      028bb9b   fix: revert aquarium heat-wave override, restore daytime schedule
      ```
    - `git diff 028bb9b..HEAD --stat` must list ONLY the 6 files from step 18 (NOT `.gitignore`, NOT `prompts/completed/001-revert-aquarium-heat-wave-override.md`, NOT `CHANGELOG.md` from the prior revert). The prior revert's `CHANGELOG.md` edit and the new `CHANGELOG.md` edit must appear as a single modified file in this diff (the new commit overwrites the prior content). Using the explicit base `028bb9b` rather than `HEAD~1` keeps the check correct even if master moves under us between prompt creation and execution.

</requirements>

<scope_out>
The following are EXPLICITLY out of scope — do NOT touch them:

- `pkg/handler/list-lights_test.go` — sibling change from `607c454`; stays as-is.
- `pkg/handler/status_test.go` — sibling change from `607c454`; stays as-is.
- `pkg/handler/status.go` — `ctx.Err()` fast-fail from `607c454`; stays as-is.
- `pkg/check/checks-cron.go` — `CheckCreator` interface unchanged; no signature ripple.
- `pkg/check/checks-runner.go` — `ChecksRunner` interface unchanged; no signature ripple.
- `pkg/check/checks-cron.go` — `NewCheckCron` takes the creator, not flags; signature unchanged.
- The 4 sibling files from `028bb9b` (`.gitignore`, `CHANGELOG.md` for the revert bullet, `pkg/check/checks-creator.go` revert portion, `prompts/completed/001-revert-aquarium-heat-wave-override.md`) — these are already correct and must not be re-touched or re-staged.
- The `LICENSE` file and the license header in any Go file (lines 1-3 of every `.go` file).
- `go.mod` / `go.sum` — no dependency surface change; do not run `go mod tidy`.
- Any other file in `pkg/`, `cmd/`, `pkg/trigger/`, `pkg/handler/`.

If a requirement above conflicts with this scope_out list, scope_out wins.

</scope_out>

<constraints>
- Do NOT run `git commit --amend` — the prompt's base is `028bb9b` and the new commit lands on top. Use `git commit -m "..."` only.
- Do NOT run `git rebase` — same reason. Branch order is fixed.
- Do NOT push, do NOT open a PR — `.dark-factory.yaml` has `pr: false` and `autoRelease: false`; a human pushes after reviewing the 2-commit branch.
- Do NOT use `git -C <path>` or `cd <path> && git ...` patterns — follow `[[Development Guide]]` git-workflow rules. Use a separate `cd` call followed by separate `git ...` calls when needed.
- Do NOT edit the `LICENSE` file or the license header (lines 1-3) of any Go file.
- Do NOT bump `go.mod` / `go.sum` or run `go mod tidy` — no new dependency is being introduced.
- Do NOT run `go mod vendor` — vendor is regenerated by `make buca`, not by the prompt.
- Do NOT add the new test in `pkg/check/func_test.go` or `pkg/check/func.go` (Func is for the Check-interface helper, unrelated).
- Do NOT touch the prior revert's `CHANGELOG.md` bullet — append a new bullet under the existing `## Unreleased` header, do not edit the existing one.
- Do NOT reformat / reorder / modernize code outside the literal insertions described above.
- `make precommit` MUST pass at the end (DoD per `docs/dod.md`: ensure + format + generate + test + check + addlicense).
- Final branch state must have exactly 2 commits on top of `origin/master` (`7791b54`) — `028bb9b` (revert) + the new feature commit.
- The agent MUST NOT number the prompt filename (`add-summer-mode-boolean-flag.md` — dark-factory assigns numbers on approve).
- The agent MUST NOT modify any frontmatter field beyond `status` and `created`.

</constraints>

<verification>
1. `git log --oneline -2` from the repo root — must show 2 commits with the new commit's subject line reading `feat: add SUMMER_MODE flag toggling aquarium schedule` (or similar conventional prefix). The HEAD subject MUST NOT match the `028bb9b` subject (`fix: revert aquarium heat-wave override, restore daytime schedule`) — that would mean `git commit --amend` was used by accident.
2. `git diff 028bb9b..HEAD --stat` — must list exactly 6 files for the new commit:
   ```
    CHANGELOG.md
    k8s/hue-deploy.yaml
    main.go
    pkg/check/checks-creator.go
    pkg/check/checks-creator_test.go
    pkg/factory/factory.go
   ```
   `.gitignore` and `prompts/completed/001-revert-aquarium-heat-wave-override.md` MUST NOT appear (those are from `028bb9b`).
3. `make precommit` — must pass (this is the DoD gate per `docs/dod.md`).
4. `git show HEAD:pkg/check/checks-creator.go` — must contain the `if c.summerMode { ... } else { ... }` block with the `20/+3` and `10/+10` windows. MUST NOT contain the `// Temporary heat-wave ...` comment header from the prior override. The value `20` may appear INSIDE the `if c.summerMode { ... }` branch (that's the summer window), but it MUST NOT appear at package scope or outside the if/else block.
5. `git show HEAD:pkg/check/checks-creator_test.go` — must exist and must contain BOTH the `summerMode=true` and `summerMode=false` contexts (or the equivalent table-test rows). A test that only exercises one branch fails the boundary-crossing rule.
6. `grep -c 'NewCheckCreator(' pkg/check/checks-creator.go` — must show exactly 1 match (the function definition); the call site is in `pkg/factory/factory.go`.
7. `grep -n 'SummerMode' main.go` — must show exactly 2 matches: the struct field declaration and the `a.SummerMode` argument in `Run`.
8. `grep -n 'SUMMER_MODE' k8s/hue-deploy.yaml` — must show the new env entry.
9. If any other file appears in `git status` after the verification commands, STOP and report — the scope was exceeded.

</verification>