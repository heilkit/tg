package tgimage

import (
	"fmt"
	"github.com/heilkit/tg"
	"os"
	"os/exec"
)

// Convert image to suit Telegram's standards.
func Convert(opts ...*Opt) tg.ImageModifier {
	opt := parseOpts(opts...)
	resizeArg := ""
	if opt.HardResize {
		resizeArg = fmt.Sprintf("%dx%d!", opt.Height, opt.Width)
	} else {
		resizeArg = fmt.Sprintf("%dx%d>", opt.Height, opt.Width)
	}
	return func(photo *tg.Photo) (temporaries []string, err error) {
		defer func() {
			if err != nil {
				err = fmt.Errorf("tgimage.Convert: %v", err)
			}
		}()

		tmp, err := os.CreateTemp(opt.TempDir, "*.jpg")
		if err != nil {
			return nil, err
		}
		_ = tmp.Close()

		ret := []string{tmp.Name()}
		output, err := exec.Command(opt.Convert, photo.File.FileLocal, "-resize", resizeArg, tmp.Name()).CombinedOutput()
		if err != nil {
			return ret, wrapExecError(err, output)
		}

		photo.FileLocal = tmp.Name()
		return ret, nil
	}
}

// ConvertIfNeeded image only if Telegram does not support its type.
func ConvertIfNeeded(opts ...*Opt) tg.ImageModifier {
	convert := Convert(opts...)
	return func(photo *tg.Photo) (temporaries []string, err error) {
		filename := photo.FileLocal
		if !isTypeSupported(filename) {
			return convert(photo)
		}
		if stat, err := os.Stat(filename); err != nil || stat.Size() > int64(10*(1<<20)) {
			return convert(photo)
		}
		return nil, nil
	}
}

// ConvertIfTooBig tgimage only if it's too big to be a Telegram photo (10MB).
func ConvertIfTooBig(opts ...*Opt) tg.ImageModifier {
	convert := Convert(opts...)
	return func(photo *tg.Photo) (temporaries []string, err error) {
		filename := photo.FileLocal
		if stat, err := os.Stat(filename); err != nil || stat.Size() > int64(10*(1<<20)) {
			return convert(photo)
		}
		return nil, nil
	}
}
