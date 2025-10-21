# Agent Development Guide

## Build/Test/Lint Commands
```bash
go build -o omnia .
go run main.go
go test ./...
go test -run TestFunctionName ./...
go fmt ./...
go vet ./...
```

## Code Style Guidelines

**Imports**: Group as stdlib, external, local. Use aliases sparingly (e.g., `tea "github.com/charmbracelet/bubbletea"`).

**Naming**: camelCase for local vars/funcs, PascalCase for exported. Keep struct fields descriptive but concise.

**Error Handling**: Always handle errors explicitly. Use `log.Fatal()` for unrecoverable errors in main. Return errors up the call stack.

**Bubbletea Patterns**: Follow Elm Architecture (Model/Update/View). Keep models simple. Use type switches in Update for message handling. Style with lipgloss.

**Comments**: Minimal. Only add when logic is non-obvious or for TODOs.

**Types**: Define custom types (like `Shell`, `completion`) for clarity. Implement required interfaces (e.g., `list.Item`).
