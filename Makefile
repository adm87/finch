.PHONY: align lint test test-integration cover check check-integration check-verbose

## Local

align:
	betteralign -apply ./...

lint: align
	golangci-lint run ./...

test:
	go test ./...

test-integration:
	go test -tags integration ./...

cover:
	go test -coverprofile=cover.out ./pkg/...
	go tool cover -html=cover.out

## Dagger

check:
	dagger check -s

check-integration:
	dagger check --check=integration

check-verbose:
	dagger check --progress plain
