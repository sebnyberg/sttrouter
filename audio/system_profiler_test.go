package audio

import (
	"reflect"
	"testing"
)

// systemProfilerOutputSample is a sample JSON output from system_profiler SPAudioDataType -json
const systemProfilerOutputSample = `{
  "SPAudioDataType" : [
    {
      "_items" : [
        {
          "_name" : "DELL U2724DE",
          "coreaudio_device_manufacturer" : "DEL",
          "coreaudio_device_output" : 2,
          "coreaudio_device_srate" : 48000,
          "coreaudio_device_transport" : "coreaudio_device_type_displayport",
          "coreaudio_output_source" : "spaudio_default"
        },
        {
          "_name" : "INZONE H3",
          "_properties" : "coreaudio_default_audio_system_device",
          "coreaudio_default_audio_input_device" : "spaudio_yes",
          "coreaudio_default_audio_output_device" : "spaudio_yes",
          "coreaudio_default_audio_system_device" : "spaudio_yes",
          "coreaudio_device_input" : 1,
          "coreaudio_device_manufacturer" : "Sony",
          "coreaudio_device_output" : 2,
          "coreaudio_device_srate" : 48000,
          "coreaudio_device_transport" : "coreaudio_device_type_usb",
          "coreaudio_input_source" : "spaudio_default",
          "coreaudio_output_source" : "spaudio_default"
        },
        {
          "_name" : "C922 Pro Stream Webcam",
          "coreaudio_device_input" : 2,
          "coreaudio_device_manufacturer" : "Unknown Manufacturer",
          "coreaudio_device_srate" : 16000,
          "coreaudio_device_transport" : "coreaudio_device_type_usb",
          "coreaudio_input_source" : "spaudio_default"
        },
        {
          "_name" : "MacBook Pro Microphone",
          "coreaudio_device_input" : 1,
          "coreaudio_device_manufacturer" : "Apple Inc.",
          "coreaudio_device_srate" : 48000,
          "coreaudio_device_transport" : "coreaudio_device_type_builtin",
          "coreaudio_input_source" : "MacBook Pro Microphone"
        },
        {
          "_name" : "MacBook Pro Speakers",
          "coreaudio_device_manufacturer" : "Apple Inc.",
          "coreaudio_device_output" : 2,
          "coreaudio_device_srate" : 48000,
          "coreaudio_device_transport" : "coreaudio_device_type_builtin",
          "coreaudio_output_source" : "MacBook Pro Speakers"
        },
        {
          "_name" : "Microsoft Teams Audio",
          "coreaudio_device_input" : 1,
          "coreaudio_device_manufacturer" : "Microsoft Corp.",
          "coreaudio_device_output" : 1,
          "coreaudio_device_srate" : 48000,
          "coreaudio_device_transport" : "coreaudio_device_type_virtual",
          "coreaudio_input_source" : "Microsoft Teams Audio Device",
          "coreaudio_output_source" : "Microsoft Teams Audio Device"
        }
      ],
      "_name" : "coreaudio_device"
    }
  ]
}`

func TestParseSystemProfilerDevices(t *testing.T) {
	expectedDevices := []Device{
		{Name: "DELL U2724DE", Mode: DeviceFlagOutput},
		{Name: "INZONE H3", Mode: DeviceFlagInput | DeviceFlagOutput | DeviceFlagIsDefault},
		{Name: "C922 Pro Stream Webcam", Mode: DeviceFlagInput},
		{Name: "MacBook Pro Microphone", Mode: DeviceFlagInput},
		{Name: "MacBook Pro Speakers", Mode: DeviceFlagOutput},
		{Name: "Microsoft Teams Audio", Mode: DeviceFlagInput | DeviceFlagOutput},
	}
	expectedDefault := "INZONE H3"
	actualDevices, actualDefault, err := parseSystemProfilerDevices([]byte(systemProfilerOutputSample))
	if err != nil {
		t.Fatalf("parseSystemProfilerDevices() error = %v", err)
	}
	if !reflect.DeepEqual(actualDevices, expectedDevices) {
		t.Errorf("parseSystemProfilerDevices() devices = %v, want %v", actualDevices, expectedDevices)
	}
	if actualDefault != expectedDefault {
		t.Errorf("parseSystemProfilerDevices() default = %v, want %v", actualDefault, expectedDefault)
	}
}
