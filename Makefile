precommit: ensure format generate test check
	@echo "ready to commit"

ensure:
	go mod verify
	go mod vendor

format:
	go install -mod=vendor github.com/incu6us/goimports-reviser
	find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w "{}" +
	find . -type f -name '*.go' -not -path './vendor/*' -exec goimports-reviser -project-name bitbucket.apps.seibert-media.net -file-path "{}" \;

generate:
	rm -rf mocks avro
	go generate -mod=vendor ./...

test:
	go test -mod=vendor -cover -race $(shell go list -mod=vendor ./... | grep -v /vendor/)

check: lint vet errcheck

lint:
	go install -mod=vendor golang.org/x/lint/golint
	@GOFLAGS=-mod=vendor golint -min_confidence 1 $(shell go list -mod=vendor ./... | grep -v /vendor/)

vet:
	@go vet -mod=vendor $(shell go list -mod=vendor ./... | grep -v /vendor/)

errcheck:
	go install -mod=vendor github.com/kisielk/errcheck
	@GOFLAGS=-mod=vendor errcheck -ignore '(Close|Write|Fprint)' $(shell go list -mod=vendor ./... | grep -v /vendor/)
