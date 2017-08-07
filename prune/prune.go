package prune

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	Logger     *log.Logger // maybe use go-whosonfirst-log... I don't know yet
}

func NewDefaultPruneOptions() (*PruneOptions, error) {

	ex, err := os.Executable()

	if err != nil {
		return nil, err
	}

	app := filepath.Base(ex)
	prefix := fmt.Sprintf("[%s] ", app)

	logger := log.New(os.Stdout, prefix, log.Lshortfile)

	opts := PruneOptions{
		Debug:      false,
		Dated:      true,
		MaxBundles: 10,
		Logger:     logger,
	}

	return &opts, nil
}

func FilesToCandidates(files []File, opts *PruneOptions) (Candidates, error) {

	re_bundle := regexp.MustCompile(`^(.*)\-(?:\d{8}T\d{6})\-bundle\.tar\.bz2?$`)

	candidates := make(map[string][]File)

	for _, file := range files {

		fname := file.Name()
		fname = filepath.Base(fname)

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
