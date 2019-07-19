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
	"path/filepath"

	"github.com/piot/log-go/src/clog"
)

func isSymlink(filename string) (bool, error) {
	stat, statErr := os.Lstat(filename)
	if statErr != nil {
		return false, statErr
	}
	return (stat.Mode() & os.ModeSymlink) == os.ModeSymlink, nil
}

func removeSymlinkIfExists(filename string, log *clog.Log) error {
	symlinkExists, symlinkExistsErr := isSymlink(filename)
	if os.IsNotExist(symlinkExistsErr) {
		return nil
	}
	if symlinkExistsErr != nil {
		return symlinkExistsErr
	}

	if symlinkExists {
		log.Trace("removing symlink", clog.String("filename", filename))
		return os.Remove(filename)
	}

	return nil
}

func CreateDirectoryIfNeeded(directory string, log *clog.Log) error {
	stat, checkDirectoryErr := os.Lstat(directory)
	if checkDirectoryErr != nil || !stat.IsDir() {
		log.Debug("creating directory", clog.String("directory", directory))
		return os.MkdirAll(directory, os.ModePerm)
	}
	return nil
}

func MakeSymlink(existingFilename string, symlinkFilename string, log *clog.Log) error {
	removeSymlinkErr := removeSymlinkIfExists(symlinkFilename, log)
	if removeSymlinkErr != nil {
		return removeSymlinkErr
	}

	createDirectoryErr := CreateDirectoryIfNeeded(filepath.Dir(symlinkFilename), log)
	if createDirectoryErr != nil {
		return createDirectoryErr
	}

	log.Debug("make symlink", clog.String("from", existingFilename), clog.String("to", symlinkFilename))
	return os.Symlink(existingFilename, symlinkFilename)
}

func MakeRelativeSymlink(existingFilename string, symlinkFilename string, log *clog.Log) error {
	relativePath, err := filepath.Rel(filepath.Dir(symlinkFilename), existingFilename)
	if err != nil {
		return fmt.Errorf("not a relative path %v %v %v", existingFilename, symlinkFilename, err)
	}
	return MakeSymlink(relativePath, symlinkFilename, log)
}
