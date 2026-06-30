---
status: completed
summary: Restored aquarium light schedule to daytime window (10:00-20:00) by replacing the temporary heat-wave override with original daytime values
execution_id: hue-revert-aquarium-heat-wave-override-exec-001-revert-aquarium-heat-wave-override
dark-factory-version: v0.188.1
created: "2026-06-29T18:30:00Z"
queued: "2026-06-29T16:45:29Z"
started: "2026-06-29T16:45:54Z"
completed: "2026-06-29T16:47:02Z"
---

<summary>
- Restore the aquarium light schedule to its pre-heat-wave daytime window
- Switch the active values back to the commented-out daytime constants
- Remove the temporary heat-wave evening-only override comment block
- Leave the handler test files added in the same commit untouched
- Leave the status.go ctx.Err() fast-fail added in the same commit untouched
- No other schedule values (artemia, CO2, skimmer, jana) are touched
- Aquarium light ON hour returns to 10:00
- Aquarium light OFF hour returns to 20:00 (on + 10)
- All four aquarium-tagged checks (Aquarium Licht, Aquarium Rack, CO2 base, Garnelen Licht 1/2, Jana Aqua Light) follow the same restored window via the shared variable
- Project still passes `make precommit` after the edit; CHANGELOG entry recorded per DoD
</summary>

<objective>
Restore the long-running daytime aquarium-light schedule in `pkg/check/checks-creator.go` by undoing the temporary heat-wave evening-only override (commit `607c454`) without touching the unrelated handler-test and status-fast-fail changes that landed in the same commit. After the change the aquarium light follows the on=10 / off=on+10 daytime window again.
</objective>

<context>
- Read `~/Documents/workspaces/hue/CLAUDE.md` — hue uses `make precommit` as the gate and `docs/dod.md` as the validation prompt.
- Read `~/Documents/workspaces/hue/docs/dod.md` — Definition of Done the agent self-checks against after implementation.
- Read `~/Documents/workspaces/hue/pkg/check/checks-creator.go` lines 45-60 — the exact block to edit. Confirm the active override (`aquariumLightOnHour := 20` and `aquariumLightOffhour := aquariumLightOnHour + 3`) and the commented-out daytime target (`aquariumLightOnHour := 10` and `aquariumLightOffhour := aquariumLightOnHour + 10`).
- Read `~/Documents/workspaces/hue/Makefile` to confirm `make precommit` runs ensure + format + generate + test + check + addlicense.
- Target revert commit: `607c454` ("review: add handler tests, ctx fast-fail, document temporary aquarium window", Jun 21 2026). Run `git show 607c454 -- pkg/check/checks-creator.go` to see the exact original diff.
- The code at lines 48-53 must end up as the original daytime version: keep the variable lines uncommented and active with the daytime values; drop the override comment that frames them as a temporary heat-wave window.

Pattern references:
- `pkg/check/checks-creator.go` (the file being edited) — established style uses `errors.Wrap(ctx, err, "...")` from `github.com/bborbe/errors`. No new wrappers are introduced by this revert, but keep the existing wrap calls intact at lines 37, 41, 68.
- `pkg/check/*_test.go` (e.g. `pkg/check/alternate-switch_test.go`) — uses Ginkgo v2 / Gomega; new tests are not required here (pure mechanical revert), but the agent MUST NOT remove or skip existing tests in this package during the edit.

`docs/dod.md` (DoD) applies — the final `make precommit` is the gate; the agent self-checks Code Quality / Testing / Documentation sections before reporting done.
</context>

<requirements>
1. Open `pkg/check/checks-creator.go` and locate the block immediately under the `glog.V(2).Infof("current time ...")` line (the `// Temporary heat-wave ...` comment plus the four assignment lines).
2. Find-and-replace the current block (lines 48-53) — the temporary override — with the daytime values, removing the override comment entirely:
   - REMOVE these lines verbatim:
     ```go
     // Temporary heat-wave evening-only window — operator will revert to the
     // daytime values below by hand (~1-2 weeks) once the extreme heat is gone.
     //aquariumLightOnHour := 10
     //aquariumLightOffhour := aquariumLightOnHour + 10
     aquariumLightOnHour := 20
     aquariumLightOffhour := aquariumLightOnHour + 3
     ```
   - REPLACE them with these four lines (no surrounding comment block):
     ```go
     aquariumLightOnHour := 10
     aquariumLightOffhour := aquariumLightOnHour + 10
     ```
3. Verify the surrounding context after the edit:
   - `co2OnHour := aquariumLightOnHour - 2` and `co2OffHour := aquariumLightOffhour - 2` stay unchanged on disk (their evaluated values shift back to 8 / 18 — that is expected, do not edit the lines).
   - `artemiaLightOnHour := 8` and `artemiaLightOffhour := 23` are NOT touched.
   - The six `NewBetweenTimeSwitch` blocks below that consume `aquariumLightOnHour` / `aquariumLightOffhour` (Aquarium Licht, Aquarium Rack, Aquarium CO2, Garnelen Licht 1, Garnelen Licht 2, Jana Aqua Light) automatically follow the restored daytime window via the shared variable. Do NOT edit them.
   - The two `NewAlternateSwitch` blocks for Aquarium Skimmer and Jana Aqua Skimmer (5*time.Minute on, 25*time.Minute off) are unaffected and stay as-is.
4. Add a CHANGELOG.md entry under `## Unreleased` (DoD requirement). One line, e.g. `Restore aquarium light schedule to daytime window (10:00-20:00) — revert of temporary heat-wave override from #N`. If a PR number is not yet known, write `from <commit-SHA>` instead. Keep prior `## Unreleased` bullets intact.
5. Do NOT modify any file other than `pkg/check/checks-creator.go` and `CHANGELOG.md`.
6. Do NOT add new tests, regenerate mocks (`make generate`), bump dependencies, or edit other docs — this is a configuration-value revert with no business-logic surface change. Existing tests should already cover the schedule construction; if `make precommit` fails after the edit for a reason unrelated to the variable change, STOP and report instead of working around.
7. Do NOT delete the existing copyright header at the top of the file (lines 1-3).
</requirements>

<scope_out>
The following files are part of commit `607c454` but are NOT being reverted — explicitly do NOT touch them:
- `pkg/handler/list-lights_test.go` — added in the same commit; stays as-is.
- `pkg/handler/status_test.go` — added in the same commit; stays as-is.
- `pkg/handler/status.go` — pre-loop `ctx.Err()` fast-fail check added in the same commit; stays as-is.

If `git show 607c454 --stat` lists them, they are out of scope for this revert. Only `pkg/check/checks-creator.go` (plus `CHANGELOG.md` for DoD compliance) is in scope.
</scope_out>

<constraints>
- Do NOT commit — dark-factory handles git (`worktree: false` per `.dark-factory.yaml`).
- Do NOT push, do NOT open a PR; `pr: false` and `autoRelease: false` per `.dark-factory.yaml` — the daemon will handle the commit/PR once the prompt completes.
- Do NOT use `git -C` or `cd /path && git ...` patterns; follow `[[Development Guide]]` git-workflow rules.
- Do NOT edit any file other than `pkg/check/checks-creator.go` and `CHANGELOG.md`.
- Do NOT modify the file license header (lines 1-3 of `pkg/check/checks-creator.go`).
- Do NOT reformat / gofmt / reorder / modernize any other code in the file; perform the literal find-and-replace only.
- Do NOT add a new test file or new test case; existing tests must continue to pass unchanged.
- Do NOT bump `go.mod` / `go.sum` or run `go mod tidy`; no dependency surface change.
- `make precommit` MUST pass at the end (DoD per `docs/dod.md`: ensure + format + generate + test + check + addlicense).
</constraints>

<verification>
1. `git diff --stat` — must list exactly `pkg/check/checks-creator.go` and `CHANGELOG.md`; no other file changed. The daemon runs from the repo root, no `cd` needed.
3. `git diff pkg/check/checks-creator.go` — verify the override comment is gone and the active assignments are the daytime values:
   - `aquariumLightOnHour := 10` is present (uncommented).
   - `aquariumLightOffhour := aquariumLightOnHour + 10` is present (uncommented).
   - `aquariumLightOnHour := 20` and `aquariumLightOffhour := aquariumLightOnHour + 3` are NOT present.
   - The `// Temporary heat-wave ...` comment is NOT present.
   - No other lines in the file are changed (co2OnHour / co2OffHour / artemia / switch blocks all unchanged on disk).
4. `rg -n 'aquariumLightOnHour := ' pkg/check/checks-creator.go` — must show exactly the `:= 10` assignment and no `:= 20` assignment.
5. `grep -A1 '^## Unreleased' CHANGELOG.md` — confirm the new bullet describing the schedule restore is present under the `## Unreleased` header created by requirement #4, and that prior bullets are preserved.
6. `make precommit` — must pass (this is the gate per `docs/dod.md`).
7. If any other file appears in `git status`, STOP and report — the scope was exceeded.
</verification>
