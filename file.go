package main

import (
	"os"
)

func openMakefile(path string) (*os.File, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return fd, nil
}

func closeMakefile(fd *os.File) {
	fd.Close()
}

func makefilePath(file string) (path string, err error) {
	if file != "" {
		path = file
	} else {
		path = "Makefile"
	}

	_, err = os.Stat(path)
	return
}
