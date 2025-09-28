# PRD

## Problem Statement

Current coding workflows require frequent context switching between spoken thoughts and manual typing, creating friction in the development process. When dictating comments, documentation, commit messages, or capturing ideas while coding, the need to physically type interrupts the flow of thought and slows down productivity. Existing speech-to-text solutions either require switching to different applications, don't integrate well with development environments, or lack the accuracy needed for technical content. This creates a gap where valuable spoken insights are either lost, poorly captured, or require disruptive workflow changes to record effectively.

The problem is particularly acute during:

- Code review sessions where spoken observations need to be documented
- Writing technical documentation or comments while in flow state
- Capturing quick notes or TODO items without breaking coding concentration
- Situations where hands are occupied or typing is inconvenient

## Proposed Solution

WhisperClip provides a flexible speech-to-text transcription system through two integrated components:

**Go CLI Application:** A versatile command-line tool that handles audio capture and manages API communication with OpenAI's Whisper API. The CLI offers multiple output modes to fit different workflows:

- **Clipboard mode:** Direct transcription to system clipboard
- **Stdout mode:** Pipe transcribed text to stdout for command-line workflows
- **File mode:** Write transcription directly to specified files

**VSCode Extension:** A development environment integration that provides intuitive activation and multiple insertion methods:

- **Phase 1:** Direct-to-clipboard transcription with manual paste
- **Phase 2:** Smart cursor insertion that places transcribed text at current cursor position
- **Integration:** Communicates with Go CLI and provides visual feedback for transcription status

**Key Differentiators:**

- **Output Flexibility:** Multiple modes support different use cases and workflows
- **Progressive Enhancement:** Start simple (clipboard) and evolve to smarter integration (cursor insertion)
- **CLI Versatility:** Standalone CLI utility that works beyond just VSCode
- **Cloud-Powered Accuracy:** Leverages OpenAI's Whisper API for superior transcription quality
- **Personal Optimization:** Tailored for single-user MacOS environment

The solution succeeds by providing flexibility in how transcribed text is delivered, allowing integration into various workflows while maintaining the core benefit of seamless speech-to-text conversion.

## Goals

- Enable speech-to-clipboard transcription during coding workflows without context switching
- Provide flexible output modes (clipboard, stdout, file) to support various development workflows
- Deliver accurate transcription using services like OpenAI Whisper API and Azure Speech for technical content and coding scenarios
- Create a personal productivity CLI tool optimized for single-user MacOS development environment
- Establish foundation for progressive enhancement of transcription capabilities

## Functional Requirements

### 1. System

**FR1.1:** The system MUST support macOS 12.0+ (Monterey and later).
**FR1.2:** The system MUST support verbose logging for debugging purposes.
**FR1.3:** The system MUST support structured logging in JSON format
**FR1.4:** The system MUST log to stderr to avoid interfering with stdout output
**FR1.5:** The system MUST not print sensitive information (such as API keys) in logs or error messages
**FR1.6:** The system MUST be configurable through flags and environment variables
**FR1.7:** The system MUST be built as a Command-Line Interface (CLI) application
**FR1.8:** The system MUST support outputs in both human (TXT) and machine-readable (JSON) formats
**FR1.9:** The system MUST produce TXT outputs that are easy to parse with Unix tools like grep, awk, and sed

### 2. CLI

**FR2.1:** The CLI must support the command `list-devices`, which lists available audio devices.
**FR2.2:** The CLI MUST support the command `capture`, which captures audio from a specified device.
**FR2.3:** The CLI MUST support the command `transcribe`, which transcribes captured audio using a transcription service.
**FR2.5:** The CLI MUST provide clear help and usage information for all commands and flags
**FR2.6:** The CLI SHOULD provide auto completion for Bash and Zsh shells.

### 3. Device listing & selection

**FR3.1:** Device listing MUST use device names compatible with `sox` (via `system_profiler`)
**FR3.2:** Device selection SHOULD use system defaults when available
**FR3.3:** Device listing MAY provide additional information such as device type, sample rate, and channel count
**FR3.4:** Device listing MUST use identifiers that are agnostic across device providers

### 4. Audio capture

**FR4.1:** Audio capture MUST use `sox`
**FR4.2:** Audio capture MUST use streaming audio formats (WAV or FLAC)
**FR4.3:** Audio capture MUST support output to file or stdout
**FR4.4:** Audio capture MUST support streaming to a transcription service
**FR4.5:** Audio capture MUST support manual start/stop
**FR4.6:** Audio capture MUST support duration-based capture
**FR4.7:** Audio capture MUST support asynchronous capture with signal-based stop

### 5. Transcription

**FR5.1:** Transcription MUST support the OpenAI Whisper API
**FR5.2:** Transcription MUST provide options for language selection and model customization
**FR5.3:** Transcription MUST support manual start/stop
**FR5.4:** Transcription MUST support output to file, stdout, or clipboard
**FR5.5:** Transcription MUST support streaming of audio capture to the transcription service
**FR5.6:** Transcription MAY support streaming of transcription results prior to the full completion of audio capture
**FR5.7:** Transcription MUST support asynchronous operation with signal-based stop

### 6. User Responsibilities

**FR6.1:** Users MUST create and maintain their own OpenAI account for access to Whisper API services
**FR6.2:** Users MUST obtain and configure their own OpenAI API keys via the OPENAI_API_KEY environment variable
**FR6.3:** Users are RESPONSIBLE for all OpenAI API usage costs and billing management
**FR6.4:** Users MUST ensure their OpenAI account has sufficient credit/billing setup before using transcription features
**FR6.5:** Users MUST comply with OpenAI's terms of service and usage policies when using transcription features
**FR6.6:** Users MUST secure their API keys and are responsible for any unauthorized usage
**FR6.7:** Users SHOULD monitor their OpenAI API usage to manage costs and avoid unexpected charges
