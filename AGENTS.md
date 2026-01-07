# AGENTS.md

This file contains build commands, code style guidelines, and development conventions for the twchart project. It's designed to help agentic coding agents work effectively in this codebase.

## Build and Development Commands

### Running the Application
```bash
# Development server with data file
task dev
# or
go run cmd/twchart/main.go serve --store data.json

# Production server with directory
go run cmd/twchart/main.go serve --dir data/
```

### Testing
```bash
# Run all tests
go test ./...

# Run single test
go test -run TestParseLine ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

### Build and Lint
```bash
# Build the application
go build ./cmd/twchart

# Format code
go fmt ./...

# Run go vet (static analysis)
go vet ./...

# Install dependencies
go mod tidy
go mod download
```

### Docker
```bash
# Build Docker image
docker build -t twchart .

# Run with docker-compose
docker-compose up
```

## Code Style Guidelines

### Import Organization
- Group imports in three sections: standard library, third-party packages, and local packages
- Use blank lines between groups
- Local imports use the full module path: `github.com/calvinmclean/twchart`

Example:
```go
import (
    "fmt"
    "time"

    "github.com/go-echarts/go-echarts/v2/opts"
    "github.com/spf13/cobra"

    "github.com/calvinmclean/twchart"
)
```

### Naming Conventions
- **Packages**: lowercase, single word when possible (e.g., `api`, `twchart`)
- **Types**: PascalCase (e.g., `Session`, `Probe`, `ThermoworksData`)
- **Functions**: PascalCase for exported, camelCase for unexported
- **Variables**: camelCase, with descriptive names
- **Constants**: PascalCase for exported, camelCase for unexported
- **Interfaces**: Often end with `-er` suffix when possible (e.g., `io.Writer`)

### Error Handling
- Always handle errors explicitly
- Use `fmt.Errorf` with `%w` verb for error wrapping
- Return errors as the last return value
- Use descriptive error messages that include context

Example:
```go
if err != nil {
    return nil, fmt.Errorf("error parsing CSV: %w", err)
}
```

### Struct and Type Definitions
- Use PascalCase for exported types
- Include JSON tags where appropriate for API types
- Use pointer receivers for methods that modify the receiver
- Use value receivers for methods that don't modify the receiver

Example:
```go
type Session struct {
    Name      string    `json:"name"`
    Date      time.Time `json:"date"`
    StartTime time.Time `json:"startTime"`
    Probes    []Probe   `json:"probes"`
}

func (s *Session) LoadData(r io.Reader) error {
    // Modifies receiver, so use pointer
}
```

### Function Design
- Keep functions focused and small
- Use descriptive parameter names
- Return multiple values for error handling (result, error)
- Use interfaces for dependency injection when appropriate

### Constants and Enums
- Use `iota` for related constants
- Include descriptive comments for constant groups
- Use typed constants where possible

Example:
```go
const (
    ProbePositionNone = iota
    ProbePosition1
    ProbePosition2
    ProbePosition3
    ProbePosition4
    ProbePosition5
    probePositionInvalid
)

type ProbePosition uint
```

### Testing Conventions
- Test files end with `_test.go`
- Use table-driven tests for multiple test cases
- Use `github.com/stretchr/testify/assert` for assertions
- Test both success and error cases
- Use descriptive test names that indicate what's being tested

Example:
```go
func TestParseLine(t *testing.T) {
    t.Run("ParseName", func(t *testing.T) {
        // test implementation
    })
    
    t.Run("ParseProbePosition", func(t *testing.T) {
        // test implementation
    })
}
```

### Time Handling
- Use `time.Time` for all time values
- Be explicit about time zones (use `time.Local` for user input)
- Use `time.RFC3339` for serialization
- Use `time.Duration` for time spans

### Interface Implementation
- Declare interface compliance with `var _ Interface = &Type{}`
- Implement interfaces explicitly when they provide clear value
- Use composition over inheritance where possible

### Documentation
- Exported functions and types should have Go doc comments
- Use the godoc format: "FunctionName does..."
- Include parameter and return value descriptions for complex functions
- Add usage examples in comments for non-obvious functions

### Project Structure
- `cmd/twchart/`: Main application entry point
- `api/`: HTTP API handlers and server setup
- `twchart/`: Core business logic and data structures
- Root level: Main package files like `parse.go`, `chart.go`

### Dependencies
- Use Cobra for CLI commands
- Use babyapi for HTTP API framework
- Use go-echarts for charting
- Use testify for testing assertions

### Git and Version Control
- Conventional commits are preferred but not strictly required
- Keep commits focused on single changes
- Ensure all tests pass before committing
- Use `go mod tidy` before committing dependency changes

## Development Workflow

1. Make changes to code
2. Run `go fmt ./...` to format
3. Run `go vet ./...` for static analysis
4. Run `go test ./...` to ensure tests pass
5. Test the application manually if needed
6. Commit changes

## Key Patterns in This Codebase

- **Session parsing**: Text-based format parsing using regular expressions
- **Chart generation**: Using go-echarts to create time-series charts
- **API structure**: RESTful API using babyapi framework
- **Data loading**: CSV parsing with iterator pattern for memory efficiency
- **Time handling**: Support for both absolute timestamps and relative durations