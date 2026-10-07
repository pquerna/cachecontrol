/**
 *  Copyright 2015 Paul Querna
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package cachecontrol

import (
	"github.com/pquerna/cachecontrol/cacheobject"
	"github.com/stretchr/testify/require"

	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func roundTrip(t *testing.T, fnc func(w http.ResponseWriter, r *http.Request)) (*http.Request, *http.Response) {
	ts := httptest.NewServer(http.HandlerFunc(fnc))
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	_, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	require.NoError(t, err)
	return req, res
}

func TestCachableResponsePublic(t *testing.T) {
	req, res := roundTrip(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public")
		w.Header().Set("Last-Modified",
			time.Now().UTC().Add(time.Duration(time.Hour*-5)).Format(http.TimeFormat))
		fmt.Fprintln(w, `{}`)
	})

	opts := Options{}
	reasons, expires, err := CachableResponse(req, res, opts)
	require.NoError(t, err)
	require.Len(t, reasons, 0)
	require.WithinDuration(t,
		time.Now().UTC().Add(time.Duration(float64(time.Hour)*0.5)),
		expires,
		10*time.Second)
}

func TestCachableResponsePrivate(t *testing.T) {
	req, res := roundTrip(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "private")
		fmt.Fprintln(w, `{}`)
	})

	opts := Options{}
	reasons, expires, err := CachableResponse(req, res, opts)
	require.NoError(t, err)
	require.Len(t, reasons, 1)
	require.Equal(t, reasons[0], cacheobject.ReasonResponsePrivate)
	require.Equal(t, time.Time{}, expires)

	opts.PrivateCache = true
	reasons, expires, err = CachableResponse(req, res, opts)
	require.NoError(t, err)
	require.Len(t, reasons, 0)
	require.Equal(t, time.Time{}, expires)
}

func TestResponseWriter(t *testing.T) {
	var resp http.ResponseWriter
	var req *http.Request
	_, _ = roundTrip(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "private")
		fmt.Fprintln(w, `{}`)
		resp = w
		req = r
	})

	opts := Options{}
	reasons, expires, err := CachableResponseWriter(req, 200, resp, opts)
	require.NoError(t, err)
	require.Len(t, reasons, 1)
	require.Equal(t, reasons[0], cacheobject.ReasonResponsePrivate)
	require.Equal(t, time.Time{}, expires)

	opts.PrivateCache = true
	reasons, expires, err = CachableResponseWriter(req, 200, resp, opts)
	require.NoError(t, err)
	require.Len(t, reasons, 0)
	require.Equal(t, time.Time{}, expires)
}

func TestRepeatedCacheControl(t *testing.T) {
	tests := []struct {
		name         string
		request      []string
		response     []string
		status       int
		privateCache bool
		reasons      []cacheobject.Reason
		err          error
		freshness    time.Duration
	}{
		{name: "request no-store", request: []string{"max-age=60", "no-store"}, reasons: []cacheobject.Reason{cacheobject.ReasonRequestNoStore}},
		{name: "response no-store", response: []string{"max-age=60", "no-store"}, reasons: []cacheobject.Reason{cacheobject.ReasonResponseNoStore}},
		{name: "response private", response: []string{"max-age=60", "private"}, reasons: []cacheobject.Reason{cacheobject.ReasonResponsePrivate}},
		{name: "private cache", response: []string{"max-age=60", "private"}, privateCache: true, freshness: time.Minute},
		{name: "empty first request", request: []string{"", "no-store"}, reasons: []cacheobject.Reason{cacheobject.ReasonRequestNoStore}},
		{name: "empty first response", response: []string{"", "no-store"}, reasons: []cacheobject.Reason{cacheobject.ReasonResponseNoStore}},
		{name: "unquoted field names", response: []string{"private=Set-Cookie", "no-store"}, reasons: []cacheobject.Reason{cacheobject.ReasonResponsePrivate, cacheobject.ReasonResponseNoStore}},
		{name: "quoted field names", response: []string{`private="Set-Cookie, X-Private"`, "no-store"}, reasons: []cacheobject.Reason{cacheobject.ReasonResponsePrivate, cacheobject.ReasonResponseNoStore}},
		{name: "invalid later request", request: []string{"no-transform", "min-fresh=invalid"}, err: cacheobject.ErrMinFreshDeltaSeconds},
		{name: "invalid later response", response: []string{"public", "max-age"}, err: cacheobject.ErrMaxAgeDeltaSeconds},
		{name: "later public", response: []string{"no-transform", "public"}, status: http.StatusInternalServerError},
		{name: "shared freshness", response: []string{"max-age=60", "s-maxage=120"}, freshness: 2 * time.Minute},
		{name: "private freshness", response: []string{"max-age=60", "s-maxage=120"}, privateCache: true, freshness: time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://example.com/", nil)
			require.NoError(t, err)
			for _, value := range tt.request {
				req.Header.Add("Cache-Control", value)
			}
			writer := httptest.NewRecorder()
			for _, value := range tt.response {
				writer.Header().Add("Cache-Control", value)
			}
			status := tt.status
			if status == 0 {
				status = http.StatusOK
			}
			resp := &http.Response{StatusCode: status, Header: writer.Header()}
			requestHeaders, responseHeaders := req.Header.Clone(), resp.Header.Clone()
			opts := Options{PrivateCache: tt.privateCache}
			for _, api := range []string{"response", "writer"} {
				t.Run(api, func(t *testing.T) {
					var reasons []cacheobject.Reason
					var expires time.Time
					now := time.Now().UTC()
					if api == "response" {
						reasons, expires, err = CachableResponse(req, resp, opts)
					} else {
						reasons, expires, err = CachableResponseWriter(req, status, writer, opts)
					}
					require.Equal(t, tt.err, err)
					require.Equal(t, tt.reasons, reasons)
					if tt.freshness > 0 {
						require.WithinDuration(t, now.Add(tt.freshness), expires, time.Second)
					}
					require.Equal(t, requestHeaders, req.Header)
					require.Equal(t, responseHeaders, resp.Header)
				})
			}
		})
	}
}
