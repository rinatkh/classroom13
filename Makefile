SHELL := /bin/bash
GO ?= go
BASE_URL ?= http://localhost:8080
HTTP_ADDR ?= :8080
TOKEN_TTL ?= 2h
export BASE_URL HTTP_ADDR TOKEN_TTL JWT_SECRET

.PHONY: help run build test test-unit test-race fmt fmt-check vet deps ci curl-demo clean
help:
	@echo "Task API: make run / make curl-demo / make ci"
	@echo "Исходный API: classroom12 из предыдущего занятия"
run:
	$(GO) run ./cmd/api
build:
	@mkdir -p bin
	$(GO) build -buildvcs=false -o bin/task-api ./cmd/api
test:
	$(GO) test ./...
test-unit:
	$(GO) test ./internal/...
test-race:
	$(GO) test -race ./...
fmt:
	gofmt -w $$(find . -type f -name '*.go' -not -path './bin/*')
fmt-check:
	@files="$$(gofmt -l $$(find . -type f -name '*.go' -not -path './bin/*'))"; \
	if [ -n "$$files" ]; then echo "$$files"; exit 1; fi
vet:
	$(GO) vet ./...
deps:
	$(GO) list -buildvcs=false -deps ./... >/dev/null
ci: fmt-check vet deps test test-race build
curl-demo:
	bash ./scripts/curl-demo.sh
clean:
	rm -f bin/task-api coverage.out
