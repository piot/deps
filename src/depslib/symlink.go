/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package depslib

import (
	"fmt"
	"os"
	"os/exec"
	"path"
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
		return os.MkdirAll(directory, os.ModePerm)
	}
	return nil
}

func MakeSymlink(existingFilename string, symlinkFilename string) error {
	/*
		removeSymlinkErr := removeSymlinkIfExists(symlinkFilename)
		if removeSymlinkErr != nil {
			fmt.Printf("couldn't remove symlink\n")
			return removeSymlinkErr
		}

		createDirectoryErr := CreateDirectoryIfNeeded(filepath.Dir(symlinkFilename))
		if createDirectoryErr != nil {
			fmt.Printf("directory existed\n")
			return createDirectoryErr
		}
	*/

	cmd := exec.Command("ln", "-s", existingFilename, symlinkFilename)

	//cmd.Dir = depsPath

	cmd.Start()

	cmd.Wait()

	return nil // os.Symlink(existingFilename, symlinkFilename)
}

func MakeRelativeSymlink(existingFilename string, symlinkFilename string, rootDirectory string) error {
	relativePath, err := filepath.Rel(path.Dir(symlinkFilename), existingFilename)
	if err != nil {
		return fmt.Errorf("not a relative path %v %v %v", existingFilename, symlinkFilename, err)
	}
	fmt.Printf("relative symlink '%v' to '%v'\n", relativePath, symlinkFilename)

	return MakeSymlink(relativePath, symlinkFilename)
}
