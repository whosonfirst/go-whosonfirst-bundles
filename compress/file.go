package compress

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func CompressFile(source string, chroot string, opts *CompressOptions) (string, error) {

	dest, err := CompressedFilePath(source, chroot)

	tar := "bzip2"

	args := []string{
		"-c", // send output to stdout which we capture below
		"-k", // Keep (don't delete) input files during compression or decompression.
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

	fh, err := os.Create(dest)

	if err != nil {
		return "", err
	}

	defer fh.Close()

	reader := bytes.NewReader(out.Bytes())
	_, err = io.Copy(fh, reader)

	if err != nil {
		return "", err
	}

	if opts.RemoveSource {

		err = os.Remove(source)

		if err != nil {
			return "", err
		}
	}

	return dest, nil
}

func CompressedFilePath(source string, chroot string) (string, error) {

	abs_source, err := filepath.Abs(source)

	if err != nil {
		return "", err
	}

	abs_chroot, err := filepath.Abs(chroot)

	if err != nil {
		return "", err
	}

	fname := filepath.Base(abs_source)
	fname = fmt.Sprintf("%s.bz2", fname)

	dest := filepath.Join(abs_chroot, fname)
	return dest, nil
}
