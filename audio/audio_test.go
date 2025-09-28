package audio

import (
	"errors"
	"reflect"
	"testing"
)

func TestListDevices(t *testing.T) {
	tests := []struct {
		name          string
		ffmpegDevices []Device
		spDevices     []Device
		expected      []Device
		expectedErr   error
	}{
		{
			name: "merge and sort devices",
			ffmpegDevices: []Device{
				{Name: "Device A", Mode: DeviceFlagSource},
				{Name: "Device B", Mode: DeviceFlagSink},
			},
			spDevices: []Device{
				{Name: "Device A", Mode: DeviceFlagSink},
				{Name: "Device C", Mode: DeviceFlagSource},
			},
			expected: []Device{
				{Name: "Device A", Mode: DeviceFlagSource | DeviceFlagSink},
				{Name: "Device B", Mode: DeviceFlagSink},
				{Name: "Device C", Mode: DeviceFlagSource},
			},
			expectedErr: nil,
		},
		{
			name: "multiple current source devices",
			ffmpegDevices: []Device{
				{Name: "Device A", Mode: DeviceFlagCurrentSource},
				{Name: "Device B", Mode: DeviceFlagCurrentSource},
			},
			spDevices:   []Device{},
			expected:    nil,
			expectedErr: ErrMultipleCurrentDevices,
		},
		{
			name: "multiple current sink devices",
			ffmpegDevices: []Device{
				{Name: "Device A", Mode: DeviceFlagCurrentSink},
				{Name: "Device B", Mode: DeviceFlagCurrentSink},
			},
			spDevices:   []Device{},
			expected:    nil,
			expectedErr: ErrMultipleCurrentDevices,
		},
		{
			name: "single current source and sink devices",
			ffmpegDevices: []Device{
				{Name: "Source Device", Mode: DeviceFlagCurrentSource},
				{Name: "Sink Device", Mode: DeviceFlagCurrentSink},
			},
			spDevices: []Device{},
			expected: []Device{
				{Name: "Sink Device", Mode: DeviceFlagCurrentSink},
				{Name: "Source Device", Mode: DeviceFlagCurrentSource},
			},
			expectedErr: nil,
		},
		{
			name:          "empty device lists",
			ffmpegDevices: []Device{},
			spDevices:     []Device{},
			expected:      []Device{},
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ListDevices(tt.ffmpegDevices, tt.spDevices)

			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error %v but got none", tt.expectedErr)
				} else if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(result, tt.expected) {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}
