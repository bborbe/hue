---
status: approved
spec: [002-migrate-glog-slog]
created: "2026-09-06T20:55:00Z"
queued: "2026-09-06T19:18:14Z"
---

# Logger setup per binary + `/setloglevel` conversion to slog

<summary>
- The controller installs a default text-to-stderr slog handler at its composition root with a `slog.LevelVar` initialized to `slog.LevelDebug`, so the deployed controller keeps emitting the same lines the `LOGLEVEL=2` deployment emits today (output parity)
- The three slog-emitting CLIs (list-lights, turnon-light, turnoff-light) each install a text-to-stderr slog handler defaulting to `slog.LevelInfo`
- The `/setloglevel/{level}` route keeps its path and operational purpose but is converted from the glog-typed `log.NewSetLoglevelHandler(ctx, log.NewLogLevelSetter(...))` to a new slog `LevelVar` setter: the numeric path segment is validated (0–10), mapped onto `slog.Level` (0 → Info, ≥1 → Debug), and reverted to the Debug default after 5 minutes unless the level was re-set in the meantime
- Invalid level values return a 400 response, change nothing, and log nothing — no panic, no raw-input echo into logs
- The now-unused `github.com/bborbe/log` import is removed from `main.go` so the binary keeps compiling (the next prompt re-adds it for the sampler)
- `make precommit` + `make test` stay green
</summary>

<objective>
Give each binary a real slog default handler and keep the `/setloglevel` endpoint functional by driving a slog `LevelVar` instead of a glog verbosity knob. This is spec 002's Desired Behavior 5 and the load-bearing piece of the migration's output parity: without it, the controller would silently drop the per-check / sunrise-sunset debug lines operators see today.
</objective>

<context>
Read first (the prompt depends on these being read before any edit):

- `./CLAUDE.md` — project conventions: `make precommit` gate, `errors.Wrap(ctx, err, "...")`, factory functions are pure composition, `github.com/bborbe/time` for `Now()` (stdlib `time` only for `time.Duration` / `time.Location`), handler wiring goes through `pkg/factory` and `pkg/handler`.
- `./docs/dod.md` — DoD: exported symbols need doc comments; Ginkgo v2 / Gomega tests; CHANGELOG entry under `## Unreleased`.
- `./.dark-factory.yaml` — `worktree: false`, `pr: false`; the daemon handles git, the agent MUST NOT commit.
- `./specs/in-progress/002-migrate-glog-slog.md` — Desired Behavior 5 (logger setup + `/setloglevel` conversion), Constraints (output parity; `/setloglevel/{level}` path stays; k8s untouched; no new env/flag knob), Failure Modes (level-parity regression; mis-wired setter; invalid level input), Security (validate the path segment, bounded, no panic, no raw-input echo, `LevelVar` write is atomic), Assumptions (numeric level mapping 0 → Info, ≥1 → Debug is an implementation decision this prompt makes).
- `./main.go` — current state after the previous prompt of this spec: slog calls converted, no glog; `router.Path("/setloglevel/{level}").Handler(log.NewSetLoglevelHandler(ctx, log.NewLogLevelSetter(2, 5*time.Minute)))` is the only remaining `github.com/bborbe/log` usage in this file (import at top, usage at lines 84-85). This prompt replaces that route and removes the import so the binary compiles.
- `./pkg/factory/factory.go` — current factory wiring; `CreateListLightsHandler` / `CreateStatusHandler` show the wrap-in-factory pattern. This prompt adds `CreateSetLogLevelHandler` here (pure composition).
- `./pkg/handler/status.go` and `./pkg/handler/list-lights.go` — the handler-package shape (constructor returning a handler, doc comment) and `./pkg/handler/status_test.go` / `./pkg/handler/list-lights_test.go` — the Ginkgo `handler_test` test shape with `httptest.ResponseRecorder` and `httptest.NewRequest`.
- `./cmd/list-lights/main.go`, `./cmd/turnon-light/main.go`, `./cmd/turnoff-light/main.go` — the three slog-emitting CLIs; each gets a default Info handler at the top of `Run`.
- `./CHANGELOG.md` — add entries under the existing `## Unreleased` section created by the previous prompt.
- Coding plugin docs (in-container paths): `/home/node/.claude/plugins/marketplaces/coding/docs/go-logging-guide.md` (slog handler setup, structured output, `no-sensitive-data-in-logs` — no raw-input echo), `/home/node/.claude/plugins/marketplaces/coding/docs/go-http-service-guide.md` (the always-required admin endpoints; `/setloglevel/{level}` is the cross-service contract this conversion must preserve), `/home/node/.claude/plugins/marketplaces/coding/docs/go-http-handler-refactoring-guide.md` (handler + httptest shape), `/home/node/.claude/plugins/marketplaces/coding/docs/go-composition.md` and `/home/node/.claude/plugins/marketplaces/coding/docs/go-factory-pattern.md` (factory = pure composition), `/home/node/.claude/plugins/marketplaces/coding/docs/go-error-wrapping-guide.md`, `/home/node/.claude/plugins/marketplaces/coding/docs/changelog-guide.md`, `/home/node/.claude/plugins/marketplaces/coding/docs/go-precommit.md`.

Verified facts the requirements rely on (read from source):

- `github.com/bborbe/log` v1.6.25 has NO slog-level setter — its `LogLevelSetter` / `NewSetLoglevelHandler` are glog-typed (`Set(ctx, glog.Level) error`). The slog setter is therefore implemented in hue's own `pkg/handler` in this prompt; `NewLogLevelSetter` / `NewSetLoglevelHandler` disappear from hue entirely (spec AC 6).
- `slog.LevelVar` is stdlib (`log/slog`): `func (v *LevelVar) Set(l Level)` is atomic, `func (v *LevelVar) Level() Level` reads it — the Security section's "no new concurrency" holds.
- `github.com/bborbe/time` provides `libtime.Now()` (repo convention for wall-clock reads; the old `log.NewLogLevelSetter` used `libtime.Now()` for its `lastSetTime` guard — replicate that).
- `github.com/gorilla/mux` is already a hue dependency; the handler reads its path variable via `mux.Vars(req)["level"]`.
</context>

<requirements>

### Step 1 — New slog setloglevel handler (`./pkg/handler/set-loglevel.go`, new file)

Create `./pkg/handler/set-loglevel.go` with the BSD license header (copy the 3-line header from `./pkg/handler/status.go`), `package handler`, and this shape:

```go
import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	libtime "github.com/bborbe/time"
	"github.com/gorilla/mux"
)
```

1. `func NewSetLogLevelHandler(levelVar *slog.LevelVar, defaultLevel slog.Level, autoResetDuration time.Duration) http.Handler` — returns a `*logLevelSetter` holding `levelVar`, `defaultLevel`, `autoResetDuration`, and the state below. Doc comment per DoD (mention: `/setloglevel/{level}`, validated 0–10, mapping 0 → Info / ≥1 → Debug, invalid input → 400 and unchanged level, auto-reset to `defaultLevel` after `autoResetDuration` unless re-set).

2. `type logLevelSetter struct` with fields:

```go
type logLevelSetter struct {
	levelVar          *slog.LevelVar
	defaultLevel      slog.Level
	autoResetDuration time.Duration

	mux         sync.Mutex
	lastSetTime time.Time
}
```

3. `ServeHTTP(resp http.ResponseWriter, req *http.Request)`:

```go
func (l *logLevelSetter) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	n, err := strconv.Atoi(mux.Vars(req)["level"])
	if err != nil || n < 0 || n > 10 {
		http.Error(resp, "invalid loglevel", http.StatusBadRequest)
		return
	}
	level := slog.LevelInfo
	if n >= 1 {
		level = slog.LevelDebug
	}
	l.setLevel(level)
	fmt.Fprintf(resp, "set loglevel to %d completed\n", n)
}
```

This is the Security-critical path: the path segment is validated (numeric via `strconv.Atoi`, bounded to 0–10) BEFORE it touches the `LevelVar`; invalid input yields a 400 error response, logs nothing, leaves the `LevelVar` unchanged, and cannot panic (spec Security + Failure Modes row "Invalid level value"). Do not log the raw input.

4. `setLevel(level slog.Level)` and `resetLater()` implement the TTL-reset semantics of the old `log.NewLogLevelSetter` (same lastSetTime guard, so an operator re-setting within the window does not get clobbered by an earlier reset):

```go
func (l *logLevelSetter) setLevel(level slog.Level) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.lastSetTime = libtime.Now()
	l.levelVar.Set(level)
	go l.resetLater()
}

func (l *logLevelSetter) resetLater() {
	time.Sleep(l.autoResetDuration)
	l.mux.Lock()
	defer l.mux.Unlock()
	if libtime.Now().Sub(l.lastSetTime) <= l.autoResetDuration {
		return // re-set in the meantime; keep the newer level
	}
	l.levelVar.Set(l.defaultLevel)
}
```

### Step 2 — Unit tests for the handler (`./pkg/handler/set-loglevel_test.go`, new file)

Create `./pkg/handler/set-loglevel_test.go`, `package handler_test`, Ginkgo v2 / Gomega, following the shape of `./pkg/handler/list-lights_test.go`. Register the handler on a real `mux.NewRouter()` at `router.Path("/setloglevel/{level}")` so the `mux.Vars` path-variable seam is exercised (this is the boundary the production traffic crosses).

1. In `BeforeEach`: `levelVar := &slog.LevelVar{}` with `levelVar.Set(slog.LevelDebug)` (the controller default); router with `handler.NewSetLogLevelHandler(levelVar, slog.LevelDebug, 20*time.Millisecond)`; `recorder := httptest.NewRecorder()`.
2. It "maps level 0 to Info and level 1 to Debug": `router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/setloglevel/0", nil))` → `Expect(recorder.Code).To(Equal(http.StatusOK))` and `Expect(levelVar.Level()).To(Equal(slog.LevelInfo))`; repeat for `/setloglevel/1` → `slog.LevelDebug`.
3. It "rejects invalid levels without changing the level": loop over `[]string{"abc", "-1", "99"}`, each → `Expect(recorder.Code).To(Equal(http.StatusBadRequest))` and `Expect(levelVar.Level()).To(Equal(slog.LevelDebug))` (unchanged).
4. It "resets to the default level after the auto-reset duration": `/setloglevel/0` (Info), then `Eventually(levelVar.Level).Should(Equal(slog.LevelDebug), "level should auto-reset to Debug", 5*time.Second, 10*time.Millisecond)` (TTL is 20ms; use `Eventually` to absorb goroutine-scheduling jitter — a fixed `time.Sleep` would flake under CI load).
5. It "keeps a re-set level past the auto-reset": `/setloglevel/0` (Info), then immediately `/setloglevel/1` (Debug), then `Consistently(levelVar.Level).Should(Equal(slog.LevelDebug))` for 100ms (the earlier reset is skipped by the lastSetTime guard; the level stays at the re-set value).

### Step 3 — Factory wiring (`./pkg/factory/factory.go`)

Add `"log/slog"` to the imports (stdlib group) and append this function (pure composition — no conditionals, no I/O, no `context.Background()`):

```go
// CreateSetLogLevelHandler wraps handler.NewSetLogLevelHandler with the
// controller's default level (Debug, matching the LOGLEVEL=2 deployment) and
// the 5-minute auto-reset so it can be mounted on a mux.Router.
func CreateSetLogLevelHandler(levelVar *slog.LevelVar) http.Handler {
	return handler.NewSetLogLevelHandler(levelVar, slog.LevelDebug, 5*time.Minute)
}
```

`handler` and `time` are already imported in this file. Nothing else in `factory.go` changes.

### Step 4 — Controller composition root (`./main.go`)

1. Add `"log/slog"` to the stdlib import group.
2. In `func (a *application) Run(...)`, at the top (before any slog call), install the default handler and create the `LevelVar` (spec DB 5: the controller default `Debug` is what makes the `LOGLEVEL=2` output parity hold):

```go
	var logLevel slog.LevelVar
	logLevel.Set(slog.LevelDebug)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: &logLevel})))
```

3. Pass the `LevelVar` into the HTTP server builder: change the call `a.createHttpServer()` to `a.createHttpServer(&logLevel)`, and change the signature to `func (a *application) createHttpServer(logLevel *slog.LevelVar) run.Func`.
4. In `createHttpServer`, replace the route line

```go
router.Path("/setloglevel/{level}").
	Handler(log.NewSetLoglevelHandler(ctx, log.NewLogLevelSetter(2, 5*time.Minute)))
```

with

```go
router.Path("/setloglevel/{level}").Handler(factory.CreateSetLogLevelHandler(logLevel))
```

The `/setloglevel/{level}` path is unchanged (spec Constraint). The local `ctx` remains used by `context.WithCancel` and `libhttp.NewServer(...).Run(ctx)` — do not remove it.
5. REMOVE the now-unused `"github.com/bborbe/log"` import from `main.go` — after this step it has no remaining usages in the file, and leaving it breaks compilation. (The next prompt of this spec re-adds it for the sampler factory.)

### Step 5 — CLI defaults (`./cmd/list-lights/main.go`, `./cmd/turnon-light/main.go`, `./cmd/turnoff-light/main.go`)

In each of the three slog-emitting CLIs, add `"log/slog"` to the stdlib import group and the first line of `Run`:

```go
slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))
```

Each already imports `os` (used by `os.Exit`). `cmd/create-token` has no slog and no glog — it is untouched (out of scope).

### Step 6 — CHANGELOG

Append bullets under the existing `## Unreleased` section in `./CHANGELOG.md` (conventional prefixes), e.g.:

```markdown
- feat: Install a default text-to-stderr slog handler per binary — controller default level Debug (matches the LOGLEVEL=2 deployment), the list-lights / turnon-light / turnoff-light CLIs default Info
- refactor: Convert the /setloglevel/{level} endpoint from the glog-typed log.NewSetLoglevelHandler to a validated slog LevelVar setter (0 → Info, ≥1 → Debug, invalid input → 400, 5-minute auto-reset)
```

### Step 7 — Self-check

Walk spec Desired Behavior 5 and Acceptance Criterion 6 against the change. Re-run the `<verification>` block below and confirm every check passes, then run `make precommit` and `make test` and confirm both exit 0.
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git. No `git commit`, `git push`, `git add`, no PR (spec repo CLAUDE.md + `.dark-factory.yaml`).
- Spec Constraints (copied verbatim): never code directly (dark-factory pipeline); repo conventions frozen — `github.com/bborbe/time` not stdlib `time` (use `libtime.Now()` for the reset guard), `github.com/bborbe/errors` with `errors.Wrap(ctx, err, "...")`, Ginkgo v2 / Gomega, external test packages (`pkg_test`); factory functions stay pure composition (no conditionals, no I/O, no `context.Background()`); no new counterfeiter mock for the sampler (sibling PR 2 owns mocks); the controller's default slog level must emit the same lines the `LOGLEVEL=2` deployment emits today (output parity — `LevelVar` default `slog.LevelDebug`); the `/setloglevel` endpoint path stays `/setloglevel/{level}`; `k8s/hue-deploy.yaml` and `k8s/k8s.env` stay untouched (the `-v={LOGLEVEL}` line is glog's flag and inert after the migration — no new env/flag log-level knob is added); `make precommit` mandatory before commit; existing tests keep passing; `main_test.go` / `cmd/*/main_test.go` `gexec.Build` compile checks must stay green; no scenario is added.
- Spec Non-goals that stay out of scope: introducing a logging wrapper package (use stdlib slog package-level calls directly), keeping `/setloglevel` wired to `log.NewLogLevelSetter` (dead knob after the migration), adding any env/flag log-level config surface, re-plumbing the k8s `LOGLEVEL` env / `-v` flag.
- Do NOT do the next prompt's work: no sampler wiring in `main.go` / `pkg/factory` / `pkg/check/checks-cron.go` (that prompt re-adds the `github.com/bborbe/log` import to `main.go`), no list-lights aggregation.
- Do NOT touch `k8s/`, `go.mod`, `tools.env`, `mocks/`, or any `cmd/create-token` file.
- Do NOT renumber the prompt filename — dark-factory assigns numbers on approve.
</constraints>

<verification>
All commands run from the repo root. Zero-match greps use a form with a truthful exit code.

1. `[ -z "$(grep -rn 'NewLogLevelSetter\|NewSetLoglevelHandler' . --include='*.go')" ]` — must exit 0 (spec AC 6: glog-typed setter symbols gone).
2. `grep -n 'setloglevel' main.go` — must print ≥1 line (spec AC 6: route still registered).
3. `grep -n 'LevelVar' main.go` — must print ≥1 line (spec AC 6: the endpoint drives a slog level).
4. `grep -rn 'slog.SetDefault' main.go cmd/list-lights/main.go cmd/turnon-light/main.go cmd/turnoff-light/main.go` — must print exactly 4 lines (one default handler per binary).
5. `[ -z "$(grep -rn 'github.com/bborbe/log' main.go)" ]` — must exit 0 (import removed; the next prompt re-adds it for the sampler).
6. `[ -z "$(grep -r 'glog' --include='*.go' .)" ]` — must still exit 0 (spec AC 1 holds).
7. `[ -z "$(grep -rn 'slog.Debug' cmd/ --include='main.go')" ]` — must still exit 0 (spec AC 3 holds).
8. `grep -rn 'slog.Info' cmd/ --include='main.go'` — must still print ≥6 lines (spec AC 3 holds).
9. `make precommit` — must exit 0 (DoD gate per `docs/dod.md`).
10. `make test` — must exit 0.
</verification>
