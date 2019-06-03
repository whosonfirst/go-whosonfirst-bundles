package main

import (
	"context"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-bundles"
	"github.com/whosonfirst/go-whosonfirst-log"
	"io"
	_ "log"
	"os"
)

func main() {

	var dest = flag.String("dest", "", "Where to write files")
	var mode = flag.String("mode", "repo", "...")

	var sqlite = flag.Bool("sqlite", false, "...")
	var dsn = flag.String("sqlite-dsn", "", "...")

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

	if *sqlite {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err = b.BundleMetafilesFromSQLite(ctx, *dsn, to_index...)

	} else {
		err = b.Bundle(to_index...)
	}

	if err != nil {
		logger.Fatal("Failed to create bundle because %s", err)
	}

	logger.Info("Created bundle in %s", b.Options.Destination)
}
