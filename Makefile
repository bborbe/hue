DOCKER_REGISTRY ?= docker.prod.nuke.benjamin-borbe.de:443
IMAGE ?= hue
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD | tr '/' '-')
DIRS += $(shell find */* -maxdepth 0 -name Makefile -exec dirname "{}" \;)

# Which machine builds the image. `local` builds here with the local daemon; any
# other value is a cluster alias (sun, nukedev, nukeprod) or a raw kubectl
# context name, and the build runs as a Job there.
# Unset BUILD_TARGET resolves automatically: a bounded `kubectl get --raw /readyz`
# probe through hack/cluster-alias.sh picks sun when it answers and the local
# daemon when it does not. An explicit BUILD_TARGET always wins and the probe is
# never run. SUN_PROBE=fail short-circuits the probe to "unreachable", so the
# fallback can be exercised without touching the cluster:
#   make -n build SUN_PROBE=fail
#
# The probe is cheap where sun is absent: with no sun kubeconfig on the host,
# kubectl fails immediately, so CI pays nothing. It costs the timeout only on a
# machine that holds the kubeconfig but cannot reach the cluster.
BUILD_TARGET_PROBE_TIMEOUT ?= 5
SUN_PROBE ?=

ifeq ($(origin BUILD_TARGET),undefined)
  ifeq ($(SUN_PROBE),fail)
    BUILD_TARGET := local
    BUILD_TARGET_REASON := sun unreachable: SUN_PROBE=fail
  else
    BUILD_TARGET_PROBE_OUT := $(shell BUILD_TARGET=sun bash -c 'source hack/cluster-alias.sh && kc --request-timeout=$(BUILD_TARGET_PROBE_TIMEOUT)s get --raw /readyz' 2>/dev/null | grep -qx ok && echo ok)
    ifeq ($(BUILD_TARGET_PROBE_OUT),ok)
      BUILD_TARGET := sun
      BUILD_TARGET_REASON := probe ok
      # No REF/VERSION default on this arm, deliberately: a Makefile-level default
      # gives VERSION origin `file`, which the cluster arm's guard rejects, and a
      # branch name carrying `/` is rejected by hack/build-cluster.sh. A cluster
      # build is invoked as `make build VERSION=<tag>`; the local fallback stays
      # available via BUILD_TARGET=local.
    else
      BUILD_TARGET := local
      BUILD_TARGET_PROBE_ERR := $(shell BUILD_TARGET=sun bash -c 'source hack/cluster-alias.sh && kc --request-timeout=$(BUILD_TARGET_PROBE_TIMEOUT)s get --raw /readyz' 2>&1 >/dev/null | head -1)
      BUILD_TARGET_REASON := sun unreachable: $(if $(BUILD_TARGET_PROBE_ERR),$(BUILD_TARGET_PROBE_ERR),probe failed)
    endif
  endif
  $(info build target: $(BUILD_TARGET) ($(BUILD_TARGET_REASON)))
else
  $(info build target: $(BUILD_TARGET) (explicit))
endif

# Cluster-build inputs. The vendored cluster script defaults the namespace,
# timeout and Dockerfile; the clone URL, tag and image come from this repository.
BUILD_TIMEOUT ?= 3600
BUILD_REPO_URL ?= https://github.com/bborbe/$(IMAGE).git
BUILD_DOCKERFILE ?= Dockerfile
BUILD_PUSH_SECRET ?= buildkit-push

# The commit the cluster clones. Empty on purpose: the cluster arm must be told
# which commit to build, and the script refuses to run without it.
REF ?=

# The tag the cluster pushes. Empty on purpose, and separate from BRANCH: a
# cluster build must never overwrite the moving `:master` tag.
VERSION ?=

# ⚠️ The cluster arm publishes to Docker Hub, not to the private mirror this
# repository defaults to. DOCKER_REGISTRY above is the base-image source; a
# cluster build would otherwise push `docker.prod.nuke.benjamin-borbe.de:443/hue`,
# which the Job has no credential for, and would publish under the bare name
# `hue` rather than `bborbe/hue`. PUSH_REGISTRY is the push target, and applies
# to the cluster arm only.
PUSH_REGISTRY ?= docker.io

# Export BRANCH so k8s/Makefile's bash subshell sees it for
# {{"BRANCH" | env}} template substitution in hue-deploy.yaml.
export BRANCH

include tools.env

.PHONY: default
default: precommit

.PHONY: run
run:
	@go run -mod=mod main.go \
	-token=$$(teamvault-password --teamvault-config ~/.teamvault.json --teamvault-key=QL3QQw) \
	-v=2

.PHONY: precommit
precommit: ensure format generate test check addlicense
	@echo "ready to commit"

.PHONY: ensure
ensure:
	go mod tidy -e
	go mod verify
	rm -rf vendor

.PHONY: format
format:
	find . -type f -name 'go.mod' -not -path './vendor/*' -exec go run github.com/shoenig/go-modtool@$(GO_MODTOOL_VERSION) -w fmt "{}" \;
	go run github.com/incu6us/goimports-reviser/v3@$(GOIMPORTS_REVISER_VERSION) -project-name github.com/bborbe/hue -format -excludes vendor ./...
	find . -type d -name vendor -prune -o -type f -name '*.go' -print0 | xargs -0 -n 10 go run github.com/segmentio/golines@$(GOLINES_VERSION) --max-len=100 -w
	# golines last, then gofmt last so its wrapping is normalized and the gofmt lint check passes
	find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w "{}" +

.PHONY: generate
generate:
	rm -rf mocks avro
	mkdir -p mocks
	echo "package mocks" > mocks/mocks.go
	go generate -mod=mod ./...

# -race=true catches data races but flakes on some CI runners (rare SIGSEGV
# during gexec.Build in cmd/*-style binary smoke tests). Default off; opt in
# via ENABLE_RACE=true for nightly/manual hardening runs.
TESTFLAGS_RACE = -race=false
ifdef ENABLE_RACE
	TESTFLAGS_RACE = -race=true
endif

.PHONY: test
test:
	go test -mod=mod -p=$${GO_TEST_PARALLEL:-1} -cover $(TESTFLAGS_RACE) $(shell go list -mod=mod ./... | grep -v /vendor/)

.PHONY: check
check: vet vulncheck osv-scanner trivy
# TODO: add lint after fixing pre-existing revive/dupl findings

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run --allow-parallel-runners --config .golangci.yml ./...

.PHONY: vet
vet:
	go vet -mod=mod $(shell go list -mod=mod ./... | grep -v /vendor/)


VULNCHECK_IGNORE ?= GO-2026-4923 GO-2026-4514 GO-2022-0470 GO-2026-4772 GO-2026-4771 GO-2026-6599 GO-2026-6600 GO-2026-6603 GO-2026-6604 GO-2026-6605 GO-2026-6607 GO-2026-6608 GO-2026-6609 GO-2026-6610 GO-2026-6611 GO-2026-6612 GO-2026-6613 GO-2026-6617

.PHONY: vulncheck
vulncheck:
	@PKGS="$(shell go list -mod=mod ./... | grep -v /vendor/)"; \
	IGNORE_JSON=$$(printf '%s\n' $(VULNCHECK_IGNORE) | jq -R . | jq -s .); \
	REMAIN=$$(go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) -format json $$PKGS 2>/dev/null | \
		jq -rs --argjson ignore "$$IGNORE_JSON" \
			'(map(select(.osv != null)) | map({key: .osv.id, value: (.osv.summary // "")}) | from_entries) as $$sum | \
			 map(select(.finding != null) | .finding) | \
			 map(select(.osv as $$o | $$ignore | index($$o) | not)) | \
			 map("\(.osv)\t\(.trace[-1].module)@\(.trace[-1].version) -> \(.fixed_version)\t\($$sum[.osv] // "")") | \
			 unique | .[]'); \
	if [ -n "$$REMAIN" ]; then \
		echo "Unexpected vulnerabilities (ignored: $(VULNCHECK_IGNORE)):"; \
		printf '%s\n' "$$REMAIN" | column -t -s "$$(printf '\t')"; \
		exit 1; \
	else \
		echo "No unignored vulnerabilities found"; \
	fi

.PHONY: osv-scanner
osv-scanner:
	@if [ -f .osv-scanner.toml ]; then \
		echo "Using .osv-scanner.toml"; \
		go run github.com/google/osv-scanner/v2/cmd/osv-scanner@$(OSV_SCANNER_VERSION) --config .osv-scanner.toml --recursive .; \
	else \
		echo "No config found, running default scan"; \
		go run github.com/google/osv-scanner/v2/cmd/osv-scanner@$(OSV_SCANNER_VERSION) --recursive .; \
	fi

.PHONY: trivy
trivy:
	trivy fs \
	--db-repository ghcr.io/aquasecurity/trivy-db \
	--scanners vuln,secret \
	--quiet \
	--no-progress \
	--disable-telemetry \
	--exit-code 1 .

.PHONY: addlicense
addlicense:
	go run github.com/google/addlicense@$(ADDLICENSE_VERSION) -c "Benjamin Borbe" -y $$(date +'%Y') -l bsd $$(find . -name "*.go" -not -path './vendor/*')

.PHONY: buca
buca: build upload clean apply


.PHONY: build
ifeq ($(BUILD_TARGET),local)
build: check-go-mod
	DOCKER_BUILDKIT=1 \
	docker build \
	--no-cache \
	--rm=true \
	--platform=linux/amd64 \
	--build-arg DOCKER_REGISTRY=$(DOCKER_REGISTRY) \
	--build-arg BRANCH=$(BRANCH) \
	--build-arg BUILD_GIT_VERSION=$$(git describe --tags --always --dirty) \
	--build-arg BUILD_GIT_COMMIT=$$(git rev-parse --short HEAD) \
	--build-arg BUILD_DATE=$$(date -u +%Y-%m-%dT%H:%M:%SZ) \
	-t $(DOCKER_REGISTRY)/$(IMAGE):$(BRANCH) \
	-f Dockerfile .
else
build:
	@if [ -z "$(REF)" ]; then \
		echo "ERROR: REF is required for a cluster build — pass REF=<commit>" >&2; \
		exit 1; \
	fi
	@if [ -z "$(VERSION)" ]; then \
		echo "ERROR: VERSION is required for a cluster build — pass VERSION=<scratch-tag>." >&2; \
		echo "  The cluster arm pushes, so an unset tag would overwrite a moving reference." >&2; \
		exit 1; \
	fi
	BUILD_TARGET='$(BUILD_TARGET)' \
	NAMESPACE='build' \
	TIMEOUT='$(BUILD_TIMEOUT)' \
	REPO_URL='$(BUILD_REPO_URL)' \
	REF='$(REF)' \
	TAG='$(VERSION)' \
	DOCKERFILE='$(BUILD_DOCKERFILE)' \
	DOCKER_REGISTRY='$(DOCKER_REGISTRY)' \
	IMAGE='$(PUSH_REGISTRY)/bborbe/$(IMAGE)' \
	PUSH='true' \
	PUSH_SECRET='$(BUILD_PUSH_SECRET)' \
	hack/build-cluster.sh
endif

.PHONY: check-go-mod
check-go-mod:
	@if [ -f "go.mod" ]; then \
		echo "go.mod found, running go mod vendor..."; \
		go mod vendor; \
	else \
		echo "go.mod not found, skipping go mod vendor."; \
	fi


.PHONY: upload
upload:
ifeq ($(BUILD_TARGET),local)
	docker push $(DOCKER_REGISTRY)/$(IMAGE):$(BRANCH)
else
	@echo "upload: skipped — BUILD_TARGET=$(BUILD_TARGET); the cluster Job already pushed $(PUSH_REGISTRY)/bborbe/$(IMAGE):$(VERSION)"
endif

.PHONY: clean
clean:
ifeq ($(BUILD_TARGET),local)
	docker rmi $(DOCKER_REGISTRY)/$(IMAGE):$(BRANCH) || true
	# docker builder prune -a -f
	docker builder prune --max-used-space 21474836480 -f || true
	rm -rf vendor
else
	@echo "clean: skipped — BUILD_TARGET=$(BUILD_TARGET); no local image to remove"
endif

.PHONY: apply
ifeq ($(BUILD_TARGET),local)
apply:
	@for i in $(DIRS); do \
		cd $$i; \
		echo "apply $${i}"; \
		make apply; \
		cd ..; \
	done
else
# A cluster build pushes $(PUSH_REGISTRY)/bborbe/$(IMAGE):$(VERSION), but
# k8s/hue-deploy.yaml renders the image from k8s.env's DOCKER_REGISTRY and the exported
# BRANCH — a different registry and a different tag. Applying it here would deploy an
# image this run never built, so `make buca BUILD_TARGET=<cluster>` stops after the push.
apply:
	@echo "apply: skipped — BUILD_TARGET=$(BUILD_TARGET); the cluster Job pushed $(PUSH_REGISTRY)/bborbe/$(IMAGE):$(VERSION)."
	@echo "  k8s/hue-deploy.yaml renders \$$(DOCKER_REGISTRY)/\$$(IMAGE):\$$(BRANCH), which is not the image this run built."
	@echo "  Deploy deliberately once the cluster build has been verified."
endif