package main

import (
	"os"
)

func inArray(array []string, target string) bool {
	for _, e := range array {
		if e == target {
			return true
		}
	}
	return false
}

func searchMapKeys(m map[string]int, value int) []string {
	keys := []string{}
	for k, v := range m {
		if v == value {
			keys = append(keys, k)
		}
	}
	return keys
}

func modTime(path string) (int64, error) {
	fs, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return fs.ModTime().UnixNano(), nil
}
