package fileutil

import (
	"log"
	"os"
	"path/filepath"
)

func targetPath(dir string, namebase string, ext string) string {
	targetpath := filepath.Join(dir, namebase)
	if ext != "" {
		targetpath += "." + ext
	}
	return targetpath
}
func EnsureFile(dir string, namebase string, ext string,
	dirmode os.FileMode) string {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, dirmode)
	}
	targetpath := targetPath(dir, namebase, ext)
	if _, err := os.Stat(targetpath); os.IsNotExist(err) {
		f, err := os.Create(targetpath)
		if err != nil {
			log.Fatal("Can't create file " + targetpath)
		}
		f.Close()
	}

	return targetpath
}
func DeleteFile(dir string, namebase string, ext string) error {
	targetpath := targetPath(dir, namebase, ext)
	return os.Remove(targetpath)
}
