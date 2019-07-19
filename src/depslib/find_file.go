/*

MIT License

Copyright (c) 2019 Peter Bjorklund

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

package depslib

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/piot/log-go/src/clog"
)

func find(startPath string, log *clog.Log) ([]string, error) {
	wd, wdErr := os.Getwd()
	if wdErr != nil {
		log.Err(wdErr)
		return nil, wdErr
	}
	log.Info("getwd", clog.String("wd", wd))

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

func FindClosestConfigurationFiles(startPath string, log *clog.Log) ([]string, error) {
	roots, rootsErr := find(startPath, log)
	if len(roots) == 0 {
		return nil, fmt.Errorf("no deps.toml file found")
	}
	return roots, rootsErr
}
