package hash

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
)

func WriteHashFile(source string) (string, error) {

	dest := fmt.Sprintf("%s.sha1.txt", source)
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

func HashFile(path string) (string, error) {

	fh, err := os.Open(path)

	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(fh)

	if err != nil {
		return "", err
	}

	defer fh.Close()

	hash := sha1.Sum(body)
	enc := hex.EncodeToString(hash[:])

	return enc, nil
}
