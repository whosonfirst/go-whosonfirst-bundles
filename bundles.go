package bundles

import (
	"context"
	"errors"
	"fmt"
	"github.com/facebookgo/atomicfile"
	"github.com/whosonfirst/go-whosonfirst-csv"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-meta"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type BundleOptions struct {
	Mode        string
	Destination string
	Metafile    bool
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
		Metafile:    true,
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
	b.Options.Metafile = false

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

	var meta_writer *csv.DictWriter
	var meta_fh *atomicfile.File

	mu := new(sync.Mutex)

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

		// PLEASE MAKE THIS BETTER (20180620/thisisaaronland)

		id := whosonfirst.Id(f)

		abs_path, err := uri.Id2AbsPath(root, id)

		if err != nil {
			return nil
		}

		abs_root := filepath.Dir(abs_path)

		_, err = os.Stat(abs_root)

		if os.IsNotExist(err) {

			mu.Lock()

			err = os.MkdirAll(abs_root, 0755)

			mu.Unlock()

			if err != nil {
				return err
			}
		}

		err = b.cloneFH(fh, abs_path)

		if err != nil {
			return err
		}

		if opts.Metafile {

			row, err := meta.FeatureToRow(f.Bytes())

			if err != nil {
				return err
			}

			if meta_writer == nil {

				// Basically all of this stuff around fieldnames and headers
				// should be moved in to go-whosonfirst-csv itself...
				// (20180620/thisisaaronland)

				fieldnames := make([]string, 0)

				for k, _ := range row {
					fieldnames = append(fieldnames, k)
				}

				sort.Strings(fieldnames)

				dest := b.Options.Destination
				dest_fname := filepath.Base(dest)

				meta_fname := fmt.Sprintf("%s-latest.csv", dest_fname)
				meta_path := filepath.Join(dest, meta_fname)

				fh, err := atomicfile.New(meta_path, 0644)

				if err != nil {
					return err
				}

				meta_fh = fh

				wr, err := csv.NewDictWriter(meta_fh, fieldnames)

				if err != nil {
					return err
				}

				meta_writer = wr
				meta_writer.WriteHeader()
			}

			meta_writer.WriteRow(row)
		}

		return nil
	}

	idx, err := index.NewIndexer(mode, f)

	if err != nil {
		return err
	}

	for _, path := range to_index {

		err := idx.IndexPath(path)

		if err != nil {

			if opts.Metafile && meta_fh != nil {
				meta_fh.Abort()
			}

			return err
		}
	}

	if opts.Metafile && meta_fh != nil {
		meta_fh.Close()
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
