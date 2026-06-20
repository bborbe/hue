# Changelog

All notable changes to this project will be documented in this file.

Please choose versions by [Semantic Versioning](http://semver.org/).

* MAJOR version when you make incompatible API changes,
* MINOR version when you add functionality in a backwards-compatible manner, and
* PATCH version when you make backwards-compatible bug fixes.

## Unreleased

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
