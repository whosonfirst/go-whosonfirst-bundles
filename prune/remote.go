package prune

import (
	"github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/awserr"
	// "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/whosonfirst/go-whosonfirst-bundles/hash"
	_ "log"
)

type RemotePruner struct {
	Pruner
	Bucket  string
	Prefix  string
	Region  string
	Service *s3.S3
	Options *PruneOptions
}

type RemoteFile struct {
	File
	object *s3.Object
}

func NewRemoteFile(o *s3.Object) (File, error) {

	file := RemoteFile{
		object: o,
	}

	return &file, nil
}

func (f *RemoteFile) Name() string {
	return *f.object.Key
}

func NewRemotePruner(bucket string, prefix string, region string, opts *PruneOptions) (Pruner, error) {

	cfg := aws.NewConfig()
	cfg.WithRegion(region)

	/*
		if creds != nil {
			cfg.WithCredentials(creds)
		}
	*/

	sess := session.New(cfg)
	svc := s3.New(sess)

	p := RemotePruner{
		Bucket:  bucket,
		Prefix:  prefix,
		Service: svc,
		Options: opts,
	}

	return &p, nil
}

func (p *RemotePruner) ListFiles() ([]File, error) {

	input := &s3.ListObjectsInput{
		Bucket: aws.String(p.Bucket),
		Prefix: aws.String(p.Prefix),
	}

	files, err := p.Service.ListObjects(input)

	if err != nil {
		return nil, err
	}

	localfiles := make([]File, 0)

	for _, file := range files.Contents {

		localfile, err := NewRemoteFile(file)

		if err != nil {
			return nil, err
		}

		localfiles = append(localfiles, localfile)
	}

	return localfiles, nil
}

func (p *RemotePruner) PruneFiles(files []File) error {

	candidates, err := FilesToCandidates(files, p.Options)

	if err != nil {
		return err
	}

	/*
		for name, files := range candidates {
		    log.Println(name, len(files))
		}
	*/

	max_bundles := p.Options.MaxBundles

	for _, files := range candidates {

		if len(files) <= max_bundles {
			continue
		}

		for c := len(files); c > max_bundles; c-- {

			f := files[0]
			fname := f.Name()

			bundle_path := fname
			bundle_hash := hash.HashFilePath(bundle_path)

			to_remove := []string{bundle_path, bundle_hash}

			for _, path := range to_remove {

				if p.Options.Debug {
					continue
				}

				input := &s3.DeleteObjectInput{
					Bucket: aws.String(p.Bucket),
					Key:    aws.String(path),
				}

				_, err := p.Service.DeleteObject(input)

				if err != nil {
					return err
				}
			}
		}

		if len(files) > 1 {
			files = files[1:]
		}
	}

	return nil
}
