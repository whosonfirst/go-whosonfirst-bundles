package main

// MOST OF THIS CODE WILL MOVE IN TO bundle/*.go
// (20180620/thisisaaronland)

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-bundles"
	log "github.com/whosonfirst/go-whosonfirst-log"
	"io"
	_ "log"
	"os"
)

func main() {

	var dest = flag.String("dest", "", "Where to write files")
	var mode = flag.String("mode", "repo", "...")

	var loglevel = flag.String("loglevel", "status", "The level of detail for logging")
	// var strict = flag.Bool("strict", false, "Exit (1) if any meta file fails cloning")

	flag.Parse()

	stdout := io.Writer(os.Stdout)
	stderr := io.Writer(os.Stderr)

	logger := log.NewWOFLogger("wof-bundle-metafiles")
	logger.AddLogger(stdout, *loglevel)
	logger.AddLogger(stderr, "error")

	opts := bundles.DefaultBundleOptions()

	opts.Mode = *mode
	opts.Destination = *dest
	opts.Logger = logger

	b, err := bundles.NewBundle(opts)

	if err != nil {
		logger.Fatal("Failed to create bundle because %s", err)
	}

	to_index := flag.Args()
	err = b.Bundle(to_index...)

	if err != nil {
		logger.Fatal("Failed to create bundle because %s", err)
	}

	logger.Info("Created bundle in %s", b.Options.Destination)
}
