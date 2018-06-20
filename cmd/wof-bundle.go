package main

// MOST OF THIS CODE WILL MOVE IN TO bundle/*.go
// (20180620/thisisaaronland)

import (
       "context"
	"flag"
	"github.com/facebookgo/atomicfile"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"		
	"github.com/whosonfirst/go-whosonfirst-index"
	"github.com/whosonfirst/go-whosonfirst-index/utils"
	log "github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"io"
	_ "log"
	"os"
	"path/filepath"
)

func main() {

	var dest = flag.String("dest", "", "Where to write files")
	var mode = flag.String("mode", "repo", "...")

	var loglevel = flag.String("loglevel", "status", "The level of detail for logging")
	// var strict = flag.Bool("strict", false, "Exit (1) if any meta file fails cloning")

	flag.Parse()
	args := flag.Args()

	stdout := io.Writer(os.Stdout)
	stderr := io.Writer(os.Stderr)

	logger := log.NewWOFLogger("wof-bundle-metafiles")
	logger.AddLogger(stdout, *loglevel)
	logger.AddLogger(stderr, "error")

	// SOMETHING SOMETHING SOMETHING CREATE METAFILE writer INSIDE
	// OF *root HERE...
	
	f := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		path, err := index.PathForContext(ctx)

		if err != nil {
			return err
		}

		if !utils.IsWOFFile(path) {
			return nil
		}

		f, err := feature.LoadFeatureFromReader(fh)

		if err != nil {
			return err
		}

		id := whosonfirst.Id(f)

		abs_path, err := uri.Id2AbsPath(*root, id)

		if err != nil {
			return nil
		}

		abs_root := filepath.Dir(abs_path)

		_, err := os.Stat(abs_root)

		if os.IsNotExist(err) {

			// SOMETHING SOMETHING SOMETHING LOCK/UNLOCK MUTEX HERE
			
			err = os.MkdirAll(abs_root, 0755)

			if err != nil {
				return err
			}
		}

		out, err := atomicfile.New(abs_path, 0644)

		if err != nil {
			return err
		}

		_, err = io.Copy(out, fh)

		if err != nil {
			out.Abort()
			return err
		}

		out.Close()

		// SOMETHING SOMETHING SOMETHING WRITE f TO METAFILE HERE...
		
		return nil
	}

	idx, err := index.NewIndexer(*mode, f)

	if err != nil {
		log.Fatal(err)
	}

	for _, path := range flag.Args() {

		err := idx.IndexPath(path)

		if err != nil {
			log.Fatal(err)
		}
	}

}
