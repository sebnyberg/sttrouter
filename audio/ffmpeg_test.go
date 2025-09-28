package audio

import (
	"reflect"
	"testing"
)

// ffmpegOutputSample is a sample output from ffmpeg -f avfoundation -list_devices true -i ""
const ffmpegOutputSample = `ffmpeg version 7.1.1 Copyright (c) 2000-2025 the FFmpeg developers
  built with clang version 19.1.7
  configuration: --disable-static --prefix=/nix/store/eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee-ffmpeg-7.1.1 --target_os=darwin --arch=aarch64 --pkg-config=pkg-config --enable-gpl --enable-version3 --disable-nonfree --disable-static --enable-shared --enable-pic --disable-thumb --disable-small --enable-runtime-cpudetect --disable-gray --enable-swscale-alpha --enable-hardcoded-tables --enable-safe-bitstream-reader --enable-pthreads --disable-w32threads --disable-os2threads --enable-network --enable-pixelutils --datadir=/nix/store/snkhcygf72yv9qg2a47haim7zy00gg99-ffmpeg-7.1.1-data/share/ffmpeg --enable-ffmpeg --enable-ffplay --enable-ffprobe --bindir=/nix/store/eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee-ffmpeg-7.1.1-bin/bin --enable-avcodec --enable-avdevice --enable-avfilter --enable-avformat --enable-avutil --enable-postproc --enable-swresample --enable-swscale --libdir=/nix/store/eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee-ffmpeg-7.1.1-lib/lib --incdir=/nix/store/eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee-ffmpeg-7.1.1-dev/include --enable-doc --enable-htmlpages --enable-manpages --mandir=/nix/store/eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee-ffmpeg-7.1.1-man/share/man --enable-podpages --enable-txtpages --docdir=/nix/store/eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee-ffmpeg-7.1.1-doc/share/doc/ffmpeg --disable-alsa --disable-amf --enable-libaom --disable-libaribb24 --disable-libaribcaption --enable-libass --disable-avisynth --enable-libbluray --disable-libbs2b --enable-bzlib --disable-libcaca --disable-libcdio --disable-libcelt --disable-chromaprint --disable-libcodec2 --disable-cuda --enable-cuda-llvm --disable-cuda-nvcc --disable-cuvid --enable-libdav1d --disable-libdavs2 --disable-libdc1394 --disable-libdrm --disable-libdvdnav --disable-libdvdread --disable-libfdk-aac --disable-ffnvcodec --disable-libflite --enable-fontconfig --enable-libfontconfig --enable-libfreetype --disable-frei0r --enable-libfribidi --disable-libgme --enable-gmp --enable-gnutls --disable-libgsm --enable-libharfbuzz --enable-iconv --disable-libilbc --disable-libjack --disable-libjxl --disable-libkvazaar --disable-ladspa --disable-liblc3 --disable-liblcevc-dec --disable-lcms2 --enable-lzma --disable-metal --disable-libmfx --disable-libmodplug --enable-libmp3lame --disable-libmysofa --disable-libnpp --disable-nvdec --disable-nvenc --disable-openal --enable-opencl --disable-libopencore-amrnb --disable-libopencore-amrwb --disable-opengl --disable-libopenh264 --enable-libopenjpeg --enable-libopenmpt --enable-libopus --disable-libplacebo --disable-libpulse --disable-libqrencode --disable-libquirc --disable-librav1e --enable-librist --disable-librtmp --disable-librubberband --disable-libsmbclient --enable-sdl2 --disable-libshaderc --disable-libshine --disable-libsnappy --enable-libsoxr --enable-libspeex --enable-libsrt --enable-libssh --disable-librsvg --enable-libsvtav1 --disable-libtensorflow --enable-libtheora --disable-libtwolame --disable-libuavs3d --disable-libv4l2 --disable-v4l2-m2m --disable-vaapi --enable-vdpau --disable-libvpl --enable-libvidstab --disable-libvmaf --disable-libvo-amrwbenc --enable-libvorbis --enable-libvpx --disable-vulkan --disable-libvvenc --enable-libwebp --enable-libx264 --enable-libx265 --disable-libxavs --disable-libxavs2 --disable-libxcb --disable-libxcb-shape --disable-libxcb-shm --disable-libxcb-xfixes --disable-libxevd --disable-libxeve --disable-xlib --enable-libxml2 --enable-libxvid --enable-libzimg --enable-zlib --disable-libzmq --enable-libzvbi --disable-debug --enable-optimizations --disable-extra-warnings --disable-stripping --cc=clang --cxx=clang++
  libavutil      59. 39.100 / 59. 39.100
  libavcodec     61. 19.101 / 61. 19.101
  libavformat    61.  7.100 / 61.  7.100
  libavdevice    61.  3.100 / 61.  3.100
  libavfilter    10.  4.100 / 10.  4.100
  libswscale      8.  3.100 /  8.  3.100
  libswresample   5.  3.100 /  5.  3.100
  libpostproc    58.  3.100 / 58.  3.100
[AVFoundation indev @ 0x83307c000] AVFoundation video devices:
[AVFoundation indev @ 0x83307c000] [0] C922 Pro Stream Webcam
[AVFoundation indev @ 0x83307c000] [1] FaceTime HD Camera
[AVFoundation indev @ 0x83307c000] [2] Capture screen 0
[AVFoundation indev @ 0x83307c000] [3] Capture screen 1
[AVFoundation indev @ 0x83307c000] [4] Capture screen 2
[AVFoundation indev @ 0x83307c000] AVFoundation audio devices:
[AVFoundation indev @ 0x83307c000] [0] MacBook Pro Microphone
[AVFoundation indev @ 0x83307c000] [1] Microsoft Teams Audio
[AVFoundation indev @ 0x83307c000] [2] INZONE H3
[AVFoundation indev @ 0x83307c000] [3] C922 Pro Stream Webcam
[in#0 @ 0x832c38500] Error opening input: Input/output error
Error opening input file .
Error opening input files: Input/output error
`

func TestParseFFmpegDevices(t *testing.T) {
	expected := []Device{
		{Name: "MacBook Pro Microphone", Mode: 0},
		{Name: "Microsoft Teams Audio", Mode: 0},
		{Name: "INZONE H3", Mode: 0},
		{Name: "C922 Pro Stream Webcam", Mode: 0},
	}
	actual := parseFFmpegDevices(ffmpegOutputSample)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("parseFFmpegDevices() = %v, want %v", actual, expected)
	}
}
