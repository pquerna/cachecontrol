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
	"net/http"
	"time"
)

type Options struct {
	// Set to True for a prviate cache, which is not shared amoung users (eg, in a browser)
	// Set to False for a "shared" cache, which is more common in a server context.
	PrivateCache bool
}

// Given an HTTP Request, the future Status Code, and an ResponseWriter,
// determine the possible reasons a response SHOULD NOT be cached.
func CachableResponse(req *http.Request,
	statusCode int,
	resp http.ResponseWriter,
	opts Options) ([]Reason, time.Time, error) {
	return usingRequestResponse(req, statusCode, resp.Header(), opts)
}

// Given an HTTP Request and Response, determine the possible reasons a response SHOULD NOT
// be cached.
func Cachable(req *http.Request,
	resp *http.Response,
	opts Options) ([]Reason, time.Time, error) {
	return usingRequestResponse(req, resp.StatusCode, resp.Header, opts)
}