# Changelog

All notable changes to this project will be documented in this file.

Please choose versions by [Semantic Versioning](http://semver.org/).

* MAJOR version when you make incompatible API changes,
* MINOR version when you add functionality in a backwards-compatible manner, and
* PATCH version when you make backwards-compatible bug fixes.

## Unreleased

- feat: Add canonical `build_info` Prometheus gauge — wire `BUILD_GIT_VERSION` / `BUILD_GIT_COMMIT` / `BUILD_DATE` build-args (already passed by Makefile) through `Dockerfile` (ARG → OCI labels + ENV) into `main.go` (3 new fields + `libmetrics.NewBuildInfoMetrics().SetBuildInfo(...)`). Enables `count by (version) (build_info)` Prometheus query across the fleet + populates OCI image labels. Matches go-skeleton / recurring-task-creator / kafka-topic-reader.
- refactor: Extract inline `/lights` handler from `main.go` into `pkg/handler/list-lights.go` + `factory.CreateListLightsHandler`. Per [HTTP Handler Refactoring Guide](https://github.com/bborbe/coding-guidelines/blob/master/docs/go-http-handler-refactoring-guide.md): handlers belong in `pkg/handler/`, factory wires them. Side effect: the cache layer in `BridgesProvider` (`NewBridgeProviderCache`) was being thrown away every request because the provider was rebuilt inline; the refactor builds it once at server-creation time and shares it across requests.
- feat: Add `/status` JSON endpoint exposing `{"on": [...names...], "off": [...names...]}` for at-a-glance fleet view across the bridge's lights. Sorted name lists for stable output.
- chore: Right-size k8s resources — live steady-state is 2m CPU / 26Mi RAM; limits `1000m`/`1000Mi` were 500×/38× headroom (effectively no limit). New: request `20m`/`100Mi`, limit `200m`/`256Mi`. Frees ~100Mi reservation back to the node, keeps generous safety margin (100× CPU / 10× RAM at limit).

## v0.0.4

- fix: `export BRANCH` in main Makefile so the `k8s/Makefile apply` bash subshell can substitute `{{"BRANCH" | env}}` in `hue-deploy.yaml`. Without this, `make buca` produced `image: docker.quant.benjamin-borbe.de:443/hue:` (empty tag) and the pod failed with `InvalidImageName`.

## v0.0.3

- fix: Tag Docker image with `$(BRANCH)` instead of `$(VERSION)` so it matches `k8s/hue-deploy.yaml` (`image: bborbe/hue:{{BRANCH}}`) and keel.sh auto-roll. Regression from kafka-topic-reader template; sibling services (backup etc.) use `$(BRANCH)` for keel-driven deploys.
- fix: `k8s/Makefile apply` now uses the `kubectlquant` wrapper instead of `kubectl --context=$$CLUSTER_CONTEXT`. `quant` is the wrapper alias, not a kubeconfig context name, so the old form failed with `error: context "quant" does not exist`. Matches sibling pattern (backup uses `kubectlhell`).
- chore: Switch Docker registry from `docker.io/bborbe/hue` to the private `docker.quant.benjamin-borbe.de:443/hue` (matches `recurring-task-creator` pattern). Push image stays internal, pulls are colocated with the quant cluster, no public registry bloat for an internal infra service.
- fix: Add `imagePullSecrets: docker-quant` to the Deployment + `k8s/hue-docker-secret.yaml` (`kubernetes.io/dockerconfigjson` Secret using `DOCKER_REGISTRY_KEY` via teamvault). Without these, the pod would fail with `ImagePullBackOff` against the private registry. Matches `recurring-task-creator`.

## v0.0.2

- chore: Modernize Makefile to canonical bborbe pattern (`tools.env`, overridable `VULNCHECK_IGNORE`, panic-safe vulncheck, osv-scanner, trivy)
- chore: Bump golang.org/x/net v0.43.0 → v0.56.0 (CVE fixes: GO-2026-4440, GO-2026-4441, GO-2026-4918, GO-2026-5025..5030)
- chore: Bump golang.org/x/sys v0.35.0 → v0.46.0 (GO-2026-5024)
- chore: Drop tools.go in favor of pinned `@VERSION` invocations from `tools.env`
- test: Gate `-race` behind `ENABLE_RACE` to avoid cmd/* gexec SIGSEGV flakes
- chore: Add .golangci.yml, .osv-scanner.toml, .trivyignore for new check targets
- fix: Correct `CLUSTER_CONTEXT=fire` → `quant` in k8s/k8s.env (matches actual deploy target `hue.quant.benjamin-borbe.de`)
- docs: Add CLAUDE.md with dark-factory workflow, architecture map, and key design decisions
- chore: Remove leftover `fix` Makefile target inherited from kafka-topic-reader template (`go get @latest` is non-reproducible; `updater go` handles dep bumps)

## v0.0.1

- Initial Version
