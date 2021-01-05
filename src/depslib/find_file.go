/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package depslib

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

func find(startPath string) ([]string, error) {
	_, wdErr := os.Getwd()
	if wdErr != nil {
		return nil, wdErr
	}

	directory, directoryErr := filepath.Abs(startPath)
	if directoryErr != nil {
		return nil, directoryErr
	}
	var foundRoots []string

	for {
		foundDirectory, foundDirectoryErr := os.Lstat(directory)
		if foundDirectoryErr != nil || !foundDirectory.IsDir() {
			return foundRoots, nil
		}
		configurationFilename := path.Join(directory, "deps.toml")
		foundConfiguration, foundConfigurationErr := os.Lstat(configurationFilename)
		if foundConfigurationErr == nil && !foundConfiguration.IsDir() {
			foundRoots = append(foundRoots, configurationFilename)
		}

		foundGitDir, foundGitErr := os.Lstat(path.Join(directory, ".git"))
		if foundGitErr == nil && foundGitDir.IsDir() {
			return foundRoots, nil
		}
		if directory == "." || directory == "/" {
			return foundRoots, nil
		}
		directory = filepath.Dir(directory)
	}

}

func FindClosestConfigurationFiles(startPath string) ([]string, error) {
	roots, rootsErr := find(startPath)
	if len(roots) == 0 {
		return nil, fmt.Errorf("no deps.toml file found")
	}
	return roots, rootsErr
}
