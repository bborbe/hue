---
status: completed
summary: Split the cluster build arm's base-image source (DOCKER_REGISTRY) from its push target (new PUSH_REGISTRY), wired the build Job build arg through __DOCKER_REGISTRY__, and added an Unreleased changelog entry
execution_id: hue-exec-015-separate-cluster-base-image-from-push-target
dark-factory-version: v0.198.17
created: "2026-11-04T01:13:52Z"
queued: "2026-11-04T01:13:52Z"
started: "2026-11-04T01:14:20Z"
completed: "2026-11-04T01:15:32Z"
branch: dark-factory/015-separate-cluster-base-image-from-push-target
---

# fix: separate the cluster base-image source from the push target

<summary>
- The cluster build arm hardcoded its base-image source to `docker.io` in the shared build Job template, while its push target came from a separate pinned pair
- `DOCKER_REGISTRY` now names only the base-image source
- A new `PUSH_REGISTRY` names where the built image is published, defaulting to `docker.io`
- `PUSH_REGISTRY` applies to the cluster arm only; the local arm is unchanged
- The published image is `docker.io/bborbe/hue`, never a bare `docker.io/hue`
- The build Job's Dockerfile build arg now comes from `DOCKER_REGISTRY` instead of a hardcoded `docker.io`
- The per-repo `CLUSTER_DOCKER_REGISTRY`/`CLUSTER_IMAGE` pair is removed
- CHANGELOG gets a new `## Unreleased` heading with one bullet
</summary>

<objective>
Split the cluster build arm's registry configuration into a base-image source (`DOCKER_REGISTRY`) and a push target (`PUSH_REGISTRY`, default `docker.io`), so the cluster arm takes its base-image source from the repository's mirror and its published image name from the push target, independently.
</objective>

<context>
Precondition: `hack/build-job.yaml` and `CLUSTER_DOCKER_REGISTRY` must already exist on this branch — they landed on `master` in release `v0.8.0`. If either is missing, the checkout is stale: stop and report, do not proceed. Every requirement below is a no-op otherwise, and the greps in requirement 7 would pass without the change being made.

Read first:

- `./CLAUDE.md` — dark-factory workflow; `make precommit` is the gate.
- `./Makefile` — the `CLUSTER_DOCKER_REGISTRY ?= docker.io` and `CLUSTER_IMAGE ?= bborbe/$(IMAGE)` pair near the top, the cluster arm of `build:` that passes `IMAGE='$(CLUSTER_DOCKER_REGISTRY)/$(CLUSTER_IMAGE)'`, and the cluster arms of `upload:` and `apply:` that echo the same pair.
- `./hack/build-job.yaml` — the `# Template tokens:` table near the top, and the `build` container's `buildctl build` line that hardcodes `--opt build-arg:DOCKER_REGISTRY=docker.io`.
- `./hack/build-cluster.sh` — already carries `DOCKER_REGISTRY="${DOCKER_REGISTRY:-docker.io}"` and substitutes `__DOCKER_REGISTRY__`; no change is needed there.
- `./CHANGELOG.md` — the latest released heading is `## v0.8.1`; there is no `## Unreleased` section yet.

`hue`'s `Dockerfile` starts `FROM golang:1.27.3` and never references `${DOCKER_REGISTRY}`, so the base pull stays public. This change is variable separation, not a pull-behaviour change; requirement 5 keeps the shared template's build arg wired to the repo's own `DOCKER_REGISTRY` for fleet consistency, not to change hue's pull.

`hack/build-job.yaml` is also rendered by the release-to-image dispatcher on release (outside this repo). A new `__DOCKER_REGISTRY__` token is safe in this repository because `hack/build-cluster.sh` already substitutes it (see the `sed` list). Whether the out-of-repo dispatcher substitutes it is verified outside this prompt; if it does not, the literal token reaches BuildKit as a build arg that `hue`'s `Dockerfile` ignores, so the release build is unaffected.
</context>

<requirements>
1. In `Makefile`, replace the `CLUSTER_DOCKER_REGISTRY ?= docker.io` / `CLUSTER_IMAGE ?= bborbe/$(IMAGE)` pair with a single `PUSH_REGISTRY ?= docker.io`. Update the explanatory comment so it names `DOCKER_REGISTRY` (the base-image source) and `PUSH_REGISTRY` (the push target, cluster arm only).
2. In the cluster arm of `Makefile`'s `build:` target, replace `IMAGE='$(CLUSTER_DOCKER_REGISTRY)/$(CLUSTER_IMAGE)'` with two environment lines: `DOCKER_REGISTRY='$(DOCKER_REGISTRY)'` and `IMAGE='$(PUSH_REGISTRY)/bborbe/$(IMAGE)'`.
3. Update the cluster-arm `upload:` echo to name `$(PUSH_REGISTRY)/bborbe/$(IMAGE):$(VERSION)`.
4. In the cluster arm of `Makefile`'s `apply:` target, update the explanatory comment and the one `@echo` line that carries the old pair (the first of the two) to name `$(PUSH_REGISTRY)/bborbe/$(IMAGE):$(VERSION)` instead of `$(CLUSTER_DOCKER_REGISTRY)/$(CLUSTER_IMAGE):$(VERSION)`. Leave the second `@echo` line unchanged: it describes what `k8s/hue-deploy.yaml` renders from `k8s.env`, which is a different registry and must keep saying so. The `clean:` cluster arm does not reference the pair and needs no change.
5. In `hack/build-job.yaml`, replace `--opt build-arg:DOCKER_REGISTRY=docker.io` with `--opt build-arg:DOCKER_REGISTRY=__DOCKER_REGISTRY__`, and add `__DOCKER_REGISTRY__      base-image registry passed as a build arg` to the file's `# Template tokens:` table so every token stays documented.
6. Add a new `## Unreleased` heading above `## v0.8.1` in `CHANGELOG.md`, with one bullet starting with the `fix:` prefix (conventional prefix required so dark-factory derives the version bump; match the existing entries) describing the split and naming `PUSH_REGISTRY`.
7. Grep `Makefile`, `hack/`, `k8s/` and `docs/` for `CLUSTER_DOCKER_REGISTRY` and `CLUSTER_IMAGE` — no match may remain. Exclude `CHANGELOG.md`: its released `## v0.8.1` entry records the old names and must not be rewritten. (`CLUSTER_SERVICE` does not exist in this repo; do not grep for it.)
</requirements>

<constraints>
- Do NOT commit — dark-factory handles git.
- Do NOT change the local arm's behaviour.
- Do NOT add a merge init container: `hue`'s base images are public, so no pull credential is needed.
- Do NOT rewrite released CHANGELOG sections (`## v0.8.1` and older); only add under a new `## Unreleased` heading.
- `make precommit` must stay green.
</constraints>

<verification>
make -n build BUILD_TARGET=nukedev REF=HEAD VERSION=v0.0.0-test | grep -F "IMAGE='docker.io/bborbe/hue'"
make -n build BUILD_TARGET=nukedev REF=HEAD VERSION=v0.0.0-test | grep -F "DOCKER_REGISTRY='docker.prod.nuke.benjamin-borbe.de:443'"
grep -F 'build-arg:DOCKER_REGISTRY=__DOCKER_REGISTRY__' hack/build-job.yaml
grep -F 's|__DOCKER_REGISTRY__|${DOCKER_REGISTRY}|g' hack/build-cluster.sh
sed -e 's|__DOCKER_REGISTRY__|docker.example:443|g' hack/build-job.yaml | grep -F 'build-arg:DOCKER_REGISTRY=docker.example:443'
! grep -F 'DOCKER_REGISTRY' Dockerfile
! grep -rn 'CLUSTER_DOCKER_REGISTRY\|CLUSTER_IMAGE' Makefile hack k8s docs
make precommit
</verification>
