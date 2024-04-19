package tgvideo

import (
	"fmt"
	"github.com/heilkit/tg"
	"os"
	"os/exec"
	"strings"
)

// Convert is a general purpose VideoModifier, converts a video to h264, could decrease its dimensions.
// REQUIRES `ffmpeg` on the system, which could be passed via Opt.Convert.
func Convert(opts ...*Opt) tg.VideoModifier {
	options := parseOpts(opts...)

	scaleRule := makeScaleRule(options.Width, options.Height)
	return func(video *tg.Video) (temporaries []string, err error) {
		if video == nil || video.FileLocal == "" {
			return nil, nil
		}

		tmpFile, err := os.CreateTemp(options.TempDir, fmt.Sprintf("*_heilkit_tg_%s", filetype(video.FileLocal)))
		if err != nil {
			return nil, err
		}

		output, err := exec.Command(options.Ffmpeg, "-y",
			"-i", video.FileLocal,
			"-vf", scaleRule,
			"-vcodec", "libx264",
			"-acodec", "aac",
			"-preset", options.Preset,
			tmpFile.Name()).
			CombinedOutput()
		if err != nil {
			return []string{tmpFile.Name()}, wrapExecError(err, output)
		}

		video.FileLocal = tmpFile.Name()
		return []string{tmpFile.Name()}, nil
	}
}

// ConvertIfNeeded ensures a video is converted to a type supported by Telegram.
// REQUIRES `ffmpeg` on the system, which could be passed via Opt.Convert.
func ConvertIfNeeded(opts ...*Opt) tg.VideoModifier {
	convert := Convert(opts...)
	convertByCopy := ConvertByCopy(opts...)
	return func(video *tg.Video) (temporaries []string, err error) {
		filename := video.FileLocal
		switch {
		case isVideoTypeSupported(filename):
			return nil, nil

		case isVideoTypeConvertableByCopy(filename):
			return convertByCopy(video)

		default:
			return convert(video)
		}
	}
}

// ConvertByCopy allows to upload .webm, .m4v and other video formats without re-encoding.
// REQUIRES `ffmpeg` on the system.
func ConvertByCopy(opts ...*Opt) tg.VideoModifier {
	options := parseOpts(opts...)

	return func(video *tg.Video) (temporaries []string, err error) {
		tmpFile, err := os.CreateTemp(options.TempDir, "*_heilkit_tg.mp4")
		if err != nil {
			return nil, err
		}

		output, err := exec.Command(options.Ffmpeg, "-y",
			"-i", video.FileLocal,
			tmpFile.Name(),
			"-c", "copy",
		).CombinedOutput()
		if err != nil {
			return []string{tmpFile.Name()}, wrapExecError(err, output)
		}

		video.FileLocal = tmpFile.Name()
		if index := strings.LastIndex(video.FileName, "."); index != -1 {
			video.FileName = video.FileName[0:index] + ".mp4"
		}
		return []string{tmpFile.Name()}, nil
	}
}

// EnsureMeta ensures, Telegram would process a file correctly.
// REQUIRES `ffprobe` on the system, which could be passed via Opt.Convert
func EnsureMeta(opts ...*Opt) tg.VideoModifier {
	options := parseOpts(opts...)

	return func(video *tg.Video) (temporaries []string, err error) {
		_, _, err = getSetMetadata(video, options)
		return nil, err
	}
}

// EmbedMetadata into a file before sending.
// REQUIRES `ffmpeg` on the system, which could be passed via Opt.Convert
func EmbedMetadata(meta map[string]string, opts ...*Opt) tg.VideoModifier {
	options := parseOpts(opts...)

	return func(video *tg.Video) (temporaries []string, err error) {
		tmpFile, err := os.CreateTemp(options.TempDir, "*_heilkit_tg.mp4")
		if err != nil {
			return nil, err
		}
		defer tmpFile.Close()

		args := []string{"-y", "-i", video.FileLocal, "-vcodec", "copy", "-acodec", "copy"}
		for k, v := range meta {
			args = append(args, "-metadata", fmt.Sprintf("%s='%s'", k, v))
		}
		args = append(args, tmpFile.Name())

		output, err := exec.Command(options.Ffmpeg, args...).
			CombinedOutput()
		if err != nil {
			return []string{tmpFile.Name()}, wrapExecError(err, output)
		}

		video.FileLocal = tmpFile.Name()
		return []string{tmpFile.Name()}, nil
	}
}

// ThumbnailFrom converts a picture to a 320x320 frame, suitable from Telegram video thumbnail.
// REQUIRES `convert` on the system, could be passed via Opt.Convert.
func ThumbnailFrom(filename string, opts ...*Opt) tg.VideoModifier {
	options := parseOpts(opts...)

	return func(video *tg.Video) (temporaries []string, err error) {
		extraFile, err := formatPreview(options.TempDir, options.Convert, filename)
		if err != nil {
			return []string{extraFile}, err
		}
		video.Thumbnail = &tg.Photo{File: tg.FromDisk(extraFile)}
		_, _, err = getSetMetadata(video, options)
		return []string{extraFile}, err
	}
}

// ThumbnailAt creates a thumbnail from the video frame. Position could be chosen as:
//  1. float64 -- from [0, 1], relative position in Video
//  2. string  -- position in ffmpeg format, i.e. "00:05:12.99"
//
// REQUIRES `ffmpeg`, `ffprobe` on the system, could be passed via Opt.
func ThumbnailAt(position interface{}, opts ...*Opt) tg.VideoModifier {
	switch position.(type) {
	case float64:
	case string:
	default:
		panic("ThumbnailAt: position type is not supported")
	}

	options := parseOpts(opts...)

	return func(video *tg.Video) (filename []string, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("ThumbnailAt panicked with %v", err)
			}
		}()
		if video == nil || video.FileLocal == "" {
			return nil, nil
		}

		_, videoDuration, err := getSetMetadata(video, options)

		thumbnail, err := makeThumbnailAtAlt(options.TempDir, options.Ffmpeg, video.FileLocal, calcThumbnailPosition(videoDuration, position))
		if err != nil {
			return []string{thumbnail}, err
		}

		video.Thumbnail = &tg.Photo{File: tg.FromDisk(thumbnail)}
		return []string{thumbnail}, err
	}
}

// Mute a video by creating a local muted copy.
// REQUIRES `ffmpeg` on the system. Could be passed via Opt.Convert.
func Mute(opts ...*Opt) tg.VideoModifier {
	options := parseOpts(opts...)

	return func(video *tg.Video) (temporaries []string, err error) {
		if video == nil || video.FileLocal == "" {
			return nil, nil
		}

		tmpFile, err := os.CreateTemp(options.TempDir, fmt.Sprintf("*_heilkit_tg_%s", filetype(video.FileLocal)))
		if err != nil {
			return nil, err
		}
		defer tmpFile.Close()

		output, err := exec.Command(options.Ffmpeg, "-y", "-i", video.FileLocal, "-vcodec", "copy", "-an", tmpFile.Name()).
			CombinedOutput()
		if err != nil {
			return nil, wrapExecError(err, output)
		}

		video.FileLocal = tmpFile.Name()
		return []string{tmpFile.Name()}, nil
	}
}

// OnError allows set action on error for the wrapped tg.VideoModifier. If wrapper returns nil, the error is muted.
func OnError(mod tg.VideoModifier, fn func(temporaries []string, err error) error) tg.VideoModifier {
	return func(video *tg.Video) (temporaries []string, err error) {
		ret_, err_ := mod(video)
		if err_ != nil {
			err = fn(ret_, err_)
		}
		return ret_, err
	}
}
