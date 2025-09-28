# sttrouter

A Go CLI application for speech-to-text transcription, designed to enable voice-driven text input in development workflows.

## Overview

sttrouter combines audio capture capabilities with transcription services to provide a speech-to-text solution. The tool supports multiple output modes (clipboard, stdout, file) and is designed for integration into development workflows.

## Features

- **Device Listing**: List available audio input devices with default indication
- **Audio Capture**: Record from system microphone with configurable duration
- **Transcription Service Integration**: Configurable transcription services for speech-to-text conversion
- **Multiple Output Modes**:
  - Clipboard mode for direct system clipboard integration
  - Stdout mode for command-line workflows and piping
  - File mode for direct file output
- **Cross-platform Audio Support**: Uses Sox for audio capture
- **VSCode Extension Ready**: Designed for integration with development environments

## Prerequisites

- **Sox**: Required for audio capture and processing
- **Transcription service credentials**: API keys for your chosen transcription service

## Installation

### Pre-built Binaries

**Note: Releases not done yet**

Download the latest release from [GitHub Releases](https://github.com/sebnyberg/sttrouter/releases).

```bash
# macOS example
curl -L https://github.com/sebnyberg/sttrouter/releases/download/v1.0.0/sttrouter-darwin-arm64 -o sttrouter
chmod +x sttrouter
sudo mv sttrouter /usr/local/bin/
```

### From Source (Recommended)

If you have Go 1.24+ installed:

```bash
go install github.com/sebnyberg/sttrouter@latest
```

## Setup

1. Configure your transcription service API credentials as required by your chosen service.

2. Verify the installation:
   ```bash
   sttrouter --help
   ```

## Development Setup

For contributors who want to build and develop the project:

### Prerequisites

- **Go 1.24.x**: Primary development language
- **Nix**: Reproducible development environment (optional but recommended)
- **direnv**: Automatic environment loading (optional but recommended)
- **FFmpeg**: Audio capture and processing (installed via Nix or system PATH)

### Option 1: Using Nix (Recommended for Development)

1. Install Nix package manager:
   ```bash
   curl -L https://nixos.org/nix/install | sh
   ```

2. Install direnv:
   ```bash
   # macOS with Homebrew
   brew install direnv

   # Or manually
   curl -sfL https://direnv.net/install.sh | bash
   ```

3. Clone the repository and enter the directory:
   ```bash
   git clone https://github.com/sebnyberg/sttrouter.git
   cd sttrouter
   ```

4. Allow direnv to load the development environment:
   ```bash
   direnv allow
   ```

The Nix environment will automatically provide Go, FFmpeg, and golangci-lint in your development environment.

### Option 2: Manual Development Setup

1. Install Go 1.24.x from [golang.org](https://golang.org/dl/)
2. Install Sox:
   ```bash
   # macOS with Homebrew
   brew install sox
   ```
3. Clone and build:
   ```bash
   git clone https://github.com/sebnyberg/sttrouter.git
   cd sttrouter
   go mod download
   ```

## Usage

### Basic Commands

```bash
# List available audio devices
sttrouter list-devices

# Record and transcribe (clipboard output)
sttrouter transcribe --duration 5s

# Record and transcribe (stdout output)
sttrouter transcribe --duration 5s --output stdout

# Record and transcribe to file
sttrouter transcribe --duration 5s --output file --file output.txt
```

### Build Instructions

The project uses a Makefile for common development tasks. Run `make help` to see all available commands.

### Basic Build

```bash
# Build the application
make build

# The binary will be created at bin/sttrouter
```

### Testing

```bash
# Run all tests with race detection
make test
```

### Linting

```bash
# Run golangci-lint to check code quality
make lint
```

The project uses golangci-lint with a configuration that focuses on practical code quality improvements while avoiding unhelpful complexity metrics. The linter checks for:

- Static analysis issues (staticcheck)
- Go vet warnings (govet)
- Unused variables/constants/functions (unused)
- Ineffectual assignments (ineffassign)
- Unchecked errors (errcheck, excluded from test files)
- Misspelled words (misspell)

Complexity metrics like cyclomatic complexity and cognitive complexity are explicitly disabled as they often penalize valid code patterns.

## Development Workflow

### Environment Setup

The project uses Nix for reproducible development environments. When you enter the project directory, direnv automatically loads the development shell with all required tools.

The Nix environment provides:
- **Go 1.24.x**: Primary development language
- **golangci-lint**: Code linting and quality checks
- **Sox**: Audio processing for runtime functionality
- **Git & curl**: Development utilities

### Available Make Commands

```bash
make help     # Show all available commands
make build    # Build the sttrouter binary
make test     # Run all tests with race detection
make lint     # Run golangci-lint
make clean    # Remove build artifacts
```

### Code Quality

Before committing code:
1. Run `make test` to ensure all tests pass
2. Run `make lint` to check code quality
3. Run `make build` to verify the code compiles

The project follows Go best practices as outlined in [Effective Go](https://go.dev/doc/effective_go) and the implementation standards documented in `docs/architecture/06-implementation-standards.md`.

## Contributing

### Development Requirements

- **Go 1.24.x**: Primary development language
- **Nix**: Reproducible development environment (optional but recommended)
- **direnv**: Automatic environment loading (optional but recommended)
- **FFmpeg**: Audio capture and processing (installed via Nix or system PATH)

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes with tests
4. Run the full test suite: `make test`
5. Ensure code passes linting: `make lint`
6. Submit a pull request

### Development Standards

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use table-driven tests for multiple test cases
- Document all exported symbols
- Keep interfaces small and focused
- Prefer sentinel errors over custom error types

## License

_TODO: Add license information_

## Related Projects

- [OpenAI Whisper API](https://platform.openai.com/docs/guides/speech-to-text) (one supported transcription service)
- [Azure Speech Service](https://azure.microsoft.com/en-us/products/ai-services/ai-speech) (alternative transcription service)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Sox](https://sox.sourceforge.net/)