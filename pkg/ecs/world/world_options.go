package world

type WorldOptions struct {
	PartitionResolutions []int
}

type WorldOption func(*WorldOptions)

func DefaultWorldOptions() *WorldOptions {
	return &WorldOptions{
		PartitionResolutions: []int{32, 128, 512},
	}
}

func WithPartitionResolutions(resolutions ...int) WorldOption {
	return func(opts *WorldOptions) {
		opts.PartitionResolutions = resolutions
	}
}
