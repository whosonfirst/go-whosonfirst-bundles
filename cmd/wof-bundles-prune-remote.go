package main

import (
	"flag"
	"github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/awserr"
	// "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"path/filepath"
	"regexp"
)

func main() {

	// please reconcile me with wof-bundles-prune-local
	// probably something like a interface for bundle source to
	// list all the files and to prune anything with more than
	// x instances (20170807/thisisaaronland)

	var bucket = flag.String("bucket", "", "...")
	var prefix = flag.String("prefix", "", "...")
	var region = flag.String("region", "us-east-1", "...")
	var max_bundles = flag.Int("max-bundles", 2, "...")

	var debug = flag.Bool("debug", false, "...")

	flag.Parse()

	re_bundle := regexp.MustCompile(`^(.*)\-(?:\d{8}T\d{6})\-bundle\.tar\.bz2?$`)

	cfg := aws.NewConfig()
	cfg.WithRegion(*region)

	/*
		if creds != nil {
			cfg.WithCredentials(creds)
		}
	*/

	sess := session.New(cfg)
	svc := s3.New(sess)

	input := &s3.ListObjectsInput{
		Bucket: aws.String(*bucket),
		Prefix: aws.String(*prefix),
	}

	result, err := svc.ListObjects(input)

	if err != nil {
		log.Fatal(err)
	}

	lookup := make(map[string][]*s3.Object)

	for _, file := range result.Contents {

		key := file.Key
		fname := filepath.Base(*key)

		if !re_bundle.MatchString(fname) {
			continue
		}

		m := re_bundle.FindAllStringSubmatch(fname, -1)

		if m == nil {
			continue
		}

		short_name := m[0][1]

		_, ok := lookup[short_name]

		if !ok {
			lookup[short_name] = make([]*s3.Object, 0)
		}

		lookup[short_name] = append(lookup[short_name], file)
	}

	for short_name, bundles := range lookup {

		log.Println(short_name, len(bundles))

		if len(bundles) <= *max_bundles {
			continue
		}

		for _, f := range bundles {
			log.Println(short_name, f.Key)
		}

		if *debug {
			continue
		}
	}

}
