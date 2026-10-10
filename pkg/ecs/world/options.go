package world

type Options struct {
	Resolutions []int
}

type Option func(*Options)

func DefaultOptions() *Options {
	return &Options{
		Resolutions: []int{16, 32, 64},
	}
}

func WithResolutions(resolutions ...int) Option {
	return func(o *Options) {
		o.Resolutions = resolutions
	}
}
