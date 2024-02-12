package tg

import (
	"io"
	"os"
	"strings"
)

// Local submodule handles weird tg style of handling local servers.
type Local interface {
	Download(b *Bot, file *File, dst string) error
}

var _ Local = localCopying{}

var _ Local = localMoving{}
var _ Local = localMovingCrossDevice{}

// localCopying copies the file from local telegram-bot-api data directory to dst,
// providing the path to original copy to file.FileLocal.
type localCopying struct{}

func LocalCopying() Local {
	return &localCopying{}
}

func (loc localCopying) Download(b *Bot, file *File, dst string) error {
	localPath := file.FilePath
	if file.FilePath == "" {
		f, err := b.FileByID(file.FileID)
		if err != nil {
			return err
		}
		// FilePath is updated, allowing user to delete the file from the local server's cache
		localPath = f.FilePath
		file.FilePath = localPath
	}

	reader, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	out, err := os.Create(dst)
	if err != nil {
		return wrapError(err)
	}
	defer out.Close()

	if _, err := io.Copy(out, reader); err != nil {
		return wrapError(err)
	}

	file.FileLocal = localPath
	return nil
}

// LocalMoving the file from local telegram-bot-api storage directory after downloading it.
// If crossDevice=true on "invalid cross-device link" error perform a local copy and delete the original file afterward.
func LocalMoving(crossDevice ...bool) Local {
	if len(crossDevice) == 0 || crossDevice[0] {
		return &localMovingCrossDevice{}
	}
	return &localMoving{}
}

// localMoving move the file from telegram-bot-api directory, re-uploading is not promised,
// if you care about possible multiple file downloads, you should consider localCopying.
type localMoving struct{}

// localMovingCrossDevice wraps around localMoving to support cross-device file movement, i.e., into Docker containers.
type localMovingCrossDevice struct{}

func (l localMovingCrossDevice) Download(b *Bot, file *File, dst string) error {
	localPath := file.FilePath
	if file.FilePath == "" {
		f, err := b.FileByID(file.FileID)
		if err != nil {
			return err
		}
		localPath = f.FilePath
		file.FilePath = localPath
	}

	if err := move(localPath, dst); err != nil {
		return wrapError(err)
	}

	file.FileLocal = dst
	return nil
}

func (loc localMoving) Download(b *Bot, file *File, dst string) error {
	localPath := file.FilePath
	if file.FilePath == "" {
		f, err := b.FileByID(file.FileID)
		if err != nil {
			return err
		}
		localPath = f.FilePath
		file.FilePath = localPath
	}

	if err := os.Rename(localPath, dst); err != nil {
		return wrapError(err)
	}

	file.FileLocal = dst
	return nil
}

func move(source, destination string) error {
	err := os.Rename(source, destination)
	if err != nil && strings.Contains(err.Error(), "invalid cross-device link") {
		return moveCrossDevice(source, destination)
	}
	return err
}

func moveCrossDevice(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		_ = srcFile.Close()
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	stat, err := os.Stat(src)
	if err != nil {
		_ = os.Remove(dst)
		return err
	}

	err = os.Chmod(dst, stat.Mode())
	if err != nil {
		_ = os.Remove(dst)
		return err
	}
	_ = os.Remove(src)
	return nil
}
