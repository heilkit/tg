package tgmedia

import (
	"github.com/heilkit/tg"
	"github.com/heilkit/tg/tgimage"
	"github.com/heilkit/tg/tgvideo"
	"path/filepath"
	"strings"
)

// FromDisk tries to guess telegram file type based of filename.
// The function has okay defaults, but I encourage you to handle everything yourself if you want to be sure.
func FromDisk(filename string, opts ...interface{}) tg.Inputtable {
	convert := true
	imageMods := []tg.ImageModifier{}
	videoMods := []tg.VideoModifier{}
	for _, opt := range opts {
		switch val := opt.(type) {
		case tg.ImageModifier:
			imageMods = append(imageMods, val)
		case tg.VideoModifier:
			videoMods = append(videoMods, val)
		case bool:
			convert = val
		}
	}
	if convert {
		if len(imageMods) == 0 {
			imageMods = append(imageMods, tgimage.ConvertIfTooBig())
		}
		if len(videoMods) == 0 {
			videoMods = append(videoMods, tgvideo.ThumbnailAt(0.05))
		}
	}
	nativeImage, nativeVideo := []string{".jpg", ".jpeg", ".png"}, []string{".mp4", ".mov", ".mpeg4"}
	convertImage, convertVideo, convertByCopyVideo := []string{}, []string{}, []string{}
	if convert {
		convertImage = append(convertImage, ".webp", ".heic", ".jxl")
		convertVideo = append(convertVideo, ".avi", ".gif", ".wmv", ".amv", ".qt")
		convertByCopyVideo = append(convertByCopyVideo, ".webm", ".m4v")
	}

	return FromDiskVerbose(filename, imageMods, videoMods, nativeImage, convertImage, nativeVideo, convertVideo, convertByCopyVideo)
}

// FromDiskVerbose uploads media from file.
//
// modsImage,    modsVideo    -- are common for native and non-native media,
// nativeImage,  nativeVideo  -- filetypes that do not require conversion before uploading,
// convertImage, convertVideo -- filetypes that require conversion before uploading,
// nativeImage,  nativeVideo  -- filetypes that do not require conversion before uploading.
//
// Note: filetypes has to be lowercase, the FromDiskVerbose is case-insensitive by design.
func FromDiskVerbose(filename string, modsImage []tg.ImageModifier, modsVideo []tg.VideoModifier,
	nativeImage, convertImage,
	nativeVideo, convertVideo, convertByCopyVideo []string) tg.Inputtable {
	lower := strings.ToLower(filename)

	switch {
	case isOneOf(lower, nativeImage):
		return tg.Photo{File: tg.FromDisk(filename)}.With(modsImage...)
	case isOneOf(lower, convertImage):
		return tg.Photo{File: tg.FromDisk(filename)}.With(modsImage...).With(tgimage.Convert())

	case isOneOf(lower, nativeVideo):
		return tg.Video{File: tg.FromDisk(filename), FileName: filepath.Base(filename)}.With(modsVideo...)
	case isOneOf(lower, convertVideo):
		return tg.Video{File: tg.FromDisk(filename), FileName: filepath.Base(filename)}.With(modsVideo...).With(tgvideo.Convert())
	case isOneOf(lower, convertByCopyVideo):
		return tg.Video{File: tg.FromDisk(filename), FileName: filepath.Base(filename)}.With(modsVideo...).With(tgvideo.ConvertByCopy())
	}

	return &tg.Document{File: tg.FromDisk(filename), FileName: filepath.Base(filename)}
}

func isOneOf(item string, what []string) bool {
	for _, type_ := range what {
		if strings.HasSuffix(item, type_) {
			return true
		}
	}
	return false
}
