package compress

type CompressOptions struct {
	RemoveSource bool
}

func DefaultCompressOptions() *CompressOptions {

	opts := CompressOptions{
		RemoveSource: false,
	}

	return &opts
}
