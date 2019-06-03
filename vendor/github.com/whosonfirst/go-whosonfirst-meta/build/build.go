package build

import (
	"context"
	"errors"
	"fmt"
	"github.com/facebookgo/atomicfile"
	"github.com/whosonfirst/go-whosonfirst-csv"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-index/utils"
	"github.com/whosonfirst/go-whosonfirst-meta"
	"github.com/whosonfirst/go-whosonfirst-meta/options"
	"github.com/whosonfirst/go-whosonfirst-placetypes/filter"
	"github.com/whosonfirst/go-whosonfirst-repo"
	"github.com/whosonfirst/warning"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func BuildFromIndex(opts *options.BuildOptions, mode string, indices []string) ([]string, error) {

	placetype_filter, err := filter.NewPlacetypesFilter(opts.Placetypes, opts.Roles, opts.Exclude)

	if err != nil {
		return nil, err
	}

	mu := new(sync.Mutex)

	throttle := make(chan bool, opts.MaxFilehandles)

	for i := 0; i < opts.MaxFilehandles; i++ {
		throttle <- true
	}

	var count int32
	var open int32
	var pending int32
	var scheduled int32

	count = 0
	open = 0
	pending = 0
	scheduled = 0

	filehandles := make(map[string]*atomicfile.File)
	writers := make(map[string]*csv.DictWriter)
	paths := make(map[string]string)

	wg := new(sync.WaitGroup)

	cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		atomic.AddInt32(&pending, 1)
		// log.Printf("pending %d scheduled %d\n", pending, scheduled)

		<-throttle

		atomic.AddInt32(&pending, -1)
		atomic.AddInt32(&scheduled, 1)

		wg.Add(1)

		defer func() {
			throttle <- true
			wg.Done()
		}()

		path, err := wof_index.PathForContext(ctx)

		if err != nil {
			return err
		}

		// TBD
		// PLEASE MAKE THIS SUPPORT ALT FILES, YEAH
		// (20190601/thisisaaronland)

		ok, err := utils.IsPrincipalWOFRecord(fh, ctx)

		if err != nil {
			return err
		}

		if !ok {
			return nil
		}

		f, err := feature.LoadFeatureFromReader(fh)

		if err != nil && !warning.IsWarning(err) {
			return err
		}

		atomic.AddInt32(&open, 1)
		defer atomic.AddInt32(&open, -1)

		placetype := f.Placetype()

		allow, err := placetype_filter.AllowFromString(placetype)

		if err != nil && !warning.IsWarning(err) {
			log.Println(fmt.Sprintf("Unable to validate placetype (%s) for %s", placetype, path))
			return err
		}

		if !allow {
			return nil
		}

		// THIS NEEDS TO BE MORE SOPHISTICATED THAN "just placetype" BUT
		// TODAY IT IS NOT... (20180531/thisisaaronland)

		target := placetype

		row, err := meta.FeatureToRow(f.Bytes())

		if err != nil {
			return err
		}

		var r repo.Repo

		if opts.Strict {
			r, err = repo.NewDataRepoFromString(whosonfirst.Repo(f))
		} else {

			r, err = repo.NewCustomRepoFromString(whosonfirst.Repo(f))
		}

		if err != nil {
			return err
		}

		mu.Lock()
		defer mu.Unlock()

		writer, ok := writers[target]

		if !ok {

			fieldnames := make([]string, 0)

			for k, _ := range row {
				fieldnames = append(fieldnames, k)
			}

			sort.Strings(fieldnames)

			var fname string

			if opts.Combined {

				if opts.CombinedName == "" {
					return errors.New("Missing opts.CombinedName")
				}

				if strings.HasSuffix(opts.CombinedName, ".csv") {
					fname = opts.CombinedName
				} else {
					fname = fmt.Sprintf("%s.csv", opts.CombinedName)
				}

			} else {

				repo_opts := repo.DefaultFilenameOptions()
				repo_opts.Placetype = placetype
				fname = r.MetaFilename(repo_opts)
			}

			root := opts.Workdir

			// this is just for backwards compatibility
			// (20180531/thisisaaronland)

			if root == "" && mode == "repo" {
				root = filepath.Join(root, "meta")
			}

			outfile := filepath.Join(root, fname)

			fh, err := atomicfile.New(outfile, os.FileMode(0644))

			if err != nil {
				return err
			}

			writer, err = csv.NewDictWriter(fh, fieldnames)

			if err != nil {
				return err
			}

			writer.WriteHeader()

			filehandles[placetype] = fh
			writers[placetype] = writer
			paths[placetype] = outfile
		}

		writer.WriteRow(row)

		atomic.AddInt32(&count, 1)
		return nil
	}

	i, err := wof_index.NewIndexer(mode, cb)

	if err != nil {
		return nil, err
	}

	var index_err error

	t1 := time.Now()

	for _, to_index := range indices {

		ta := time.Now()

		err := i.IndexPath(to_index)

		tb := time.Since(ta)

		if opts.Timings {
			log.Printf("time to prepare %s %v\n", to_index, tb)
		}

		if err != nil {
			index_err = err
			break
		}
	}

	t2 := time.Since(t1)

	if opts.Timings {
		c := atomic.LoadInt32(&count)
		log.Printf("time to prepare all %d records %v\n", c, t2)
	}

	metafiles := make([]string, 0)

	for placetype, fh := range filehandles {

		if index_err != nil {
			fh.Abort()
		} else {
			fh.Close()

			path, ok := paths[placetype]

			if ok {
				metafiles = append(metafiles, path)
			}

		}
	}

	return metafiles, nil
}
