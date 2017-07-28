package main

// ./bin/wof-bundle-metafiles -force-updates -compress -dest /usr/local/data/bundles/ -mode repo /usr/local/data/whosonfirst-data

import (
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-bundles"
	"github.com/whosonfirst/go-whosonfirst-bundles/compress"
	"github.com/whosonfirst/go-whosonfirst-bundles/hash"
	log "github.com/whosonfirst/go-whosonfirst-log"
	_ "github.com/whosonfirst/go-whosonfirst-repo"
	"io"
	"io/ioutil"
	_ "log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	var dest = flag.String("dest", "", "Where to write files")
	var mode = flag.String("mode", "repo", "...")

	var compress_bundle = flag.Bool("compress", false, "...")
	var dated = flag.Bool("dated", false, "...")

	var skip_existing = flag.Bool("skip-existing", false, "Skip existing files on disk (without checking for remote changes)")
	var force_updates = flag.Bool("force-updates", false, "Force updates to files on disk (without checking for remote changes)")

	var loglevel = flag.String("loglevel", "info", "The level of detail for logging")
	// var strict = flag.Bool("strict", false, "Exit (1) if any meta file fails cloning")

	flag.Parse()
	args := flag.Args()

	stdout := io.Writer(os.Stdout)
	stderr := io.Writer(os.Stderr)

	logger := log.NewWOFLogger("wof-bundle-metafiles")
	logger.AddLogger(stdout, *loglevel)
	logger.AddLogger(stderr, "error")

	if *mode == "repo" {

		for _, path := range args {

			abs_repo, err := filepath.Abs(path)

			if err != nil {
				logger.Fatal("failed to make absolute path for %s, because %s", path, err)
			}

			// please make sure these exist...

			abs_meta := filepath.Join(abs_repo, "meta")
			abs_data := filepath.Join(abs_repo, "data")

			metafiles := make([]string, 0)

			files, err := ioutil.ReadDir(abs_meta)

			if err != nil {
				logger.Fatal("failed to readdir for %s, because %s", abs_meta, err)
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
				opts.Logger = logger

				b, err := bundles.NewBundle(opts)

				if err != nil {
					logger.Fatal("failed to create new bundle for %s (%s), because %s", metafile, bundle_name, err)
				}

				if !*force_updates {

					sha1_path, err := compress.CompressedFilePath(metafile, opts.Destination)

					if err != nil {
						logger.Fatal("failed to determined compressed path for %s, because %s", metafile, err)
					}

					_, err = os.Stat(sha1_path)

					if !os.IsNotExist(err) {

						last_hash, err := hash.ReadHashFile(sha1_path)

						if err != nil {
							logger.Fatal("failed to read hash file for %s, because %s", sha1_path, err)
						}

						current_hash, err := hash.HashFile(metafile)

						if err != nil {
							logger.Fatal("failed to hash metafile %s, because %s", metafile, err)
						}

						if last_hash == current_hash {
							logger.Info("no changes to %s, skipping", metafile)
							continue
						}
					}
				}

				bundle_path, err := b.BundleMetafile(metafile)

				if err != nil {
					logger.Fatal("failed to bundle metafile %s, because %s", metafile, err)
				}

				logger.Info("%s", bundle_path)

				if *compress_bundle {

					chroot := opts.Destination

					compress_opts := compress.DefaultCompressOptions()
					compress_opts.RemoveSource = true

					compressed_path, err := compress.CompressBundle(bundle_path, chroot, compress_opts)

					if err != nil {
						logger.Fatal("failed to compress bundle %s, because %s", bundle_path, err)
					}

					sha1_path, err := hash.WriteHashFile(compressed_path)

					if err != nil {
						logger.Fatal("failed to write hash file for %s, because %s", compressed_path, err)
					}

					logger.Info(compressed_path)
					logger.Info(sha1_path)
				}

				compress_opts := compress.DefaultCompressOptions()
				compressed_path, err := compress.CompressFile(metafile, opts.Destination, compress_opts)

				if err != nil {
					logger.Fatal("failed to compresse metafile %s, because %s", metafile, err)
				}

				sha1_path, err := hash.WriteHashFile(compressed_path)

				if err != nil {
					logger.Fatal("failed to write hash file for %s, because %s", compressed_path, err)
				}

				logger.Info("%s (%s)", compressed_path, sha1_path)
			}
		}

	} else {
		logger.Fatal("Unsupported mode '%s'", *mode)
	}

	os.Exit(0)
}
