package repo

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type CustomRepo struct {
	Repo
	name string
}

func NewCustomRepoFromPath(path string, opts *FilenameOptions) (Repo, error) {

	abs_path, err := filepath.Abs(path)

	if err != nil {
		return nil, err
	}

	if opts.Extension != "" && strings.HasSuffix(abs_path, opts.Extension) {
		abs_path = strings.Replace(abs_path, opts.Extension, "", -1)
	}

	if opts.Suffix != "" {

		fq_suffix := fmt.Sprintf("-%s", opts.Suffix)

		if strings.HasSuffix(abs_path, fq_suffix) {
			abs_path = strings.Replace(abs_path, fq_suffix, "", -1)
		}
	}

	repo := filepath.Base(abs_path)

	return NewCustomRepoFromString(repo)
}

func NewCustomRepoFromMetafile(path string) (Repo, error) {

	opts := DefaultFilenameOptions()
	opts.Extension = ".csv"

	return NewCustomRepoFromPath(path, opts)
}

func NewCustomRepoFromSQLitefile(path string) (Repo, error) {

	opts := DefaultFilenameOptions()
	opts.Extension = ".db"

	return NewCustomRepoFromPath(path, opts)
}

func NewCustomRepoFromString(repo string) (Repo, error) {

	r := CustomRepo{
		name: repo,
	}

	return &r, nil
}

func (r *CustomRepo) String() string {
	return r.Name()
}

func (r *CustomRepo) Name() string {

	return r.name
}

func (r *CustomRepo) MetaFilename(opts *FilenameOptions) string {

	opts.Extension = "csv"
	return r.filename(opts)
}

func (r *CustomRepo) ConcordancesFilename(opts *FilenameOptions) string {

	opts.Suffix = "concordances"
	opts.Extension = "csv"

	return r.filename(opts)
}

func (r *CustomRepo) BundleFilename(opts *FilenameOptions) string {

	opts.Extension = ""
	return r.filename(opts)
}

func (r *CustomRepo) SQLiteFilename(opts *FilenameOptions) string {

	opts.Extension = "db"
	return r.filename(opts)
}

func (r *CustomRepo) filename(opts *FilenameOptions) string {

	parts := []string{
		r.name,
	}

	if opts.Suffix != "" {

		suffix := opts.Suffix

		if opts.Suffix == "{DATED}" {

			now := time.Now()
			suffix = now.Format("20060102")
		}

		parts = append(parts, suffix)
	}

	fname := strings.Join(parts, "-")

	if opts.Extension != "" {
		fname = fmt.Sprintf("%s.%s", fname, opts.Extension)
	}

	return fname
}
