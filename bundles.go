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
	_ "github.com/whosonfirst/go-whosonfirst-index/fs"
	_ "github.com/whosonfirst/go-whosonfirst-index-csv"
	_ "github.com/whosonfirst/go-whosonfirst-index-sqlite"		
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-meta"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type BundleOptions struct {
	Mode           string
	Destination    string
	Metafile       bool
	Logger         *log.WOFLogger
	MaxFileHandles int
}

type Bundle struct {
	Options  *BundleOptions
	mu       *sync.Mutex
	throttle chan bool
}

func DefaultBundleOptions() *BundleOptions {

	tmpdir := os.TempDir()
	logger := log.SimpleWOFLogger("")

	opts := BundleOptions{
		Mode:           "repo",
		Destination:    tmpdir,
		Metafile:       true,
		Logger:         logger,
		MaxFileHandles: 100,
	}

	return &opts
}

func NewBundle(options *BundleOptions) (*Bundle, error) {

	max_fh := options.MaxFileHandles
	throttle_ch := make(chan bool, max_fh)

	for i := 0; i < max_fh; i++ {
		throttle_ch <- true
	}

	mu := new(sync.Mutex)

	b := Bundle{
		Options:  options,
		mu:       mu,
		throttle: throttle_ch,
	}

	return &b, nil
}

func (b *Bundle) BundleMetafilesFromSQLite(ctx context.Context, dsn string, metafiles ...string) error {

	err_ch := make(chan error)
	done_ch := make(chan bool)

	for _, path := range metafiles {

		go func(b *Bundle, ctx context.Context, dsn string, path string) {

			defer func() {
				done_ch <- true
			}()

			select {

			case <-ctx.Done():
				return
			default:

				err := b.BundleMetafileFromSQLite(ctx, dsn, path)

				if err != nil {
					err_ch <- err
				}
			}

		}(b, ctx, dsn, path)
	}

	remaining := len(metafiles)

	for remaining > 0 {

		select {
		case <-done_ch:
			remaining -= 1
		case e := <-err_ch:
			return e
		default:
			// pass
		}
	}

	return nil
}

func (b *Bundle) BundleMetafileFromSQLite(ctx context.Context, dsn string, metafile string) error {

	db, err := database.NewDB(dsn)

	if err != nil {
		return err
	}

	defer db.Close()

	// is it worth wrapping all of this in a select / context block ?
	// today it doesn't seem like it... (20180622/thisisaaronland)

	abs_metafile, err := filepath.Abs(metafile)

	if err != nil {
		return err
	}

	reader, err := csv.NewDictReaderFromPath(abs_metafile)

	if err != nil {
		return nil
	}

	conn, err := db.Conn()

	if err != nil {
		return err
	}

	defer db.Close()

	tbl, err := tables.NewGeoJSONTable()

	if err != nil {
		return err
	}

	bundle_path := b.Options.Destination
	data_path := filepath.Join(bundle_path, "data")

	// this is necessary so we can break out of the select block which is
	// wrapped in a for block... good times (20180622/thisisaaronland)

	eof := false

	for {

		select {

		case <-ctx.Done():
			return nil
		default:

			csv_row, err := reader.Read()

			if err == io.EOF {
				eof = true
				break
			}

			if err != nil {
				return err
			}

			str_id, ok := csv_row["id"]

			if !ok {
				return errors.New("Missing ID")
			}

			// we could wait until after the DB query to do this but if
			// it's going to fail maybe we want to know sooner...
			// (20180622/thisisaaronland)

			id, err := strconv.ParseInt(str_id, 10, 64)

			if err != nil {
				return err
			}

			sql := fmt.Sprintf("SELECT body FROM %s WHERE id= ?", tbl.Name())

			db_row := conn.QueryRow(sql, id)

			var body string
			err = db_row.Scan(&body)

			if err != nil {
				return err
			}

			fh := strings.NewReader(body)

			abs_path, err := b.ensurePathForID(data_path, id)

			if err != nil {
				return nil
			}

			err = b.cloneFH(fh, abs_path)

			if err != nil {
				return err
			}
		}

		if eof {
			break
		}
	}

	fname := filepath.Base(abs_metafile)
	cp_metafile := filepath.Join(bundle_path, fname)

	in, err := os.Open(abs_metafile)

	if err != nil {
		return err
	}

	defer in.Close()

	err = b.cloneFH(in, cp_metafile)

	if err != nil {
		return err
	}

	return nil
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

	defer in.Close()

	err = b.cloneFH(in, cp_metafile)

	if err != nil {
		return err
	}

	return nil
}

func (b *Bundle) Bundle(to_index ...string) error {

	opts := b.Options
	mode := opts.Mode

	bundle_path := b.Options.Destination
	data_path := filepath.Join(bundle_path, "data")

	var meta_writer *csv.DictWriter
	var meta_fh *atomicfile.File

	f := func(ctx context.Context, fh io.Reader, args ...interface{}) error {

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

		abs_path, err := b.ensurePathForID(data_path, id)

		if err != nil {
			return nil
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

				dest_fname := filepath.Base(bundle_path)

				meta_fname := fmt.Sprintf("%s-latest.csv", dest_fname)
				meta_path := filepath.Join(bundle_path, meta_fname)

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

func (b *Bundle) ensurePathForID(root string, id int64) (string, error) {

	abs_path, err := uri.Id2AbsPath(root, id)

	if err != nil {
		return "", err
	}

	abs_root := filepath.Dir(abs_path)

	_, err = os.Stat(abs_root)

	if os.IsNotExist(err) {

		b.mu.Lock()

		err = os.MkdirAll(abs_root, 0755)

		b.mu.Unlock()

		if err != nil {
			return "", err
		}
	}

	return abs_path, nil
}

func (b *Bundle) cloneFH(in io.Reader, out_path string) error {

	<-b.throttle

	defer func() {
		b.throttle <- true
	}()

	b.Options.Logger.Debug("Clone file to %s", out_path)

	out, err := atomicfile.New(out_path, 0644)

	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)

	if err != nil {

		abort_err := out.Abort()

		if abort_err != nil {
			b.Options.Logger.Warning("Failed to remove atomicwrites file for %s, because %s", out_path, abort_err)
		}

		return err
	}

	return out.Close()
}
