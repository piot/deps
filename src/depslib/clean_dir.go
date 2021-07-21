/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package depslib

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func TempDirectory(tempSuffix string) (string, error) {
	dir, err := ioutil.TempDir("", tempSuffix)
	if err != nil {
		return "", err
	}
	return dir, err
}

func BackupDeps(depsPath string) error {
	log.Println("force clean deps")
	_, statErr := os.Stat(depsPath)
	if statErr == nil {
		_, cleanErr := CleanDirectoryWithBackup(depsPath, "deps.clean")
		if cleanErr != nil {
			return cleanErr
		}
	} else {
		log.Printf("deps path do not exists %v\n", depsPath)
		return nil
	}

	return nil
}

func CleanDirectory(directory string) error {
	os.RemoveAll(directory)
	mkdirErr := os.MkdirAll(directory, 0755)
	return mkdirErr
}

func CleanTempDirectoryEx(directory string, tempSuffix string) (string, error) {
	dir, err := ioutil.TempDir(directory, tempSuffix)
	if err != nil {
		return "", err
	}
	cleanErr := CleanDirectory(dir)
	if cleanErr != nil {
		return "", cleanErr
	}
	return dir, err
}

func CleanTempDirectory(tempSuffix string) (string, error) {
	dir, err := TempDirectory(tempSuffix)
	if err != nil {
		return "", err
	}
	cleanErr := CleanDirectory(dir)
	if cleanErr != nil {
		return "", cleanErr
	}
	return dir, err
}

func CleanDirectoryWithBackup(directory string, tempSuffix string) (string, error) {
	stat, statErr := os.Stat(directory)
	_, isPathError := statErr.(*os.PathError)
	if statErr != nil && !isPathError {
		return "", statErr
	}
	existingDir := !isPathError && stat.IsDir()
	if existingDir {
		log.Println("trying to clean")
		backupDir, tempErr := CleanTempDirectoryEx(filepath.Dir(directory), tempSuffix)
		if tempErr != nil {
			return "", tempErr
		}
		backupSubDir := filepath.Join(backupDir, tempSuffix)

		renameErr := os.Rename(directory, backupSubDir)
		if renameErr != nil {
			return "", renameErr
		}
		return backupDir, nil
	}
	log.Println("mkdir all")
	return "", os.MkdirAll(directory, 0755)
}
