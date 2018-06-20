package main

import (
	"flag"
	"github.com/facebookgo/atomicfile"
	"github.com/whosonfirst/go-whosonfirst-meta"
	"io"
	"log"
	"os"
	"time"
)

func main() {

	var input = flag.String("source", "", "...")
	var output = flag.String("dest", "", "...")
	var stdout = flag.Bool("stdout", false, "...")

	flag.Parse()

	/*

	   if *output == "" {
	   	*output = *input
	   }

	*/

	src, err := os.Open(*input)

	if err != nil {
		log.Fatal(err)
	}

	dest, err := atomicfile.New(*output, os.FileMode(0644))

	if err != nil {
		log.Fatal(err)
	}

	defer dest.Close()

	writers := []io.Writer{
		dest,
	}

	if *stdout {
		writers = append(writers, os.Stdout)
	}

	multi := io.MultiWriter(writers...)

	updated := make([]string, 0)

	t1 := time.Now()

	meta.UpdateMetafile(src, multi, updated)

	t2 := time.Since(t1)
	log.Printf("time to update %v\n", t2)
}
