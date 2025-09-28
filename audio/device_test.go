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
				{Name: "Device A", Mode: DeviceFlagInput},
				{Name: "Device B", Mode: DeviceFlagOutput},
			},
			spDevices: []Device{
				{Name: "Device A", Mode: DeviceFlagOutput},
				{Name: "Device C", Mode: DeviceFlagInput},
			},
			expected: []Device{
				{Name: "Device A", Mode: DeviceFlagInput | DeviceFlagOutput},
				{Name: "Device B", Mode: DeviceFlagOutput},
				{Name: "Device C", Mode: DeviceFlagInput},
			},
			expectedErr: nil,
		},
		{
			name: "multiple current input devices",
			ffmpegDevices: []Device{
				{Name: "Device A", Mode: DeviceFlagCurrentInput},
				{Name: "Device B", Mode: DeviceFlagCurrentInput},
			},
			spDevices:   []Device{},
			expected:    nil,
			expectedErr: ErrMultipleCurrentDevices,
		},
		{
			name: "multiple current output devices",
			ffmpegDevices: []Device{
				{Name: "Device A", Mode: DeviceFlagCurrentOutput},
				{Name: "Device B", Mode: DeviceFlagCurrentOutput},
			},
			spDevices:   []Device{},
			expected:    nil,
			expectedErr: ErrMultipleCurrentDevices,
		},
		{
			name: "single current input and output devices",
			ffmpegDevices: []Device{
				{Name: "Input Device", Mode: DeviceFlagCurrentInput},
				{Name: "Output Device", Mode: DeviceFlagCurrentOutput},
			},
			spDevices: []Device{},
			expected: []Device{
				{Name: "Input Device", Mode: DeviceFlagCurrentInput},
				{Name: "Output Device", Mode: DeviceFlagCurrentOutput},
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
