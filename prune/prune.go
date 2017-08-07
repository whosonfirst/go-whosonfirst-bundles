package prune

import (
	"regexp"
)

type Pruner interface {
	ListFiles() ([]File, error)
	PruneFiles([]File) error
}

type File interface {
	Name() string
}

type Candidates map[string][]File

type PruneOptions struct {
	Debug      bool
	Dated      bool
	MaxBundles int
}

func NewDefaultPruneOptions() (*PruneOptions, error) {

	opts := PruneOptions{
		Debug:      false,
		Dated:      true,
		MaxBundles: 10,
	}

	return &opts, nil
}

func FilesToCandidates(files []File, opts *PruneOptions) (Candidates, error) {

	re_bundle := regexp.MustCompile(`^(.*)\-(?:\d{8}T\d{6})\-bundle\.tar\.bz2?$`)

	candidates := make(map[string][]File)

	for _, file := range files {

		fname := file.Name()

		if !re_bundle.MatchString(fname) {
			continue
		}

		m := re_bundle.FindAllStringSubmatch(fname, -1)

		if m == nil {
			continue
		}

		short_name := m[0][1]

		_, ok := candidates[short_name]

		if !ok {
			candidates[short_name] = make([]File, 0)
		}

		candidates[short_name] = append(candidates[short_name], file)
	}

	return candidates, nil
}
