# Prompto Agent Guide

## Build/Test Commands
- Build: `make build` or `go build ./...`
- Test all: `make test` or `go test ./...`
- Test single package: `go test ./pkg/repositories/...`
- Test single test: `go test -run TestFunctionName ./pkg/path`
- Lint: `make lint` or `golangci-lint run -v`
- Install: `make install`

## Code Style Guidelines
- Go version: 1.24+
- Imports: Group standard library, then third-party, then internal packages
- Naming: Follow Go conventions (CamelCase for exported, camelCase for unexported)
- Types: Define custom types for better semantics (see FileType example)
- Format: Use gofmt/goimports
- Error checking: All errors must be handled (see .golangci.yml)
- Use zerolog for logging (`github.com/rs/zerolog`)


<goGuidelines>
When implementing go interfaces, use the var _ Interface = &Foo{} to make sure the interface is always implemented correctly.
When building web applications, use htmx, bootstrap and the templ templating language.
Always use a context argument when appropriate.
Use cobra for command-line applications.
Use the "defaults" package name, instead of "default" package name, as it's reserved in go.
Use github.com/pkg/errors for wrapping errors.
When starting goroutines, use errgroup.
</goGuidelines>