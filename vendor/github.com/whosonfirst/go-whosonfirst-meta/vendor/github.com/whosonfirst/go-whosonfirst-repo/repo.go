package repo

import (
	"errors"
	"fmt"
	_ "github.com/whosonfirst/go-whosonfirst-placetypes"
	_ "log"
	"path/filepath"
	"strings"
	"time"
)

type Repo interface {
	String() string
	Name() string
	ConcordancesFilename(*FilenameOptions) string
	MetaFilename(*FilenameOptions) string
	SQLiteFilename(*FilenameOptions) string
	BundleFilename(*FilenameOptions) string
}

type DataRepo struct {
	Repo
	Source    string
	Role      string
	Placetype string
	Country   string
	Region    string
	Filter    string // PLEASE DON'T CALL ME 'Filter' ...
}

type FilenameOptions struct {
	Placetype string
	Suffix    string
	Extension string
	OldSkool  bool
}

func DefaultFilenameOptions() *FilenameOptions {

	o := FilenameOptions{
		Placetype: "",
		Suffix:    "latest",
		Extension: "",
		OldSkool:  false,
	}

	return &o
}

func NewDataRepoFromPath(path string) (*DataRepo, error) {

	abs_path, err := filepath.Abs(path)

	if err != nil {
		return nil, err
	}

	repo := filepath.Base(abs_path)

	return NewDataRepoFromString(repo)
}

func NewDataRepoFromString(repo string) (*DataRepo, error) {

	parts := strings.Split(repo, "-")

	if len(parts) < 2 {
		return nil, errors.New("Invalid repo name (too short)")
	}

	if len(parts) > 6 {
		return nil, errors.New("Invalid repo name (too long)")
	}

	r := DataRepo{
		Source:    "",
		Role:      "",
		Placetype: "",
		Country:   "",
		Region:    "",
		Filter:    "",
	}

	r.Source = parts[0]
	r.Role = parts[1]

	if r.Role != "data" {
		return nil, errors.New("Unsupported role")
	}

	if len(parts) > 2 {

		placetype := parts[2]

		/*
			if opts.StrictPlacetypes && !placetypes.IsValidPlacetype(placetype) {
				return nil, errors.New("Invalid placetype")
			}
		*/

		r.Placetype = placetype
	}

	if len(parts) > 3 {

		country := parts[3]

		if len(country) != 2 {
			return nil, errors.New("Invalid country code")
		}

		// to do: validate country code

		r.Country = country
	}

	if len(parts) > 4 {

		region := parts[4]

		if len(region) != 2 {
			return nil, errors.New("Invalid region code")
		}

		// to do: validate region code

		r.Region = region
	}

	if len(parts) > 5 {

		filter := parts[5]
		r.Filter = filter
	}

	return &r, nil
}

func (r *DataRepo) String() string {
	return r.Name()
}

func (r *DataRepo) Name() string {

	parts := make([]string, 0)

	parts = append(parts, r.Source)
	parts = append(parts, r.Role)

	if r.Placetype != "" {
		parts = append(parts, r.Placetype)
	}

	if r.Country != "" {
		parts = append(parts, r.Country)
	}

	if r.Region != "" {
		parts = append(parts, r.Region)
	}

	if r.Filter != "" {
		parts = append(parts, r.Filter)
	}

	return strings.Join(parts, "-")
}

func (r *DataRepo) MetaFilename(opts *FilenameOptions) string {

	opts.Extension = "csv"
	return r.filename(opts)
}

func (r *DataRepo) ConcordancesFilename(opts *FilenameOptions) string {

	opts.Suffix = "concordances"
	opts.Extension = "csv"

	return r.filename(opts)
}

func (r *DataRepo) BundleFilename(opts *FilenameOptions) string {

	opts.Extension = ""
	return r.filename(opts)
}

func (r *DataRepo) SQLiteFilename(opts *FilenameOptions) string {

	opts.Extension = "db"
	return r.filename(opts)
}

func (r *DataRepo) filename(opts *FilenameOptions) string {

	parts := make([]string, 0)

	if r.Source == "whosonfirst" && opts.OldSkool {

		parts = append(parts, "wof")
	} else {

		parts = append(parts, r.Source)
		parts = append(parts, r.Role)
	}

	if r.Placetype != "" {
		parts = append(parts, r.Placetype)
	}

	if opts.Placetype != "" && opts.Placetype != r.Placetype {
		parts = append(parts, opts.Placetype)
	}

	if r.Country != "" {
		parts = append(parts, r.Country)
	}

	if r.Region != "" {
		parts = append(parts, r.Region)
	}

	if r.Filter != "" {
		parts = append(parts, r.Filter)
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
