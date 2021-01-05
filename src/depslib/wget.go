/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Peter Bjorklund. All rights reserved.
 *  Licensed under the MIT License. See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package depslib

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func HTTPGet(downloadURL *url.URL) (content io.Reader, err error) {
	request, err := http.NewRequest("GET", downloadURL.String(), nil)
	if err != nil {
		return
	}
	timeout := time.Second * 10
	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	request = request.WithContext(ctx)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		cancelFunc()
		return nil, fmt.Errorf("INVALID RESPONSE; status: %s", response.Status)
	}

	return response.Body, nil
}
