package tgvideo

const (
	ffmpeg  = "ffmpeg"
	ffprobe = "ffprobe"
	convert = "convert"
	preset  = "fast"
)

// Opt for modifiers. Not all of them are used every time.
type Opt struct {
	Width   int
	Height  int
	Preset  string
	Ffmpeg  string
	Ffprobe string
	Convert string
	TempDir string
}

func (opts *Opt) Defaults() *Opt {
	if opts == nil {
		opts = &Opt{}
	}
	if opts.Convert == "" {
		opts.Convert = convert
	}
	if opts.Ffmpeg == "" {
		opts.Ffmpeg = ffmpeg
	}
	if opts.Ffprobe == "" {
		opts.Ffprobe = ffprobe
	}
	if opts.Width <= 0 {
		opts.Width = 5000
	}
	if opts.Height <= 0 {
		opts.Height = 5000
	}
	if opts.Preset == "" {
		opts.Preset = preset
	}
	return opts
}
