package main

// For example:
// bin/wof-bundle-metafiles -dest /usr/local/data/bundles -compress /usr/local/data/whosonfirst-data

// And then:
// aws s3 --region us-east-1 sync /usr/local/data/bundles/ s3://whosonfirst.mapzen.com/bundles/

// And then for bonus points:
// utils/bundles3-prune-dated whosonfirst.mapzen.com us-east-1
// utils/bundles3-latest-to-dated whosonfirst.mapzen.com us-east-1
// utils/bundles3-index whosonfirst.mapzen.com us-east-1

// See also:
// https://github.com/whosonfirst/go-whosonfirst-bundles/tree/master/utils#bundles3-

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

type BundleInfo struct {
	MetafilePath               string
	MetafileHashPath           string
	MetafileHash               string
	MetafileCompressedPath     string
	MetafileCompressedHashPath string
	MetafileCompressedHash     string
	BundlePath                 string
	BundleCompressedPath       string
	BundleCompressedHashPath   string
	BundleCompressedHash       string
}

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

			t1 := time.Now()

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

				ta := time.Now()

				metafile := filepath.Join(abs_meta, fname)

				opts := bundles.DefaultBundleOptions()

				opts.Source = fmt.Sprintf("file://%s", abs_data)
				opts.Destination = *dest
				opts.Compress = *compress_bundle
				opts.Dated = *dated
				opts.SkipExisting = *skip_existing
				opts.ForceUpdates = *force
				opts.Logger = logger

				b, err := bundles.NewBundle(opts)

				if err != nil {
					logger.Fatal("failed to create new bundle for %s, because %s", metafile, err)
				}

				// working with compressed metafiles is a known-known, see also:
				// https://github.com/whosonfirst/go-whosonfirst-bundles/issues/1

				compressed_metafile_path, err := compress.CompressedFilePath(metafile, opts.Destination)

				if err != nil {
					logger.Fatal("failed to determined compressed path for %s, because %s", metafile, err)
				}

				// work in progres...

				bi := BundleInfo{
					MetafilePath:               "",
					MetafileHashPath:           "",
					MetafileCompressedPath:     "",
					MetafileCompressedHashPath: "",
					MetafileHash:               "",
					BundlePath:                 "",
					BundleCompressedPath:       "",
					BundleCompressedHashPath:   "",
					BundleCompressedHash:       "",
				}

				sha1_path := hash.HashFilePath(compressed_metafile_path)

				bi.MetafilePath = metafile
				bi.MetafileCompressedPath = compressed_metafile_path
				bi.MetafileCompressedHashPath = sha1_path

				if !*force {

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

						bi.MetafileCompressedHash = current_hash

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

				bi.BundlePath = bundle_path

				if *compress_bundle {

					chroot := opts.Destination

					compress_opts := compress.DefaultCompressOptions()
					compress_opts.RemoveSource = true

					compressed_bundle_path, err := compress.CompressBundle(bundle_path, chroot, compress_opts)

					if err != nil {
						logger.Fatal("failed to compress bundle %s, because %s", bundle_path, err)
					}

					bi.BundleCompressedPath = compressed_bundle_path

					sha1_bundle_path, err := hash.WriteHashFile(compressed_bundle_path)

					if err != nil {
						logger.Fatal("failed to write hash file for %s, because %s", compressed_bundle_path, err)
					}

					bi.BundleCompressedHashPath = sha1_bundle_path
				}

				_, err = os.Stat(compressed_metafile_path)

				if os.IsNotExist(err) {

					compress_opts := compress.DefaultCompressOptions()
					compressed_metafile_path, err = compress.CompressFile(metafile, opts.Destination, compress_opts)

					if err != nil {
						logger.Fatal("failed to compresse metafile %s, because %s", metafile, err)
					}

					bi.MetafileCompressedPath = compressed_metafile_path
				}

				sha1_metafile_path, err := hash.WriteHashFile(compressed_metafile_path)

				if err != nil {
					logger.Fatal("failed to write hash file for %s, because %s", compressed_metafile_path, err)
				}

				bi.MetafileCompressedHashPath = sha1_metafile_path

				if opts.Dated {
					// copy DATED to latest
				}

				tb := time.Since(ta)
				logger.Status("finished bundling %s in %v", metafile, tb)
			}

			t2 := time.Since(t1)
			logger.Status("finished bundling metafiles in %s in %v", abs_repo, t2)

		}

	} else {
		logger.Fatal("Unsupported mode '%s'", *mode)
	}

	os.Exit(0)
}
