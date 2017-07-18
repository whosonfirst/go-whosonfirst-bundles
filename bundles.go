package bundles

import (
	"github.com/whosonfirst/go-whosonfirst-clone"
	"github.com/whosonfirst/go-whosonfirst-log"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

type BundleOptions struct {
	Source       string
	Destination  string
	Compress     bool
	Dated        bool
	SkipExisting bool
	ForceUpdates bool
	Processes    int
	Logger	     *log.WOFLogger
}

type Bundle struct {
	Options *BundleOptions
}

func DefaultBundleOptions() *BundleOptions {

	tmpdir := os.TempDir()
	processes := runtime.NumCPU() * 2
	logger := log.SimpleWOFLogger("")

	opts := BundleOptions{
		Source:       "",
		Destination:  tmpdir,
		Compress:     false,
		Dated:        false,
		SkipExisting: false,
		ForceUpdates: false,
		Processes: processes,
		Logger: logger,
	}

	return &opts
}

func NewBundle(options *BundleOptions) (*Bundle, error) {

	b := Bundle{
		Options: options,
	}

	return &b, nil
}

func (b *Bundle) BundleMetafile(source_meta string) error {

	opts := b.Options

	cl, err := clone.NewWOFClone(opts.Source, opts.Destination, opts.Processes, opts.Logger)

	if err != nil {
		return err
	}

	err = cl.CloneMetaFile(source_meta, opts.SkipExisting, opts.ForceUpdates)

	if err != nil {
		return err
	}

	// copy the metafile in to the bundle

	meta_fname := filepath.Base(source_meta)
	dest_meta := filepath.Join(opts.Destination, meta_fname)

	infile, err := os.Open(source_meta)

	if err != nil {
		return err
	}

	defer infile.Close()

	outfile, err := os.Create(dest_meta)

	if err != nil {
		return err
	}

	defer outfile.Close()

	_, err = io.Copy(outfile, infile)

	if err != nil {
		return err
	}

	// copy other files in to the bundle?

	// compress the bundle

	return nil
}
