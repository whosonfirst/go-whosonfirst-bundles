package options

import (
	"github.com/whosonfirst/go-whosonfirst-log"
	"io/ioutil"
)

type BuildOptions struct {
	Timings        bool
	Placetypes     []string
	Roles          []string
	Exclude        []string
	Workdir        string
	MaxFilehandles int
	Logger         *log.WOFLogger
}

func DefaultBuildOptions() (*BuildOptions, error) {

	workdir, err := ioutil.TempDir("", "meta")

	if err != nil {
		return nil, err
	}

	logger := log.SimpleWOFLogger()

	opts := BuildOptions{
		Timings:        false,
		Placetypes:     []string{},
		Roles:          []string{},
		Exclude:        []string{},
		Workdir:        workdir,
		MaxFilehandles: 1024,
		Logger:         logger,
	}

	return &opts, nil
}
