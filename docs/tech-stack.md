# Technology Stack

## Core Technologies

- **Go 1.24.x**: Primary language for CLI implementation
- **urfave/cli**: Command-line interface structure and flag management
- **Sox**: External subprocess for audio capture
- **macOS system_profiler**: macOS system utility for device enumeration
- **Azure OpenAI GPT-4o**: Remote transcription service

## Standard Library Dependencies

- **net/http**: HTTP client for Azure OpenAI API communication
- **context**: Control flow and cancellation management
- **io**: Streaming data interfaces (Reader/Writer)
- **os/exec**: Sox and system_profiler subprocess management
- **encoding/json**: API request/response handling and system_profiler JSON parsing
- **strings**: String manipulation for parsing command outputs
- **os/signal**: Signal handling for async capture mode

## Testing and Development

- **testify**: Test assertions and mocking framework
- **slog**: Structured logging (Go 1.24.x standard)

## External Dependencies

- **Sox**: Must be available in system PATH for audio capture
- **macOS system_profiler**: macOS system utility for device enumeration
