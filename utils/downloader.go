// Copyright (C) 2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package utils

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// This is a generic interface for performing highly testable downloads. All methods here involve
// external http requests. To write tests using these functions, provide a mocked version of this
// interface to your application object.
type Downloader interface {
	Download(url string, token string) ([]byte, error)
}

type downloader struct{}

func NewDownloader() Downloader {
	return &downloader{}
}

func (d downloader) Download(url string, token string) ([]byte, error) {
	body, err := d.doAPIRequest(url, token)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	bs, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failure downloading %s: %w", url, err)
	}
	return bs, nil
}

func (downloader) doAPIRequest(url, token string) (io.ReadCloser, error) {
	retries := 0
	for {
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create http request to download %s: %w", url, err)
		}
		if token != "" {
			// avoid rate limitation issues at CI
			request.Header.Set("authorization", fmt.Sprintf("Bearer %s", token))
		}
		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			return nil, fmt.Errorf("failure downloading %s: %w", url, err)
		}
		if resp.StatusCode != http.StatusOK {
			// http.StatusForbidden is also obtained when hitting github API rate limits
			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusForbidden {
				if retries <= 5 {
					retries++
					toSleep := time.Duration(retries) * 10 * time.Second
					time.Sleep(toSleep)
					continue
				}
			}
			return nil, fmt.Errorf("failure downloading %s: unexpected http status code: %d", url, resp.StatusCode)
		}
		return resp.Body, nil
	}
}
