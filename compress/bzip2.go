package compress

import (
	"bytes"
	"fmt"
	"os/exec"
)

func CompressBundle(path string) (string, error) {

	cpath := fmt.Sprintf("%s.tar.bz2", path)

	tar := "tar"

	args := []string{
		"-cvjf",
		path,
		cpath,
	}

	cmd := exec.Command(tar, args...)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		return "", err
	}

	return cpath, nil
}
