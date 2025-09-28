package audio

import "errors"

// ErrCommandNotFound indicates that the required command is not found in PATH
var ErrCommandNotFound = errors.New("command not found in PATH")

// ErrCommandExecutionFailed indicates that the command execution failed
var ErrCommandExecutionFailed = errors.New("command execution failed")

// ErrOutputParsingFailed indicates that parsing the command output failed
var ErrOutputParsingFailed = errors.New("output parsing failed")

// ErrMultipleCurrentDevices indicates that multiple devices are marked as current for the same direction
var ErrMultipleCurrentDevices = errors.New("multiple devices marked as current for the same direction")

// ErrNoDefaultDevice indicates that no default device is set for the requested direction
var ErrNoDefaultDevice = errors.New("no default device set")

// ErrInvalidCodecContainerCombination indicates an invalid codec/container combination
var ErrInvalidCodecContainerCombination = errors.New("invalid codec/container combination")

// ErrAudioCaptureFailed indicates that audio capture failed
var ErrAudioCaptureFailed = errors.New("audio capture failed")
