# Source Tree

## Project Structure

sttrouter is a Go CLI application for development workflow management.

```
sttrouter/
├── audio/                  # Audio device listing and capture implementations
│   ├── audio.go            # Audio conversion utilities
│   ├── device.go           # Device data structures and utilities
│   ├── errors.go           # Sentinel error definitions
│   ├── silence_splitter.go   # Silence detection for audio capture
│   ├── sox.go              # Shared sox types and structures
│   ├── sox_darwin.go       # macOS-specific sox audio capture (CoreAudio)
│   ├── sox_linux.go        # Linux-specific sox audio capture (PulseAudio)
│   ├── device_lister_darwin.go  # macOS device listing using system_profiler
│   └── device_lister_linux.go   # Linux device listing using pactl
├── clipboard/              # Clipboard operations
│   ├── clipboard_darwin.go # macOS clipboard (pbcopy)
│   └── clipboard_linux.go  # Linux clipboard (wl-copy/xclip)
├── cmd/                    # CLI commands (urfave/cli)
│   ├── capture.go          # capture command implementation
│   ├── config.go           # Global configuration structures
│   ├── format.go           # Output formatting utilities
│   ├── list_devices.go     # list-devices command implementation
│   ├── root.go             # Root command definition with global flags
│   └── transcribe.go       # transcribe command implementation
├── docs/                   # Documentation
│   ├── architecture/       # System architecture documentation
│   │   ├── coding-standards.md
│   │   └── principles.md
│   ├── prd.md
│   ├── source-tree.md
│   └── tech-stack.md
├── openaix/                # Azure OpenAI API client
│   └── transcription.go    # Transcription API client
├── .envrc
├── .gitignore
├── .golangci.yml
├── AGENTS.md               # AI agent configuration for OpenCode
├── LICENSE
├── Makefile
├── README.md
├── flake.lock
├── flake.nix
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── main.go                 # Application entry point
├── opencode.jsonc          # OpenCode IDE configuration
└── prompt.md
```

## Go File Specifications

### Entry Point

- **`main.go`** - Application entry point that delegates to cmd.Execute()

### CLI Command Layer (`cmd/`)

- **`root.go`** - Root urfave/cli command with global configuration and platform detection
  - Defines global flags: `--verbose`
  - Platform-aware default device source selection
  - Command description and help text
- **`list_devices.go`** - Implementation of the list-devices command
  - Uses dependency injection for device listers
  - Outputs device names with default indication
- **`capture.go`** - Implementation of the capture command
  - Records audio from microphone to file or stdout
  - Supports duration and format options
- **`transcribe.go`** - Implementation of the transcribe command
  - Captures audio and sends to Azure OpenAI for transcription
  - Supports various output modes (clipboard, stdout, file)
- **`config.go`** - Global configuration structures and validation
- **`format.go`** - Output formatting utilities

### Audio Package (`audio/`)

- **`audio.go`** - Audio conversion utilities (FLAC encoding)
- **`device.go`** - Device data structures and utilities
- **`errors.go`** - Sentinel error definitions
- **`silence_splitter.go`** - Silence detection for audio capture
- **`sox.go`** - Shared sox types and argument structures
- **`sox_darwin.go`** - macOS sox audio capture using CoreAudio driver
- **`sox_linux.go`** - Linux sox audio capture using PulseAudio driver
- **`device_lister_darwin.go`** - macOS device listing via system_profiler
  - Parses system_profiler JSON output for audio devices and defaults
- **`device_lister_linux.go`** - Linux device listing via pactl
  - Parses pactl output for PulseAudio devices and defaults

### Clipboard Package (`clipboard/`)

- **`clipboard_darwin.go`** - macOS clipboard using pbcopy
- **`clipboard_linux.go`** - Linux clipboard using wl-copy (Wayland) or xclip (X11)

### OpenAI Package (`openaix/`)

- **`transcription.go`** - Azure OpenAI API client for transcription

## Documentation Structure

### Core Documentation (`docs/`)

- **`architecture/`** - System architecture documentation
  - `coding-standards.md` - Coding standards and practices
  - `principles.md` - Architectural principles
- **`prd.md`** - Product requirements documentation
- **`tech-stack.md`** - Technology stack and dependencies
- **`source-tree.md`** - This file

## Key Directories

- **`audio/`** - Audio device listing and capture implementations with platform-specific backends
- **`clipboard/`** - Clipboard operations for output modes
- **`cmd/`** - urfave/cli commands with global configuration
- **`docs/`** - Living documentation with architecture, PRD, and technical details
- **`openaix/`** - Azure OpenAI API client for transcription

## Guidelines

### File Naming Conventions

- Commands: `{action}.go` (e.g., `list_devices.go`)
- Platform code: `{file}_{platform}.go` (e.g., `sox_darwin.go`, `sox_linux.go`)
- Tests: `{file}_test.go`
- Architecture docs: `{nn}-{topic}.md` (numbered for sequence)

### Implementation Patterns

- Separate files for platform-specific implementations
- Configuration via CLI flags and environment variables
- Use dependency injection pattern for testability