# sttrouter

A macOS-only Go CLI tool for speech-to-text transcription using the Azure OpenAI hosted version of GPT-4o.

## How it works

Captures audio from microphone using Sox, converts to FLAC, sends to Azure OpenAI GPT-4o for transcription, outputs to clipboard, stdout, or file.

## Prerequisites

- macOS 12.0+
- Sox (audio capture)
- Azure OpenAI API key

## Installation

### From source

Requires Go 1.24+:

```bash
go install github.com/sebnyberg/sttrouter@latest
```

### Pre-built binaries

Not available yet.

## Usage

### Commands

- `list-devices`: List available audio input devices
- `capture`: Record audio to file
- `transcribe`: Transcribe audio file or capture from microphone

### Examples

```bash
# List devices
sttrouter list-devices

# Capture audio
sttrouter capture --duration 5s output.flac

# Transcribe file
sttrouter transcribe --api-key YOUR_AZURE_KEY file.flac

# Capture and transcribe from microphone
sttrouter transcribe --capture --api-key YOUR_AZURE_KEY

# Capture and copy to clipboard
sttrouter transcribe --capture --api-key YOUR_AZURE_KEY --clipboard

# Azure OpenAI example (default configuration)
sttrouter transcribe --capture --api-key YOUR_AZURE_KEY --base-url https://your-resource.openai.azure.com/openai/deployments/{deployment_id} --query-params "api-version=2025-03-01-preview"
```

Set API key as environment variable:

```bash
export API_KEY="your-azure-openai-key"
sttrouter transcribe --capture
```

## Technology

- Go 1.24
- Sox for audio capture
- Azure OpenAI GPT-4o (hosted version, currently only supported)
- urfave/cli for CLI framework

## Restrictions

- macOS only
- Users responsible for API costs and compliance with Azure OpenAI terms
- Requires API key configuration
