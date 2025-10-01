# Technology Stack

## Core Technologies

- **Go 1.24.x**: Primary language for CLI implementation
- **urfave/cli**: Command-line interface structure and flag management
- **Sox**: External subprocess for audio capture
- **Azure OpenAI GPT-4o**: Remote transcription service

## Platform-Specific Technologies

### macOS
- **system_profiler**: macOS command-line utility used by DeviceLister for device enumeration
- **CoreAudio**: Audio driver backend for Sox
- **pbcopy**: System clipboard utility

### Linux
- **pactl (PulseAudio)**: Command-line utility used by DeviceLister for device enumeration and management
- **PulseAudio**: Audio driver backend for Sox
- **wl-copy/xclip**: Clipboard utilities for Wayland/X11

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

### Common (All Platforms)
- **Sox**: Must be available in system PATH for audio capture

### macOS
- **system_profiler**: Built-in macOS command-line utility used for device enumeration
- **pbcopy**: Built-in macOS clipboard utility

### Linux
- **pactl**: PulseAudio command-line utility used for device enumeration (usually pre-installed)
- **PulseAudio**: Audio server (usually pre-installed)
- **wl-copy** (Wayland) or **xclip** (X11): Clipboard utilities (require installation)
