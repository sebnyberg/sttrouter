package audio

import "errors"

// ErrCommandExecutionFailed indicates that the command execution failed
var ErrCommandExecutionFailed = errors.New("command execution failed")

// ErrOutputParsingFailed indicates that parsing the command output failed
var ErrOutputParsingFailed = errors.New("output parsing failed")

// ErrNoDefaultDevice indicates that no default device is set for the requested direction
var ErrNoDefaultDevice = errors.New("no default device set")

// ErrInvalidCodecContainerCombination indicates an invalid codec/container combination
var ErrInvalidCodecContainerCombination = errors.New("invalid codec/container combination")

// ErrAudioCaptureFailed indicates that audio capture failed
var ErrAudioCaptureFailed = errors.New("audio capture failed")
