package main

import (
	"encoding/csv"
	"flag"
	"github.com/whosonfirst/go-whosonfirst-meta"
	"io/ioutil"
	"log"
	"os"
	"sort"
)

func main() {

	var debug = flag.Bool("debug", false, "...")

	flag.Parse()

	writer := csv.NewWriter(os.Stdout)
	rows := 0

	for _, path := range flag.Args() {

		fh, err := os.Open(path)

		if err != nil {
			log.Fatal(err)
		}

		body, err := ioutil.ReadAll(fh)

		if err != nil {
			log.Fatal(err)
		}

		row, err := meta.DumpFeature(body)

		if err != nil {
			log.Fatal(err)
		}

		if *debug {

			keys := make([]string, 0)

			for k, _ := range row {
				keys = append(keys, k)
			}

			sort.Strings(keys)

			for _, k := range keys {
				log.Printf("[dump][%s] '%s'\n", k, row[k])
			}
		}

		defaults, err := meta.GetDefaults()

		if err != nil {
			log.Fatal(err)
		}

		row = defaults.EnsureDefaults(row)

		if *debug {

			keys := make([]string, 0)

			for k, _ := range row {
				keys = append(keys, k)
			}

			sort.Strings(keys)

			for _, k := range keys {
				log.Printf("[defaults][%s] '%s'\n", k, row[k])
			}
		}

		if *debug {
			continue
		}

		header := make([]string, 0)
		values := make([]string, 0)

		for k, v := range row {

			header = append(header, k)
			values = append(values, v)

			log.Println(k, v)
		}

		if rows == 0 {
			writer.Write(header)
		}

		writer.Write(values)
		writer.Flush()

		rows++
	}

}
