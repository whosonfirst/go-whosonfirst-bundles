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
	"log"
	"os"
	"path/filepath"
)

func main() {

	var source = flag.String("source", "https://s3.amazonaws.com/whosonfirst.mapzen.com/data/", "Where to look for files")
	var dest = flag.String("dest", "", "Where to write files")

	var name = flag.String("name", "", "...")

	var compress_bundle = flag.Bool("compress", false, "...")
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
	opts.BundleName = *name
	opts.Compress = *compress_bundle
	opts.Dated = *dated
	opts.SkipExisting = *skip_existing
	opts.ForceUpdates = *force_updates

	b, err := bundles.NewBundle(opts)

	if err != nil {
		log.Fatal(err)
	}

	for _, metafile := range args {

		abs_path, err := filepath.Abs(metafile)

		if err != nil {
			log.Fatal(err)
		}

		if !*force_updates {

			fname := filepath.Base(metafile)
			hname := fmt.Sprintf("%s.sha1.txt", fname)

			hfile := filepath.Join(*dest, hname)

			_, err = os.Stat(hfile)

			if !os.IsNotExist(err) {

				last_hash, err := hash.ReadHashFile(hfile)

				if err != nil {
					log.Fatal(err)
				}

				current_hash, err := hash.HashFile(abs_path)

				if err != nil {
					log.Fatal(err)
				}

				if last_hash == current_hash {
					continue
				}
			}
		}

		//

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

		//

		check_path, err := hash.WriteHashFile(abs_path)

		if err != nil {
			log.Fatal(err)
		}

		log.Println(check_path)
	}

	os.Exit(0)
}
