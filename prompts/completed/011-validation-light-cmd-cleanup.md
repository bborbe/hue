---
status: completed
summary: Renamed slice adapters to LightList/CheckList, swapped TimeOfDay.Validate to bborbe/validation with contract tests, regenerated mocks, added CHANGELOG entry; make precommit passes
execution_id: hue-pr4-validation-cmd-cleanup-exec-011-validation-light-cmd-cleanup
dark-factory-version: dev
created: "2026-09-07T05:25:14Z"
queued: "2026-09-07T05:25:14Z"
started: "2026-09-07T05:25:38Z"
completed: "2026-09-07T05:27:52Z"
---
<summary>
- Rename the slice adapter types to the `XxxList` convention the `list-type-name` rule requires
- The light-slice sorting adapter follows the naming convention for slice-of-X types
- The check-aggregator type pairs with the `Check` interface it composes
- Generated fakes for the check-creator and check-runner are refreshed to match the renamed types
- The hand-written time-of-day nil-check is replaced by the shared validation framework
- Update every call-site and test reference so the build stays green with zero behavior change
- Add a `## Unreleased` CHANGELOG entry for the cleanup, per the project DoD
- `make precommit` passes; the `use-bborbe-validation-not-inline-checks` and `list-type-name` ast-grep findings drop to zero
</summary>

<objective>
Clear the last two code-review funnel categories for the hue repo: rename the two slice-adapter types per the `list-type-name` rule and replace the hand-rolled `TimeOfDay.Validate` body with `bborbe/validation`. End state: both ast-grep rules report zero findings and runtime behavior is unchanged.
</objective>

<context>
Read `CLAUDE.md` for project conventions first (use `github.com/bborbe/errors` for wrapping, never `fmt.Errorf`, no bare `return err`).

This is a pure mechanical cleanup — no behavior change. Three findings are being cleared:

1. `pkg/light.go` — `type Lights []huego.Light` (with `Len`/`Swap`/`Less`, a `sort.Interface` adapter). Its only call-site is `cmd/list-lights/main.go` (~line 47: `lights := pkg.Lights(hueLights)` followed by `sort.Sort(lights)`).
2. `pkg/check/check.go` — `type Checks []Check`. This is the functional-composition aggregator for the `Check` interface, so the rule requires `CheckList`. Call-sites: `pkg/check/checks-creator.go` (interface + implementation return type, composite literal), `pkg/check/checks-runner.go` (method param), tests `pkg/check/checks-parity_test.go` and `pkg/check/checks-creator_test.go`, and the counterfeiter-generated mocks `mocks/checks-creator.go` / `mocks/checks-runner.go` (regenerate, do not hand-edit).
3. `pkg/time-of-day.go` — `func (t TimeOfDay) Validate(ctx context.Context) error` currently hand-rolls `if t.Location == nil { return errors.New(ctx, "location missing") }`. Convert to the `bborbe/validation` shape (`validation.All{...}` + `validation.Name(...)` + `validation.NotNil(...)`), following the exact code in requirement 8. `github.com/bborbe/validation` v1.4.23 is already in `go.mod` (currently `// indirect`, pulled via `github.com/bborbe/time`); it is promoted to a direct dependency when the import is added and `make precommit` runs `go mod tidy -e`.
</context>

<requirements>
1. `pkg/light.go`: rename the type `Lights` → `LightList` (declaration `type LightList []huego.Light` plus the `Len`/`Swap`/`Less` receiver methods). Keep the `sort.Interface` behavior byte-for-byte identical.
2. `cmd/list-lights/main.go`: update the call-site `pkg.Lights(hueLights)` → `pkg.LightList(hueLights)` (~line 47). Do NOT touch the `application` struct's `service.Main` flag tags (`required:"true"`/`arg:`/`env:`/`usage:`) — those are `service`'s own tag-based flag validation and stay as-is.
3. `pkg/check/check.go`: rename `type Checks []Check` → `type CheckList []Check`.
4. `pkg/check/checks-creator.go`: update the `CheckCreator` interface (`CreateChecks(ctx context.Context) (CheckList, error)`), the `*checkCreator.CreateChecks` implementation signature, and the composite literal `return Checks{...}` → `return CheckList{...}`.
5. `pkg/check/checks-runner.go`: update the `ChecksRunner` interface and the `*checksRunner.RunChecks` method — parameter `checks Checks` → `checks CheckList`.
6. Update test references: `pkg/check/checks-parity_test.go` (the `checks check.Checks` parameter ~line 28 and the `check.Checks{` composite literal ~line 75) and `pkg/check/checks-creator_test.go` (the `checkNamed(checks check.Checks, ...)` parameter ~line 28). Rename only the type references — do not change test logic.
7. Regenerate the counterfeiter mocks by running `make generate` (runs `go generate ./...`). This rewrites `mocks/checks-creator.go` and `mocks/checks-runner.go` to use `check.CheckList`. Never hand-edit the mocks.
8. `pkg/time-of-day.go`: replace the `Validate` body with:
   ```go
   func (t TimeOfDay) Validate(ctx context.Context) error {
       return validation.All{
           validation.Name("location", validation.NotNil(t.Location)),
       }.Validate(ctx)
   }
   ```
   Add the import `github.com/bborbe/validation`. Remove the now-unused `github.com/bborbe/errors` import from `pkg/time-of-day.go` if no other `errors.*` usage remains (there is none).

   Add a contract test to `pkg/time-of-day_test.go` exercising the real validator boundary — `TimeOfDay.Validate` currently has no call-sites and no test coverage. Add `"context"` to the test file's import block and append two `It` blocks inside the existing `Describe("Hue Turn On Light", ...)`:
   ```go
   It("fails validation on nil location", func() {
       Expect(pkg.TimeOfDay{}.Validate(context.Background())).To(HaveOccurred())
   })
   It("passes validation with a location", func() {
       Expect(pkg.TimeOfDay{Location: time.UTC}.Validate(context.Background())).NotTo(HaveOccurred())
   })
   ```
9. `CHANGELOG.md`: add a `## Unreleased` section directly above the highest `## vX.Y.Z` (currently `## v0.4.2`; there is no `## Unreleased` on this branch yet) with a single `refactor:` bullet summarizing the rename + validation-framework swap. Required by the project DoD (`docs/dod.md`: "CHANGELOG.md has an entry under `## Unreleased`"). Bullets use conventional prefixes (`feat:` / `fix:` / `refactor:` / `test:` / `chore:`).
10. Self-check before finishing: run `grep -rnE 'type (Lights|Checks)\b|check\.Checks\b|pkg\.Lights\b' pkg/ cmd/ main.go mocks/` — it must print nothing. Walk each numbered requirement against the final diff.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git.
- Existing tests must still pass; regenerate mocks (`make generate`) before `make test`.
- This is a rename + validation-framework swap only. No sort-order, schedule, or rejection-semantics change: a nil `Location` must still fail `Validate` with the same class of rejection (the message text changes to the `bborbe/validation` format — that is the point of adopting the framework).
- Follow project conventions: use `github.com/bborbe/errors` where errors are still wrapped; no `fmt.Errorf`.
- Do not modify any `pkg/check/checks-*.go` logic beyond the `Checks` → `CheckList` type references.
</constraints>

<verification>
Run `make precommit` — must pass.
Verify the absence assertions (use `! grep -q ...` to assert):
- `! grep -qE 'type (Lights|Checks)\b|check\.Checks\b|pkg\.Lights\b' pkg/ cmd/ main.go mocks/` — must exit 0 (no matches).
- `grep -q "^## Unreleased" CHANGELOG.md` — must succeed, with the new `refactor:` bullet directly beneath it.
</verification>
