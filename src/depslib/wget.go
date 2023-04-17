/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package depslib

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

func HTTPGet(downloadURL *url.URL, targetFile string) (err error) {
	resp, err := http.Get(downloadURL.String())
	if err != nil {
		return
	}
	log.Printf("downloading %v", downloadURL)
	defer resp.Body.Close()

	out, createErr := os.Create(targetFile)
	if createErr != nil {
		return createErr
	}
	log.Printf("targetFile %v", targetFile)

	defer out.Close()

	_, copyErr := io.Copy(out, resp.Body)
	if copyErr != nil {
		log.Printf("couldnt copy %v", copyErr)
	}

	return copyErr
}
