package hash

import (
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-hash"
	"io/ioutil"
	"os"
)

func ReadHashFile(source string) (string, error) {

	fh, err := os.Open(source)

	if err != nil {
		return "", nil
	}

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return "", nil
	}

	return string(body), nil
}

func WriteHashFile(source string) (string, error) {

	dest := HashFilePath(source)
	hash, err := HashFile(source)

	if err != nil {
		return "", err
	}

	fh, err := os.Create(dest)

	if err != nil {
		return "", err
	}

	fh.WriteString(hash)
	fh.Close()

	return dest, nil
}

func HashFilePath(path string) string {
	return fmt.Sprintf("%s.sha1.txt", path)
}

func HashFile(path string) (string, error) {
	h, err := hash.NewHash("sha1")

	if err != nil {
		return "", err
	}

	return h.HashFile(path)
}
