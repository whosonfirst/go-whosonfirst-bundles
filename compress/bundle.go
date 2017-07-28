package compress

import (
	"bytes"
	"fmt"
	_ "log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CompressBundle(source string, chroot string, opts *CompressOptions) (string, error) {

	dest, err := CompressedBundlePath(source)

	if err != nil {
		return "", nil
	}

	abs_chroot, err := filepath.Abs(chroot)

	if err != nil {
		return "", err
	}

	// this bit is important - we are going to tar -C chroot
	// so we don't want to pass an absolute path (to tar)

	rel_source := strings.Replace(source, abs_chroot, "", 1)

	if strings.HasPrefix(rel_source, "/") {
		rel_source = strings.Replace(rel_source, "/", "", 1)
	}

	tar := "tar"

	args := []string{
		"-C", abs_chroot, // -C is for chroot
		"-cjf", // -c is for create; -j is for bzip; -f if for file
		dest,
		rel_source,
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

func CompressedBundlePath(source string) (string, error) {

	abs_source, err := filepath.Abs(source)

	if err != nil {
		return "", err
	}

	dest := fmt.Sprintf("%s.tar.bz2", abs_source)
	return dest, nil
}
