package depslib

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/piot/log-go/src/clog"
)

func TempDirectory(tempSuffix string) (string, error) {
	dir, err := ioutil.TempDir("", tempSuffix)
	if err != nil {
		return "", err
	}
	return dir, err
}

func CleanDirectory(directory string, log *clog.Log) error {
	os.RemoveAll(directory)
	mkdirErr := os.MkdirAll(directory, os.ModePerm)
	return mkdirErr
}

func CleanTempDirectoryEx(directory string, tempSuffix string, log *clog.Log) (string, error) {
	dir, err := ioutil.TempDir(directory, tempSuffix)
	if err != nil {
		return "", err
	}
	cleanErr := CleanDirectory(dir, log)
	if cleanErr != nil {
		return "", cleanErr
	}
	return dir, err
}

func CleanTempDirectory(tempSuffix string, log *clog.Log) (string, error) {
	dir, err := TempDirectory(tempSuffix)
	if err != nil {
		return "", err
	}
	cleanErr := CleanDirectory(dir, log)
	if cleanErr != nil {
		return "", cleanErr
	}
	return dir, err
}

func CleanDirectoryWithBackup(directory string, tempSuffix string, log *clog.Log) (string, error) {
	backupDir, tempErr := CleanTempDirectoryEx(filepath.Dir(directory), tempSuffix, log)
	if tempErr != nil {
		return "", tempErr
	}
	log.Debug("clean directory", clog.String("directory", directory), clog.String("backupDir", backupDir))
	backupSubDir := filepath.Join(backupDir, tempSuffix)

	renameErr := os.Rename(directory, backupSubDir)
	if renameErr != nil {
		return "", renameErr
	}
	return backupDir, nil
}
