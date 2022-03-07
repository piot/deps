/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package depslib

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func isSymlink(filename string) (bool, error) {
	stat, statErr := os.Lstat(filename)
	if statErr != nil {
		return false, statErr
	}
	return (stat.Mode() & os.ModeSymlink) == os.ModeSymlink, nil
}

func removeSymlinkIfExists(filename string) error {
	symlinkExists, symlinkExistsErr := isSymlink(filename)
	if os.IsNotExist(symlinkExistsErr) {
		return nil
	}
	if symlinkExistsErr != nil {
		return symlinkExistsErr
	}

	if symlinkExists {
		return os.Remove(filename)
	}

	return nil
}

func CreateDirectoryIfNeeded(directory string) error {
	stat, checkDirectoryErr := os.Lstat(directory)
	if checkDirectoryErr != nil || !stat.IsDir() {
		return os.MkdirAll(directory, 0755)
	}
	return nil
}

func MakeSymlink(existingDirectory string, targetDirectory string) error {
	removeSymlinkErr := removeSymlinkIfExists(targetDirectory)
	if removeSymlinkErr != nil {
		fmt.Printf("couldn't remove symlink\n")
		return removeSymlinkErr
	}

	createDirectoryErr := CreateDirectoryIfNeeded(filepath.Dir(targetDirectory))
	if createDirectoryErr != nil {
		fmt.Printf("directory existed\n")
		return createDirectoryErr
	}

	log.Printf("symlinking %v to %v\n", existingDirectory, targetDirectory)

	os.Symlink(existingDirectory, targetDirectory)
	//cmd := exec.Command("ln", "-s", "-r", existingDirectory, targetDirectory)

	//cmd.Dir = parentTargetDirectory

	//cmd.Start()

	//cmd.Wait()

	/*
		err := os.Symlink(existingDirectory, targetDirectory)
		if err != nil {
			log.Printf("got error from OS:%v\n", err)
		}
	*/

	return nil
}
