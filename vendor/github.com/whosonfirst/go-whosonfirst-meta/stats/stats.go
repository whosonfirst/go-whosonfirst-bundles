package stats

import (
	"errors"
	"github.com/whosonfirst/go-whosonfirst-csv"
	"io"
	"strconv"
	"sync"
)

type Stats struct {
	Path       string   `json:"path"`
	Count      int64    `json:"count"`
	LastUpdate int64    `json:"last_update"`
	Placetypes []string `json:"placetypes"`
	mu         *sync.Mutex
}

func (s *Stats) update(row map[string]string) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	pt, ok := row["placetype"]

	if !ok {
		return errors.New("Missing placetype")
	}

	seen := false

	for _, test := range s.Placetypes {

		if test == pt {
			seen = true
			break
		}
	}

	if !seen {
		s.Placetypes = append(s.Placetypes, pt)
	}

	str_lastmod, ok := row["lastmodified"]

	if !ok {
		return errors.New("Missing lastmodified")
	}

	lastmod, err := strconv.ParseInt(str_lastmod, 10, 64)

	if err != nil {
		return err
	}

	if lastmod > s.LastUpdate {
		s.LastUpdate = lastmod
	}

	s.Count += 1
	return nil
}

func newStats(path string) *Stats {

	mu := new(sync.Mutex)

	s := Stats{
		Path:       path,
		Count:      0,
		LastUpdate: 0,
		Placetypes: make([]string, 0),
		mu:         mu,
	}

	return &s
}

func Compile(path string) (*Stats, error) {

	reader, err := csv.NewDictReaderFromPath(path)

	if err != nil {
		return nil, err
	}

	s := newStats(path)

	for {
		row, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		err = s.update(row)

		if err != nil {
			return nil, err
		}

	}

	return s, nil
}
