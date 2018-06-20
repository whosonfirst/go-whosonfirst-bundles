package main

import (
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-repo"
	"log"
)

func main() {

	flag.Parse()

	opts := repo.DefaultFilenameOptions()

	for _, name := range flag.Args() {

		r, err := repo.NewDataRepoFromString(name, opts)

		if err != nil {
			log.Fatal(err)
		}

		if r.String() != name {
			msg := fmt.Sprintf("Expected '%s' but got '%s'", name, r.String())
			log.Fatal(msg)
		}

		fmt.Printf("%s\tOK\n", name)
	}
}
