package prune

import (
	"github.com/whosonfirst/go-whosonfirst-bundles/hash"
	"io/ioutil"
	_ "log"
	"os"
	"path/filepath"
)

type LocalPruner struct {
	Pruner
	Root    string
	Options *PruneOptions
}

type LocalFile struct {
	File
	info os.FileInfo
}

func NewLocalFile(info os.FileInfo) (File, error) {

	file := LocalFile{
		info: info,
	}

	return &file, nil
}

func (f *LocalFile) Name() string {
	return f.info.Name()
}

func NewLocalPruner(root string, opts *PruneOptions) (Pruner, error) {

	p := LocalPruner{
		Root:    root,
		Options: opts,
	}

	return &p, nil
}

func (p *LocalPruner) ListFiles() ([]File, error) {

	files, err := ioutil.ReadDir(p.Root)

	if err != nil {
		return nil, err
	}

	localfiles := make([]File, 0)

	for _, file := range files {

		localfile, err := NewLocalFile(file)

		if err != nil {
			return nil, err
		}

		localfiles = append(localfiles, localfile)
	}

	return localfiles, nil
}

func (p *LocalPruner) PruneFiles(files []File) error {

	candidates, err := FilesToCandidates(files, p.Options)

	if err != nil {
		return err
	}

	max_bundles := p.Options.MaxBundles

	for short_name, files := range candidates {

		if len(files) <= max_bundles {
			p.Options.Logger.Printf("SKIP %s, does not exceed max bundles\n", short_name)
			continue
		}

		for c := len(files); c > max_bundles; c-- {

			f := files[0]
			fname := f.Name()

			bundle_path := filepath.Join(p.Root, fname)
			bundle_hash := hash.HashFilePath(bundle_path)

			to_remove := []string{bundle_path, bundle_hash}

			for _, path := range to_remove {

				p.Options.Logger.Printf("PRUNE local file %s\n", path)

				if p.Options.Debug {
					p.Options.Logger.Printf("SKIP pruning local file %s, because debugging is enabled\n", path)
					continue
				}

				_, err := os.Stat(path)

				if os.IsNotExist(err) {
					continue
				}

				err = os.Remove(path)

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
