# Coding Standards

This document defines the coding standards, patterns, and practices specific to sttrouter development. For general Go best practices, refer to [Effective Go](https://go.dev/doc/effective_go), [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments), and [Google's Go Style Guide](https://google.github.io/styleguide/go/).

## Project-Specific Standards

### CLI Application Structure

- Use Cobra CLI framework for command structure and flag handling
- Implement commands as separate files in `cmd/` package
- Follow the composable streaming pipeline architecture (see section 03)
- Separate control flow (context) from data flow (io interfaces)
- Use dependency injection for testing and flexibility

**Interface Guidelines:**

- Keep interfaces small and focused (single responsibility)
- Define interfaces where they're used, not where they're implemented
- Use `-er` suffix naming convention

### Configuration Management

- Centralize configuration in a single `Config` struct
- Support both command-line flags and environment variables
- Use struct tags for configuration binding
- Validate configuration at startup with sensible defaults

### Error Handling

- Follow streaming error pattern through `TranscriptionResult`
- **Prefer sentinel error variables** over custom error structs:

  ```go
  var errNotFound = errors.New("not found")
  fmt.Errorf("article %s: %w", id, errNotFound)
  ```

- Separate operational errors (stderr) from data errors (in results)
- Never ignore errors without documented reasoning
- Define errors in `errors.go`

## Go Development Essentials

### Code Style

- Use `gofmt` and `goimports` for formatting
- Write clear, idiomatic Go code
- Keep the happy path left-aligned (minimize indentation)
- Return early to reduce nesting
- Document all exported symbols

### Package Organization

- Follow standard Go project layout
- Use `cmd/` for main packages
- Prefer root-level packages for packages beyond `cmd/`
- Avoid circular dependencies
- Choose descriptive package names (avoid `util`, `common`)

### Testing Standards

- Use table-driven tests for multiple scenarios
- Name tests: `Test_functionName_scenario`
- Place tests next to code being tested
- Mark helpers with `t.Helper()`
- Test both success and error cases

### Development Workflow

**Required Tools:**

- `golangci-lint` for comprehensive linting
- `go test` with race detection: `go test -race`
- `go mod tidy` for dependency management

**Pre-commit Checklist:**

- Run `golangci-lint run`
- Run `go test -race ./...`
- Ensure `go mod tidy` is clean
- Verify all exported symbols are documented

### Common Go Patterns

**Error Handling:**

```go
// Check errors immediately
result, err := someFunction()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Use errors.Is/As for detection
if errors.Is(err, errNotFound) {
    // handle not found
}
```

**Resource Management:**

```go
// Always use defer for cleanup
file, err := os.Open(filename)
if err != nil {
    return err
}
defer file.Close()
```

**Interface Design:**

```go
// Accept interfaces, return concrete types
func ProcessAudio(capturer AudioCapturer) *AudioResult {
    // implementation
}
```

## Security & Performance

### Input Validation

- Validate all external input before processing
- Use strong typing to prevent invalid states
- Be cautious with file paths from user input

### Performance

- Minimize allocations in hot paths
- Use `sync.Pool` for object reuse when appropriate
- Profile before optimizing (`go tool pprof`)
- Preallocate slices when size is known

### Concurrency

- Don't create goroutines in libraries; let callers control concurrency
- Always ensure goroutines can exit (avoid leaks)
- Use channels for communication, mutexes for protecting state
- Close channels from sender side only
