package bundles

import (
	"context"
	"errors"
	"github.com/facebookgo/atomicfile"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	"os"
	"path/filepath"
)

type BundleOptions struct {
	Mode        string
	Destination string
	Logger      *log.WOFLogger
}

type Bundle struct {
	Options *BundleOptions
}

func DefaultBundleOptions() *BundleOptions {

	tmpdir := os.TempDir()
	logger := log.SimpleWOFLogger("")

	opts := BundleOptions{
		Mode:        "repo",
		Destination: tmpdir,
		Logger:      logger,
	}

	return &opts
}

func NewBundle(options *BundleOptions) (*Bundle, error) {

	b := Bundle{
		Options: options,
	}

	return &b, nil
}

func (b *Bundle) BundleMetafile(metafile string) error {

	abs_metafile, err := filepath.Abs(metafile)

	if err != nil {
		return err
	}

	b.Options.Mode = "meta"
	err = b.Bundle(abs_metafile)

	if err != nil {
		return nil
	}

	fname := filepath.Base(abs_metafile)
	cp_metafile := filepath.Join(b.Options.Destination, fname)

	in, err := os.Open(abs_metafile)

	if err != nil {
		return err
	}

	err = b.cloneFH(in, cp_metafile)

	if err != nil {
		return err
	}

	return nil
}

func (b *Bundle) Bundle(to_index ...string) error {

	opts := b.Options
	root := opts.Destination
	mode := opts.Mode

	data_root := filepath.Join(root, "data")

	info, err := os.Stat(data_root)

	if err != nil {

		if !os.IsNotExist(err) {
			return err
		}

		// MkdirAll ? (20180620/thisisaaronland)
		err = os.Mkdir(data_root, 0755)

		if err != nil {
			return err
		}

		root = data_root

	} else {

		if !info.IsDir() {
			return errors.New("...")
		}
	}

	f := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		path, err := index.PathForContext(ctx)

		if err != nil {
			return err
		}

		is_wof, err := uri.IsWOFFile(path)

		if err != nil {
			return err
		}

		if !is_wof {
			return nil
		}

		f, err := feature.LoadFeatureFromReader(fh)

		if err != nil {
			return err
		}

		id := whosonfirst.Id(f)

		abs_path, err := uri.Id2AbsPath(root, id)

		if err != nil {
			return nil
		}

		abs_root := filepath.Dir(abs_path)

		_, err = os.Stat(abs_root)

		if os.IsNotExist(err) {

			// SOMETHING SOMETHING SOMETHING LOCK/UNLOCK MUTEX HERE

			err = os.MkdirAll(abs_root, 0755)

			if err != nil {
				return err
			}
		}

		err = b.cloneFH(fh, abs_path)

		if err != nil {
			return err
		}

		// SOMETHING SOMETHING SOMETHING WRITE f TO METAFILE HERE...

		return nil
	}

	idx, err := index.NewIndexer(mode, f)

	if err != nil {
		return err
	}

	for _, path := range to_index {

		err := idx.IndexPath(path)

		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Bundle) cloneFH(in io.Reader, out_path string) error {

	b.Options.Logger.Debug("Clone file to %s", out_path)

	out, err := atomicfile.New(out_path, 0644)

	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)

	if err != nil {
		out.Abort()
		return err
	}

	return out.Close()
}
