package compress

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

	tar := "tar"

	args := []string{
		"-C", abs_chroot, // -C is for chroot
		"-cjf", // -c is for create; -j is for bzip; -f if for file
		dest,
		source,
	}

	log.Println(tar, args)

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
