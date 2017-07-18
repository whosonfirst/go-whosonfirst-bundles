package compress

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func CompressBundle(source string, chroot string, opts *CompressOptions) (string, error) {

	dest := fmt.Sprintf("%s.tar.bz", source)

	tar := "tar"

	args := []string{
		"-C", chroot, // -C is for ...
		"-cjf", // -c is for create; -j is for bzip; -f if for file
		dest,
		source,
	}

	cmd := exec.Command(tar, args...)

	// to do : wire the Logger stuff in to this...

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	if err != nil {
		return "", err
	}

	if opts.RemoveSource {

		err = os.RemoveAll(source)

		if err != nil {
			return "", err
		}
	}

	return dest, nil
}
