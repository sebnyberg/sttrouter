You are working with me on the `list-devices` command. We have finished the implementation of an `audio` package that supports listing devices both with ffmpeg and system_profiler, in addition to finding the default device with system_profiler.

The list-devices command should list devices with both ffmpeg and system_profiler. Then verify that the two lists contain the same devices (eg sort and compare), then print the devices in lexicographical ordering with the default device marked in the output.

The output should be friendly to parse with e.g. AWK. This means printing a header with DEVICE_NAME and DEFAULT columns, then one line per device with the device name and whether it is the default device (true/false).
