package flags

import (
	"strings"
)

type MultiString []string

func (m *MultiString) String() string {
	return strings.Join(*m, "\n")
}

func (m *MultiString) Set(value string) error {
	*m = append(*m, value)
	return nil
}
