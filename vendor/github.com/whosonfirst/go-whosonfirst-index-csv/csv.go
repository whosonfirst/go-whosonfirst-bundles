package csv

import (
	"context"
	"errors"
	wof_csv "github.com/whosonfirst/go-whosonfirst-csv"
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-index/fs"
	"io"
	"os"
	"path/filepath"
)

func init() {
	dr := NewMetaFileDriver()
	index.Register("metafile", dr)
}

type MetaFileDriver struct {
	index.Driver
	data_root   string
	file_driver index.Driver
}

func NewMetaFileDriver() index.Driver {

	fd := fs.NewFileDriver()

	data_root := "FIXME"

	return &MetaFileDriver{
		data_root:   data_root,
		file_driver: fd,
	}
}

func (d *MetaFileDriver) Open(uri string) error {
	return d.file_driver.Open(uri)
}

func (d *MetaFileDriver) IndexURI(ctx context.Context, index_cb index.IndexerFunc, uri string) error {

	abs_path, err := filepath.Abs(uri)

	if err != nil {
		return err
	}

	fh, err := os.Open(abs_path)

	if err != nil {
		return err
	}

	defer fh.Close()

	reader, err := wof_csv.NewDictReader(fh)

	if err != nil {
		return err
	}

	for {
		row, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		rel_path, ok := row["path"]

		if !ok {
			return errors.New("Missing path key")
		}

		// TO DO: make this work with a row["repo"] key
		// (20170809/thisisaaronland)

		file_path := filepath.Join(d.data_root, rel_path)
		ctx = index.AssignPathContext(ctx, file_path)

		return d.file_driver.IndexURI(ctx, index_cb, file_path)
	}

	return nil
}

/*
	// please refactor all of this in to something... better
	// (20170823/thisisaaronland)

	parts := strings.Split(path, ":")

	if len(parts) == 1 {

		abs_root, err := filepath.Abs(parts[0])

		if err != nil {
			return err
		}

		meta_root := filepath.Dir(abs_root)
		repo_root := filepath.Dir(meta_root)
		data_root := filepath.Join(repo_root, "data")

		parts = append(parts, data_root)
	}

	if len(parts) != 2 {
		return errors.New("Invalid path declaration for a meta file")
	}

	for _, p := range parts {

		if p == STDIN {
			continue
		}

		_, err := os.Stat(p)

		if os.IsNotExist(err) {
			return errors.New("Path does not exist")
		}
	}

	meta_file := parts[0]
	data_root := parts[1]

	return i.IndexMetaFile(meta_file, data_root, args...)

*/
