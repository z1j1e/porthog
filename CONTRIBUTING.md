# Contributing to porthog

## Development Setup

1. Install Go 1.24+
2. Clone the repository
3. Run `go mod download`
4. Build: `CGO_ENABLED=0 go build -o porthog ./cmd/porthog/`
5. Test: `go test ./... -v`

## Project Structure

```
cmd/porthog/          CLI entry point (Cobra commands)
internal/core/
  domain/             Entities and value objects
  ports/              Interface definitions
  services/           Business logic (use cases)
internal/adapters/
  os/{linux,darwin,windows}/  Platform-specific port enumeration
  platform/           Platform factory (build-tag routing)
  process/            Process metadata enrichment (gopsutil)
  output/             Renderers (table, JSON, plain)
internal/tui/watch/   Bubbletea TUI for watch mode
internal/config/      Configuration loading
```

## Guidelines

- Keep platform-specific code in `internal/adapters/os/`
- Core logic must not import platform packages
- Use build tags for OS-specific files
- Run `go vet ./...` before submitting
- Add tests for new functionality

## License

MIT
