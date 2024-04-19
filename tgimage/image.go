package tgimage

const (
	convert = "convert"
)

type Opt struct {
	Width      int
	Height     int
	HardResize bool
	Convert    string
	TempDir    string
	Quality    int
}

func parseOpts(opts ...*Opt) Opt {
	opt := Opt{
		Width:      5000,
		Height:     5000,
		HardResize: false,
		Convert:    convert,
		TempDir:    "",
		Quality:    95,
	}
	if len(opts) == 0 {
		return opt
	}
	opts_ := opts[0]

	if opts_.Width != 0 {
		opt.Width = opts_.Width
	}
	if opts_.Height != 0 {
		opt.Height = opts_.Height
	}
	if opts_.Convert != "" {
		opt.Convert = opts_.Convert
	}
	if opts_.TempDir != "" {
		opt.TempDir = opts_.TempDir
	}
	if opts_.Quality != 0 {
		opt.Quality = opts_.Quality
	}
	opt.HardResize = opts_.HardResize
	return opt
}
