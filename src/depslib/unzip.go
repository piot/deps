/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package depslib

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func writeFileOrDirectory(destinationDirectory string, zipEntry *zip.File, ignorePrefix string) error {
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

func unzipFile(zipFile string, destinationDirectory string, ignorePrefix string) error {
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
		err := writeFileOrDirectory(destinationDirectory, zipEntry, ignorePrefix)
		if err != nil {
			return err
		}
	}

	return nil
}
