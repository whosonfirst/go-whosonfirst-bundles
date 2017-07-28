package main

/*

given a metafile:

      - calculate the hash (mhash) for that file (uncompressed)
      - generate a filename (hfile) that stores the contents of the hash

      - check for the existence of hfile
      - compare the contents of hfile against mhash

      - if hfile == mhash and !force, exit

      (OR NOT - MAYBE JUST ALWAYS COMPRESS AND HASH THE METAFILE AND BE DONE WITH IT...)

      - generate bundle for metafile
      - compress the bundle for metafile
      - generate hash and hashfile for (compressed) bundle

      - compress metafile
      - generate hash and hashfile for (compressed) metafile

      - generate hashfile for (uncompressed) metafile

      - cp compressed bundle, compressed bundle hash, compressed metafile, compressed metafile hash to (someplace)

*/

import (
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-bundles"
	"github.com/whosonfirst/go-whosonfirst-bundles/compress"
	"github.com/whosonfirst/go-whosonfirst-bundles/hash"
	// "github.com/whosonfirst/go-whosonfirst-repo"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	// var source = flag.String("source", "https://s3.amazonaws.com/whosonfirst.mapzen.com/data/", "Where to look for files")
	var dest = flag.String("dest", "", "Where to write files")

	var mode = flag.String("mode", "repo", "...")

	var compress_bundle = flag.Bool("compress", false, "...")
	var dated = flag.Bool("dated", false, "...")

	var skip_existing = flag.Bool("skip-existing", false, "Skip existing files on disk (without checking for remote changes)")
	var force_updates = flag.Bool("force-updates", false, "Force updates to files on disk (without checking for remote changes)")

	// var procs = flag.Int("procs", (runtime.NumCPU() * 2), "The number of concurrent processes to clone data with")
	// var loglevel = flag.String("loglevel", "info", "The level of detail for logging")
	// var strict = flag.Bool("strict", false, "Exit (1) if any meta file fails cloning")

	flag.Parse()
	args := flag.Args()

	if *mode == "repo" {

		for _, path := range args {

			abs_repo, err := filepath.Abs(path)

			if err != nil {
				log.Fatal(err)
			}

			// please make sure these exist...

			abs_meta := filepath.Join(abs_repo, "meta")
			abs_data := filepath.Join(abs_repo, "data")

			metafiles := make([]string, 0)

			files, err := ioutil.ReadDir(abs_meta)

			if err != nil {
				log.Fatal(err)
			}

			for _, file := range files {

				fname := file.Name()

				if strings.HasSuffix(fname, "-meta.csv") {
					metafiles = append(metafiles, fname)
					continue
				}

				if strings.HasSuffix(fname, "-meta-latest.csv") {
					metafiles = append(metafiles, fname)
					continue
				}

				if strings.HasSuffix(fname, "-latest.csv") && !strings.HasSuffix(fname, "-concordances-latest.csv") {
					metafiles = append(metafiles, fname)
					continue
				}
			}

			for _, fname := range metafiles {

				metafile := filepath.Join(abs_meta, fname)

				// sudo make this part of go-wof-repo... maybe?

				bundle_name := fname

				bundle_name = strings.Replace(bundle_name, "-meta.csv", "", -1)
				bundle_name = strings.Replace(bundle_name, "-latest.csv", "", -1)
				bundle_name = strings.Replace(bundle_name, "-meta-latest.csv", "", -1)

				// log.Println("bundle name", bundle_name)

				opts := bundles.DefaultBundleOptions()

				opts.Source = fmt.Sprintf("file://%s", abs_data)
				opts.Destination = *dest
				opts.BundleName = bundle_name
				opts.Compress = *compress_bundle
				opts.Dated = *dated
				opts.SkipExisting = *skip_existing
				opts.ForceUpdates = *force_updates

				b, err := bundles.NewBundle(opts)

				if err != nil {
					log.Fatal(err)
				}

				if err != nil {
					log.Fatal(err)
				}

				if !*force_updates {

					sha1_path, err := compress.CompressedFilePath(metafile, opts.Destination)

					if err != nil {
						log.Fatal(err)
					}

					_, err = os.Stat(sha1_path)

					if !os.IsNotExist(err) {

						last_hash, err := hash.ReadHashFile(sha1_path)

						if err != nil {
							log.Fatal(err)
						}

						current_hash, err := hash.HashFile(metafile)

						if err != nil {
							log.Fatal(err)
						}

						if last_hash == current_hash {
							log.Println("no changes", metafile)
							continue
						}
					}
				}

				bundle_path, err := b.BundleMetafile(metafile)

				if err != nil {
					log.Fatal(err)
				}

				log.Println(bundle_path)

				if *compress_bundle {

					chroot := opts.Destination

					compress_opts := compress.DefaultCompressOptions()
					compress_opts.RemoveSource = true

					compressed_path, err := compress.CompressBundle(bundle_path, chroot, compress_opts)

					if err != nil {
						log.Fatal(err)
					}

					sha1_path, err := hash.WriteHashFile(compressed_path)

					if err != nil {
						log.Fatal(err)
					}

					log.Println(compressed_path)
					log.Println(sha1_path)
				}

				compress_opts := compress.DefaultCompressOptions()
				compressed_path, err := compress.CompressFile(metafile, opts.Destination, compress_opts)

				if err != nil {
					log.Fatal(err)
				}

				sha1_path, err := hash.WriteHashFile(compressed_path)

				if err != nil {
					log.Fatal(err)
				}

				log.Println(compressed_path)
				log.Println(sha1_path)
			}
		}

	} else {
		log.Fatal("Unsupported mode")
	}

	os.Exit(0)
}
