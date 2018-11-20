package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-meta/stats"
	"log"
)

func main() {

	flag.Parse()

	for _, path := range flag.Args() {

		st, err := stats.Compile(path)

		if err != nil {
			log.Fatal(err)
		}

		b, err := json.Marshal(st)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(b))
	}
}
