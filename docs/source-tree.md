# Source Tree

## Project Structure

sttrouter is a Go CLI application with BMAD methodology infrastructure for development workflow management.

```
sttrouter/
├── audio/                  # Audio device listing implementations
│   ├── errors.go          # Sentinel error definitions
│   ├── ffmpeg.go          # FFmpeg-based device listing
│   ├── system_profiler.go # macOS system_profiler-based device listing
│   ├── ffmpeg_test.go     # Tests for ffmpeg parsing
│   └── system_profiler_test.go # Tests for system_profiler parsing
├── cmd/                     # CLI commands (Cobra)
│   ├── list_devices.go    # list-devices command implementation
│   └── root.go             # Root command definition with global flags
├── docs/                  # Documentation (architecture, prd, qa, stories)
│   ├── architecture/      # System architecture documentation
│   │   ├── coding-standards.md
│   │   └── principles.md
│   ├── prd.md
│   ├── source-tree.md
│   └── tech-stack.md
├── .envrc
├── .gitignore
├── .golangci.yml
├── AGENTS.md              # AI agent configuration for OpenCode
├── Makefile
├── README.md
├── flake.lock
├── flake.nix
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── main.go                # Application entry point
├── opencode.jsonc         # OpenCode IDE configuration
└── prompt.md
```

## Go File Specifications

### Entry Point

- **`main.go`** - Application entry point that delegates to cmd.Execute()

### CLI Command Layer (`cmd/`)

- **`root.go`** - Root Cobra command with global configuration and platform detection
   - Defines global flags: `--verbose`
   - Platform-aware default device source selection
   - Command description and help text
- **`list_devices.go`** - Implementation of the list-devices command
  - Uses dependency injection for device listers
  - Outputs device names with default indication

### Audio Package (`audio/`)

- **`ffmpeg.go`** - FFmpeg-based device listing implementation
  - Parses ffmpeg avfoundation output for audio devices
- **`system_profiler.go`** - macOS system_profiler-based device listing implementation
  - Parses system_profiler JSON output for audio devices and defaults

## Documentation Structure

### Core Documentation (`docs/`)

- **`architecture/`** - System architecture documentation
  - `coding-standards.md` - Coding standards and practices
  - `principles.md` - Architectural principles
- **`prd.md`** - Product requirements documentation
- **`tech-stack.md`** - Technology stack and dependencies
- **`source-tree.md`** - This file

## Key Directories

- **`audio/`** - Audio device listing implementations with platform-specific backends
- **`cmd/`** - Cobra CLI commands with global configuration
- **`docs/`** - Living documentation with architecture, PRD, and technical details

## Guidelines

### File Naming Conventions

- Commands: `{action}.go` (e.g., `list_devices.go`)
- Platform code: `{file}_{platform}.go` (future: `devices_darwin.go`)
- Tests: `{file}_test.go`
- Architecture docs: `{nn}-{topic}.md` (numbered for sequence)

### Implementation Patterns

- Separate files for platform-specific implementations
- Configuration via CLI flags and environment variables
- Follow streaming pipeline architecture for future audio processing
- Use dependency injection pattern for testability

### Development Workflow

- Architecture-first approach with living documentation
- Quality gates at story completion
- Test-driven development with comprehensive coverage
- AI-assisted development using BMAD methodology agents