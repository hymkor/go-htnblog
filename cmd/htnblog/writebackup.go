package main

import (
	"os"
)

func writeFileWithBackup(fname string, data []byte, perm os.FileMode) error {
	tmpName := fname + ".tmp"
	if err := os.WriteFile(tmpName, data, perm); err != nil {
		return err
	}
	os.Rename(fname, fname+"~")
	return os.Rename(tmpName, fname)
}
