package main

// ./bin/wof-bundle-metafiles -force-updates -compress -dest /usr/local/data/bundles/ -mode repo /usr/local/data/whosonfirst-data

import (
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-bundles"
	"github.com/whosonfirst/go-whosonfirst-bundles/compress"
	"github.com/whosonfirst/go-whosonfirst-bundles/hash"
	log "github.com/whosonfirst/go-whosonfirst-log"
	"io"
	"io/ioutil"
	_ "log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {

	var dest = flag.String("dest", "", "Where to write files")
	var mode = flag.String("mode", "repo", "...")

	var compress_bundle = flag.Bool("compress", false, "...")
	var dated = flag.Bool("dated", false, "...")

	var skip_existing = flag.Bool("skip-existing", false, "Skip existing files on disk (without checking for remote changes)")
	var force = flag.Bool("force", false, "Force updates to files (regardless of whether a metafile has changed)")

	var loglevel = flag.String("loglevel", "status", "The level of detail for logging")
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

			abs_meta := filepath.Join(abs_repo, "meta")

			_, err = os.Stat(abs_meta)

			if os.IsNotExist(err) {
				logger.Fatal("meta directory %s is missing", abs_meta)
			}

			abs_data := filepath.Join(abs_repo, "data")

			_, err = os.Stat(abs_data)

			if os.IsNotExist(err) {
				logger.Fatal("data directory %s is missing", abs_data)
			}

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

				t1 := time.Now()

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
				opts.ForceUpdates = *force
				opts.Logger = logger

				b, err := bundles.NewBundle(opts)

				if err != nil {
					logger.Fatal("failed to create new bundle for %s (%s), because %s", metafile, bundle_name, err)
				}

				compressed_metafile_path, err := compress.CompressedFilePath(metafile, opts.Destination)

				if err != nil {
					logger.Fatal("failed to determined compressed path for %s, because %s", metafile, err)
				}

				if !*force {

					sha1_path := hash.HashFilePath(compressed_metafile_path)

					_, err = os.Stat(sha1_path)

					if !os.IsNotExist(err) {

						logger.Info("comparing hashes for %s", compressed_metafile_path)

						last_hash, err := hash.ReadHashFile(sha1_path)

						if err != nil {
							logger.Fatal("failed to read hash file for %s, because %s", sha1_path, err)
						}

						compress_opts := compress.DefaultCompressOptions()
						compressed_metafile_path, err = compress.CompressFile(metafile, opts.Destination, compress_opts)

						if err != nil {
							logger.Fatal("failed to compresse metafile %s, because %s", metafile, err)
						}

						current_hash, err := hash.HashFile(compressed_metafile_path)

						if err != nil {
							logger.Fatal("failed to hash metafile %s, because %s", compressed_metafile_path, err)
						}

						logger.Debug("last hash (%s) is %s", sha1_path, last_hash)
						logger.Debug("last hash (%s) is %s", metafile, current_hash)

						if last_hash == current_hash {
							logger.Status("no changes to %s, skipping", compressed_metafile_path)
							continue
						}
					}
				}

				bundle_root, err := b.BundleRoot(metafile)

				if err != nil {
					logger.Fatal("unable to determine bundle root for %s, because %s", metafile, err)
				}

				_, err = os.Stat(bundle_root)

				if !os.IsNotExist(err) {

					err = os.RemoveAll(bundle_root)

					if err != nil {
						logger.Fatal("failed to remove %s, because %s", bundle_root, err)
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

					compressed_bundle_path, err := compress.CompressBundle(bundle_path, chroot, compress_opts)

					if err != nil {
						logger.Fatal("failed to compress bundle %s, because %s", bundle_path, err)
					}

					sha1_bundle_path, err := hash.WriteHashFile(compressed_bundle_path)

					if err != nil {
						logger.Fatal("failed to write hash file for %s, because %s", compressed_bundle_path, err)
					}

					logger.Info(compressed_bundle_path)
					logger.Info(sha1_bundle_path)
				}

				_, err = os.Stat(compressed_metafile_path)

				if os.IsNotExist(err) {

					compress_opts := compress.DefaultCompressOptions()
					compressed_metafile_path, err = compress.CompressFile(metafile, opts.Destination, compress_opts)

					if err != nil {
						logger.Fatal("failed to compresse metafile %s, because %s", metafile, err)
					}
				}

				sha1_metafile_path, err := hash.WriteHashFile(compressed_metafile_path)

				if err != nil {
					logger.Fatal("failed to write hash file for %s, because %s", compressed_metafile_path, err)
				}

				logger.Info("%s (%s)", compressed_metafile_path, sha1_metafile_path)

				t2 := time.Since(t1)
				logger.Status("finished bundling %s in %v", metafile, t2)
			}
		}

	} else {
		logger.Fatal("Unsupported mode '%s'", *mode)
	}

	os.Exit(0)
}
