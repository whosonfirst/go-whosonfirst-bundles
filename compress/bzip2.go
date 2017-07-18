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
		"-cjf", // -c is for create; -j is for bzip; -f if for file
		cpath,
		path,
	}

	cmd := exec.Command(tar, args...)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	// fmt.Println(tar, args, out.String())

	if err != nil {
		return "", err
	}

	return cpath, nil
}
