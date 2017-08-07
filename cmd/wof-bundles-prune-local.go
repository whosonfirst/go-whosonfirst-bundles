package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-bundles/hash"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

func main() {

	// this name is a bit of a misnomer - it's really about
	// pruning _dated_ bundles (20170807/thisisaaronland)

	// please reconcile me with wof-bundles-prune-remote
	// probably something like a interface for bundle source to
	// list all the files and to prune anything with more than
	// x instances (20170807/thisisaaronland)

	var root = flag.String("root", "", "...")
	var max_bundles = flag.Int("max-bundles", 2, "...")

	var debug = flag.Bool("debug", false, "...")

	flag.Parse()

	files, err := ioutil.ReadDir(*root)

	if err != nil {
		log.Fatal(err)
	}

	/*

		threescompany-venue-us-ca-lagov-20170805T060001-bundle.tar.bz2
		threescompany-venue-us-ca-lagov-20170805T060001-bundle.tar.bz2.sha1.txt
		threescompany-venue-us-ca-lagov-meta.csv.bz2
		threescompany-venue-us-ca-lagov-meta.csv.bz2.sha1.txt
		wof-borough-20170804T223358-bundle.tar.bz2
		wof-borough-20170804T223358-bundle.tar.bz2.sha1.txt
		wof-borough-latest-bundle.tar.bz2
		wof-borough-latest-bundle.tar.bz2.sha1.txt
		wof-borough-latest.csv.bz2
		wof-borough-latest.csv.bz2.sha1.txt

	*/

	// re_bundle := regexp.MustCompile(`^(.*)\-(?:\d{8}T\d{6})\-bundle\.tar\.bz2(?:\.sha1\.txt)?$`)
	re_bundle := regexp.MustCompile(`^(.*)\-(?:\d{8}T\d{6})\-bundle\.tar\.bz2?$`)

	lookup := make(map[string][]os.FileInfo)

	for _, file := range files {

		fname := file.Name()

		if !re_bundle.MatchString(fname) {
			continue
		}

		m := re_bundle.FindAllStringSubmatch(fname, -1)

		if m == nil {
			continue
		}

		short_name := m[0][1]

		_, ok := lookup[short_name]

		if !ok {
			lookup[short_name] = make([]os.FileInfo, 0)
		}

		lookup[short_name] = append(lookup[short_name], file)
	}

	for _, bundles := range lookup {

		if len(bundles) <= *max_bundles {
			continue
		}

		for c := len(bundles); c > *max_bundles; c-- {

			b := bundles[0]
			fname := b.Name()

			bundle_path := filepath.Join(*root, fname)
			bundle_hash := hash.HashFilePath(bundle_path)

			to_remove := []string{bundle_path, bundle_hash}

			for _, path := range to_remove {

				log.Println("REMOVE", path)

				if *debug {
					continue
				}

				_, err := os.Stat(path)

				if os.IsNotExist(err) {
					continue
				}

				err = os.Remove(path)

				if err != nil {
					log.Fatal(err)
				}
			}
		}

		if len(bundles) > 1 {
			bundles = bundles[1:]
		}
	}

	os.Exit(0)
}
