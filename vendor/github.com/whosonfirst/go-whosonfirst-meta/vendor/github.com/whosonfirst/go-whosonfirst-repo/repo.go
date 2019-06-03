package repo

import ()

type Repo interface {
	String() string
	Name() string
	ConcordancesFilename(*FilenameOptions) string
	MetaFilename(*FilenameOptions) string
	SQLiteFilename(*FilenameOptions) string
	BundleFilename(*FilenameOptions) string
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
