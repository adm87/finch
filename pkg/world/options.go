package world

type Options struct {
	PartitionResolutions []int
}

type Option func(*Options)

func DefaultOptions() *Options {
	return &Options{
		PartitionResolutions: []int{32, 128, 512},
	}
}

func WithPartitionResolutions(resolutions ...int) Option {
	return func(opts *Options) {
		opts.PartitionResolutions = resolutions
	}
}
