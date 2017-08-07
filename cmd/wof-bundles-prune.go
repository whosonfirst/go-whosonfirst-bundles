package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-bundles/prune"
	"log"
)

func main() {

	var root = flag.String("root", "", "...")
	var max_bundles = flag.Int("max-bundles", 10, "...")

	var remote = flag.Bool("remote", false, "")
	var bucket = flag.String("bucket", "", "...")
	var prefix = flag.String("prefix", "", "...")
	var region = flag.String("region", "us-east-1", "...")

	var debug = flag.Bool("debug", false, "...")

	flag.Parse()

	opts, err := prune.NewDefaultPruneOptions()

	if err != nil {
		log.Fatal(err)
	}

	opts.MaxBundles = *max_bundles
	opts.Debug = *debug

	var pruner prune.Pruner

	if *remote {

		pr, err := prune.NewRemotePruner(*bucket, *prefix, *region, opts)

		if err != nil {
			log.Fatal(err)
		}

		pruner = pr

	} else {

		pr, err := prune.NewLocalPruner(*root, opts)

		if err != nil {
			log.Fatal(err)
		}

		pruner = pr
	}

	files, err := pruner.ListFiles()

	if err != nil {
		log.Fatal()
	}

	err = pruner.PruneFiles(files)

	if err != nil {
		log.Fatal()
	}

}
