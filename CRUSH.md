# CRUSH Development Guide

## Build/Test/Lint Commands
```bash
# Build the application
go build -o omnia .

# Run the application
go run main.go

# Test (when tests exist)
go test ./...
go test -v ./...                    # verbose output
go test -run TestFunctionName ./... # run specific test

# Format code
go fmt ./...
gofumpt -w .  # if available

# Lint (install golangci-lint first)
golangci-lint run

# Vet for common issues
go vet ./...
```

## Code Style Guidelines

### Imports
- Use `goimports` for automatic import management
- Group imports: stdlib, external packages, local packages
- Use package aliases sparingly (e.g., `tea "github.com/charmbracelet/bubbletea"`)

### Naming Conventions
- Use camelCase for local variables and functions
- Use PascalCase for exported functions and types
- Keep struct field names descriptive but concise
- Model structs should be simple and focused

### Error Handling
- Always handle errors explicitly
- Use `log.Fatal()` for unrecoverable errors in main
- Return errors up the call stack when possible

### Bubbletea Patterns
- Follow the Elm Architecture: Model, Update, View
- Keep models simple with clear state representation
- Use type switches in Update for message handling
- Style components with lipgloss for consistent UI