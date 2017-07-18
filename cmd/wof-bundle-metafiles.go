package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-bundles"
	"log"
	"os"
)

func main() {

	var source = flag.String("source", "https://s3.amazonaws.com/whosonfirst.mapzen.com/data/", "Where to look for files")
	var dest = flag.String("dest", "", "Where to write files")

	var compress = flag.Bool("compress", false, "...")
	var dated = flag.Bool("dated", false, "...")

	var skip_existing = flag.Bool("skip-existing", false, "Skip existing files on disk (without checking for remote changes)")
	var force_updates = flag.Bool("force-updates", false, "Force updates to files on disk (without checking for remote changes)")

	// var procs = flag.Int("procs", (runtime.NumCPU() * 2), "The number of concurrent processes to clone data with")
	// var loglevel = flag.String("loglevel", "info", "The level of detail for logging")
	// var strict = flag.Bool("strict", false, "Exit (1) if any meta file fails cloning")

	flag.Parse()
	args := flag.Args()

	opts := bundles.DefaultBundleOptions()

	opts.Source = *source
	opts.Destination = *dest
	opts.Compress = *compress
	opts.Dated = *dated
	opts.SkipExisting = *skip_existing
	opts.ForceUpdates = *force_updates

	b, err := bundles.NewBundle(opts)

	if err != nil {
		log.Fatal(err)
	}

	for _, metafile := range args {

		err = b.BundleMetafile(metafile)

		if err != nil {
			log.Fatal(err)
		}

	}

	os.Exit(0)
}
