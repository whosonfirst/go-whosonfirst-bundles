package main

import (
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-cli/flags"
	"github.com/whosonfirst/go-whosonfirst-repo"
	"log"
)

func main() {

	var names flags.MultiString
	flag.Var(&names, "name", "One or more canned filename types. Valid options are: bundle, concordances, meta, sqlite or all")

	placetype := flag.String("placetype", "", "Specify a specific placetype when generating a filename")
	dated := flag.Bool("dated", false, "Use the current YYYYMMDD date as the suffix for a filename")
	old_skool := flag.Bool("old-skool", false, "Use old skool 'wof-' prefix when generating filenames")

	flag.Parse()

	meta_name := false
	sqlite_name := false
	bundle_name := false
	concordances_name := false

	for _, n := range names {

		switch n {
		case "all":
			meta_name = true
			sqlite_name = true
			bundle_name = true
			concordances_name = true
			break
		case "sqlite":
			sqlite_name = true
		case "bundle":
			bundle_name = true
		case "concordances":
			concordances_name = true
		default:
			// pass
		}
	}

	opts := repo.DefaultFilenameOptions()

	if *placetype != "" {
		opts.Placetype = *placetype
	}

	if *dated {
		opts.Suffix = "{DATED}"
	}

	if *old_skool {
		opts.OldSkool = true
	}

	for _, name := range flag.Args() {

		r, err := repo.NewDataRepoFromString(name)

		if err != nil {
			log.Fatal(err)
		}

		if r.String() != name {
			msg := fmt.Sprintf("Expected '%s' but got '%s'", name, r.String())
			log.Fatal(msg)
		}

		fmt.Printf("%s\tOK\n", name)

		if sqlite_name == true {
			fmt.Printf("sqlite filename\t%s\n", r.SQLiteFilename(opts))
		}

		if meta_name == true {
			fmt.Printf("meta filename\t%s\n", r.MetaFilename(opts))
		}

		if bundle_name == true {
			fmt.Printf("bundle filename\t%s\n", r.BundleFilename(opts))
		}

		if concordances_name == true {
			fmt.Printf("concordances filename\t%s\n", r.ConcordancesFilename(opts))
		}

	}
}
