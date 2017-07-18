package compress

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func CompressBundle(source string, chroot string, opts *CompressOptions) (string, error) {

	abs_source, err := filepath.Abs(source)

	if err != nil {
		return "", err
	}

	abs_chroot, err := filepath.Abs(chroot)

	if err != nil {
		return "", err
	}

	dest := fmt.Sprintf("%s.tar.bz2", abs_source)

	tar := "tar"

	args := []string{
		"-C", abs_chroot, // -C is for chroot
		"-cjf", // -c is for create; -j is for bzip; -f if for file
		dest,
		source,
	}

	cmd := exec.Command(tar, args...)

	// to do : wire the Logger stuff in to this...

	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Run()

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
