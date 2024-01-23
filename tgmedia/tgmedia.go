package tgmedia

import (
	"github.com/heilkit/tg"
	"github.com/heilkit/tg/image"
	"github.com/heilkit/tg/video"
	"strings"
)

func isOneOf(item string, what []string) bool {
	for _, type_ := range what {
		if strings.HasSuffix(item, type_) {
			return true
		}
	}
	return false
}

// FromDisk is not the best way to upload media to Telegram servers, but it has OK defaults.
func FromDisk(filename string, opts ...interface{}) tg.Media {
	imageOpt := &image.Opt{}
	imageMods := []tg.ImageModifier{}
	videoOpt := &video.Opt{}
	videoMods := []tg.VideoModifier{}
	for _, opt := range opts {
		switch val := opt.(type) {
		case *image.Opt:
			imageOpt = val
		case *video.Opt:
			videoOpt = val

		case tg.ImageModifier:
			imageMods = append(imageMods, val)
		case tg.VideoModifier:
			videoMods = append(videoMods, val)
		}
	}

	lower := strings.ToLower(filename)

	imageType := []string{".jpg", ".jpeg", ".png", ".webp", ".heic", ".jxl"}
	if isOneOf(lower, imageType) {
		return tg.Photo{File: tg.FromDisk(filename)}.With(image.ConvertIfNeeded(imageOpt)).With(imageMods...)
	}

	nativeVideo := []string{".mp4", ".mov", ".mpeg4"}
	if isOneOf(lower, nativeVideo) {
		return tg.Video{File: tg.FromDisk(filename)}.With(videoMods...)
	}
	nonNativeCopyConvertableVideo := []string{".webm", ".m4v"}
	if isOneOf(lower, nonNativeCopyConvertableVideo) {
		return tg.Video{File: tg.FromDisk(filename)}.With(video.ConvertByCopy(videoOpt)).With(videoMods...)
	}
	nonNativeVideo := []string{".avi", ".gif", ".wmv", ".amv", ".qt"}
	if isOneOf(lower, nonNativeVideo) {
		return tg.Video{File: tg.FromDisk(filename)}.With(video.Convert(videoOpt)).With(videoMods...)
	}

	return &tg.Document{File: tg.FromDisk(filename)}
}
