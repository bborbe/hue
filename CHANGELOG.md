# Changelog

All notable changes to this project will be documented in this file.

Please choose versions by [Semantic Versioning](http://semver.org/).

* MAJOR version when you make incompatible API changes,
* MINOR version when you add functionality in a backwards-compatible manner, and
* PATCH version when you make backwards-compatible bug fixes.

## v0.0.2

- modernize Makefile to canonical bborbe pattern (tools.env, overridable VULNCHECK_IGNORE, panic-safe vulncheck, osv-scanner, trivy)
- bump golang.org/x/net v0.43.0 → v0.56.0, golang.org/x/sys v0.35.0 → v0.46.0 (CVE fixes)
- drop tools.go in favor of pinned @VERSION invocations in Makefile
- gate -race behind ENABLE_RACE to avoid cmd/* gexec flakes
- add .golangci.yml, .osv-scanner.toml, .trivyignore

## v0.0.1

- Initial Version
