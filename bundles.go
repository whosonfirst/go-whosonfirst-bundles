package bundles

import (
	"context"
	"errors"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type BundleOptions struct {
	Source       string
	Destination  string
	BundleName   string
	Compress     bool
	Dated        bool
	DateTime     string
	SkipExisting bool
	ForceUpdates bool
	Processes    int
	Logger       *log.WOFLogger
}

type BundleInfo struct {
	MetafilePath               string
	MetafileHashPath           string
	MetafileHash               string
	MetafileCompressedPath     string
	MetafileCompressedHashPath string
	MetafileCompressedHash     string
	BundlePath                 string
	BundleCompressedPath       string
	BundleCompressedHashPath   string
	BundleCompressedHash       string
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
		BundleName:   "",
		Compress:     false,
		Dated:        false,
		DateTime:     "",
		SkipExisting: false,
		ForceUpdates: false,
		Processes:    processes,
		Logger:       logger,
	}

	return &opts
}

func NewBundle(options *BundleOptions) (*Bundle, error) {

	b := Bundle{
		Options: options,
	}

	return &b, nil
}

func (b *Bundle) BundleName(metafile string) (string, error) {

	opts := b.Options
	bundle_name := opts.BundleName

	if bundle_name != "" {
		return bundle_name, nil
	}

	if opts.Dated {
		return b.BundleNameDated(metafile)
	}

	return b.BundleNameLatest(metafile)
}

func (b *Bundle) BundleNameRoot(metafile string) (string, error) {

	opts := b.Options

	meta_fname := filepath.Base(metafile)
	bundle_name := opts.BundleName

	meta_ext := filepath.Ext(meta_fname)
	bundle_name = strings.Replace(meta_fname, meta_ext, "", -1)
	bundle_name = strings.Replace(bundle_name, "-meta-latest", "", -1)
	bundle_name = strings.Replace(bundle_name, "-meta", "", -1)
	bundle_name = strings.Replace(bundle_name, "-latest", "", -1)

	return bundle_name, nil
}

func (b *Bundle) BundleNameLatest(metafile string) (string, error) {

	bundle_name, err := b.BundleNameRoot(metafile)

	if err != nil {
		return "", nil
	}

	bundle_name = fmt.Sprintf("%s-latest-bundle", bundle_name)
	return bundle_name, nil
}

func (b *Bundle) BundleNameDated(metafile string) (string, error) {

	bundle_name, err := b.BundleNameRoot(metafile)

	if err != nil {
		return "", nil
	}

	if b.Options.DateTime == "" {
		ts := time.Now()
		dt := ts.Format("20060102T150405") // Go... Y U so weird

		b.Options.DateTime = dt
	}

	bundle_name = fmt.Sprintf("%s-%s-bundle", bundle_name, b.Options.DateTime)
	return bundle_name, nil
}

func (b *Bundle) BundleRoot(metafile string) (string, error) {

	opts := b.Options

	bundle_name, err := b.BundleName(metafile)

	if err != nil {
		return "", nil
	}

	bundle_root := filepath.Join(opts.Destination, bundle_name)

	return filepath.Abs(bundle_root)
}

func (b *Bundle) BundleMetafile(metafile string) (string, error) {

	// opts := b.Options

	bundle_root, err := b.BundleRoot(metafile)

	if err != nil {
		return "", err
	}

	info, err := os.Stat(bundle_root)

	if !os.IsNotExist(err) {

		if info.IsDir() {
			return "", errors.New("bundle already exists, please move it before proceeding")
		}
	}

	// bundle_fname := filepath.Base(bundle_root)
	// bundle_data := filepath.Join(bundle_root, "data")

	cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		/*
			path, err := index.PathForContext(ctx)

			if err != nil {
				return err
			}
		*/

		f, err := feature.LoadFeatureFromReader(fh)

		if err != nil {
			return err
		}

		id := whosonfirst.Id(f)

		_, err = uri.Id2RelPath(id)

		if err != nil {
			return err
		}

		return nil
	}

	idx, err := index.NewIndexer("meta", cb)

	if err != nil {
		return bundle_root, err
	}
	
	err = idx.IndexPath(metafile)

	if err != nil {
		return bundle_root, err
	}

	return bundle_root, nil
}
