package main

import (
	"flag"
	_ "fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

func main() {

	// please reconcile me with wof-bundles-prune-remote
	// probably something like a interface for bundle source to
	// list all the files and to prune anything with more than
	// x instances (20170807/thisisaaronland)

	var root = flag.String("", "", "...")
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

	for short_name, bundles := range lookup {

		log.Println(short_name, len(bundles))

		if len(bundles) <= *max_bundles {
			continue
		}

		for _, f := range bundles {
			log.Println(short_name, f.Name())
		}

		if *debug {
			continue
		}
	}
}
