GO ?= go
HTTP_ADDR ?= :8080
JWT_SECRET ?= classroom-secret-change-me
TOKEN_TTL ?= 2h
BASE_URL ?= http://localhost:8080

.PHONY: run test test-unit test-race fmt fmt-check vet deps ci curl-demo check-before

run:
	HTTP_ADDR='$(HTTP_ADDR)' JWT_SECRET='$(JWT_SECRET)' TOKEN_TTL='$(TOKEN_TTL)' $(GO) run ./cmd/api

test:
	$(GO) test ./...

test-unit:
	$(GO) test ./internal/user ./internal/note ./internal/auth ./internal/httpapi/...

test-race:
	$(GO) test -race ./...

fmt:
	gofmt -w $$(find . -type f -name '*.go' -not -path './before/*')

fmt-check:
	@test -z "$$(gofmt -l $$(find . -type f -name '*.go' -not -path './before/*'))" || (gofmt -l $$(find . -type f -name '*.go' -not -path './before/*'); exit 1)

vet:
	$(GO) vet ./...

deps:
	$(GO) list -deps ./... >/dev/null

ci: fmt-check vet deps test test-race

curl-demo:
	BASE_URL='$(BASE_URL)' ./scripts/curl-demo.sh

check-before:
	cd before && $(GO) test ./...
