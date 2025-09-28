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
