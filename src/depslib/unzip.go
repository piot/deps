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
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/piot/log-go/src/clog"
)

func writeFileOrDirectory(destinationDirectory string, zipEntry *zip.File, ignorePrefix string, log *clog.Log) error {
	extractingReader, err := zipEntry.Open()
	if err != nil {
		return err
	}

	defer func() {
		if err := extractingReader.Close(); err != nil {
			panic(err)
		}
	}()

	relativeName := strings.TrimPrefix(zipEntry.Name, ignorePrefix)
	targetPath := filepath.Join(destinationDirectory, relativeName)

	if zipEntry.FileInfo().IsDir() {
		os.MkdirAll(targetPath, zipEntry.Mode())
	} else {
		log.Trace("extracting file", clog.String("source", zipEntry.Name), clog.String("target", targetPath))
		os.MkdirAll(filepath.Dir(targetPath), zipEntry.Mode())
		targetFileWriter, createFileErr := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipEntry.Mode())
		if createFileErr != nil {
			return createFileErr
		}
		defer func() {
			if err := targetFileWriter.Close(); err != nil {
				panic(err)
			}
		}()

		_, copyErr := io.Copy(targetFileWriter, extractingReader)
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}

func unzipFile(zipFile string, destinationDirectory string, ignorePrefix string, log *clog.Log) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := zipReader.Close(); err != nil {
			panic(err)
		}
	}()

	for _, zipEntry := range zipReader.File {
		err := writeFileOrDirectory(destinationDirectory, zipEntry, ignorePrefix, log)
		if err != nil {
			return err
		}
	}

	return nil
}
