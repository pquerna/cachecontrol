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

type Reason int

const (
	ReasonRequestMethodPOST Reason = iota
	ReasonRequestMethodPUT
	ReasonRequestMethodDELETE
	ReasonRequestMethodCONNECT
	ReasonRequestMethodOPTIONS
	ReasonRequestMethodTRACE
	ReasonRequestMethodUnkown

	ReasonRequestNoStore
	ReasonRequestAuthorizationHeader

	ReasonResponseNoStore
	ReasonResponsePrivate

	ReasonResponseUncachableByDefault
)

type Options struct {
	// Set to True for a prviate cache, which is not shared amoung users (eg, in a browser)
	// Set to False for a "shared" cache, which is more common in a server context.
	PrivateCache bool
}

func calcReasons(req *http.Request,
	statusCode int,
	respHeaders http.Header, opts Options) ([]Reason, time.Time, error) {
	var rv []Reason
	var expiresTime time.Time

	respDir, err := ParseResponseCacheControl(respHeaders.Get("Cache-Control"))
	if err != nil {
		return nil, expiresTime, err
	}

	if req != nil {
		reqDir, err := ParseRequestCacheControl(req.Header.Get("Cache-Control"))
		if err != nil {
			return nil, expiresTime, err
		}

		switch req.Method {
		case "GET":
			break
		case "HEAD":
			break
		case "POST":
			/**
			  POST: http://tools.ietf.org/html/rfc7231#section-4.3.3

			  Responses to POST requests are only cacheable when they include
			  explicit freshness information (see Section 4.2.1 of [RFC7234]).
			  However, POST caching is not widely implemented.  For cases where an
			  origin server wishes the client to be able to cache the result of a
			  POST in a way that can be reused by a later GET, the origin server
			  MAY send a 200 (OK) response containing the result and a
			  Content-Location header field that has the same value as the POST's
			  effective request URI (Section 3.1.4.2).
			*/
			if !hasFreshness(reqDir, respDir, respHeaders, opts) {
				rv = append(rv, ReasonRequestMethodPOST)
			}

		case "PUT":
			rv = append(rv, ReasonRequestMethodPUT)

		case "DELETE":
			rv = append(rv, ReasonRequestMethodDELETE)

		case "CONNECT":
			rv = append(rv, ReasonRequestMethodCONNECT)

		case "OPTIONS":
			rv = append(rv, ReasonRequestMethodOPTIONS)

		case "TRACE":
			rv = append(rv, ReasonRequestMethodTRACE)

		// HTTP Extension Methods: http://www.iana.org/assignments/http-methods/http-methods.xhtml
		//
		// To my knowledge, none of them are cachable. Please open a ticket if this is not the case!
		//
		default:
			rv = append(rv, ReasonRequestMethodUnkown)
		}

		if reqDir.NoStore {
			rv = append(rv, ReasonRequestNoStore)
		}

		// Storing Responses to Authenticated Requests: http://tools.ietf.org/html/rfc7234#section-3.2
		authz := req.Header.Get("Authorization")
		if authz != "" {
			if respDir.MustRevalidate || respDir.Public || respDir.SMaxAge != -1 {
				// Expires of some kind present, this is potentially OK.
			} else {
				rv = append(rv, ReasonRequestAuthorizationHeader)
			}
		}
	}

	if respDir.PrivatePresent && !opts.PrivateCache {
		rv = append(rv, ReasonResponsePrivate)
	}

	if respDir.NoStore {
		rv = append(rv, ReasonResponseNoStore)
	}

	/*
	   the response either:

	         *  contains an Expires header field (see Section 5.3), or

	         *  contains a max-age response directive (see Section 5.2.2.8), or

	         *  contains a s-maxage response directive (see Section 5.2.2.9)
	            and the cache is shared, or

	         *  contains a Cache Control Extension (see Section 5.2.3) that
	            allows it to be cached, or

	         *  has a status code that is defined as cacheable by default (see
	            Section 4.2.2), or

	         *  contains a public response directive (see Section 5.2.2.5).
	*/

	expires := respHeaders.Get("Expires") != ""
	statusCachable := cachableStatusCode(statusCode)

	if expires ||
		respDir.MaxAge != -1 ||
		(respDir.SMaxAge != -1 && !opts.PrivateCache) ||
		statusCachable ||
		respDir.Public {
		/* cachable by default, at least one of the above conditions was true */
	} else {
		rv = append(rv, ReasonResponseUncachableByDefault)
	}

	/**
	 * Okay, lets calculate Freshness/Expiration now. woo:
	 *  http://tools.ietf.org/html/rfc7234#section-4.2
	 */

	/*
	   o  If the cache is shared and the s-maxage response directive
	      (Section 5.2.2.9) is present, use its value, or

	   o  If the max-age response directive (Section 5.2.2.8) is present,
	      use its value, or

	   o  If the Expires response header field (Section 5.3) is present, use
	      its value minus the value of the Date response header field, or

	   o  Otherwise, no explicit expiration time is present in the response.
	      A heuristic freshness lifetime might be applicable; see
	      Section 4.2.2.
	*/

	if respDir.SMaxAge != -1 && !opts.PrivateCache {
		expiresTime = time.Now().UTC().Add(time.Second * time.Duration(respDir.SMaxAge))
	} else if respDir.MaxAge != -1 {
		expiresTime = time.Now().UTC().Add(time.Second * time.Duration(respDir.MaxAge))
	} else if expires {
		expiresIn, err := http.ParseTime(respHeaders.Get("Expires"))
		if err != err {
			return nil, expiresTime, err
		}
		serverDate, err := http.ParseTime(respHeaders.Get("Date"))

		// We ignore any errors from this, and use our own time by default,
		// most likely this is a ResponseWriter with the Date header not set yet
		// :(
		if err != nil {
			serverDate = time.Now()
		}

		serverDate = serverDate.UTC()
		expiresIn = expiresIn.UTC()
		expiresTime = time.Now().UTC().Add(serverDate.Sub(expiresIn))
	} else {
		// heuristic freshness lifetime
	}

	return rv, expiresTime, nil
}

// Given an HTTP Request, the future Status Code, and an ResponseWriter,
// determine the possible reasons a response SHOULD NOT be cached.
func CachableResponse(req *http.Request,
	statusCode int,
	resp http.ResponseWriter,
	opts Options) ([]Reason, time.Time, error) {
	return calcReasons(req, statusCode, resp.Header(), opts)
}

// Given an HTTP Request and Response, determine the possible reasons a response SHOULD NOT
// be cached.
func Cachable(req *http.Request,
	resp *http.Response,
	opts Options) ([]Reason, time.Time, error) {
	return calcReasons(req, resp.StatusCode, resp.Header, opts)
}

// calculate if a freshness directive is present: http://tools.ietf.org/html/rfc7234#section-4.2.1
func hasFreshness(reqDir *RequestCacheDirectives, respDir *ResponseCacheDirectives, respHeaders http.Header, opts Options) bool {
	if !opts.PrivateCache && respDir.SMaxAge != -1 {
		return true
	}

	if respDir.MaxAge != -1 {
		return true
	}

	if respHeaders.Get("Expires") != "" {
		return true
	}

	return false
}

func cachableStatusCode(statusCode int) bool {
	/*
		Responses with status codes that are defined as cacheable by default
		(e.g., 200, 203, 204, 206, 300, 301, 404, 405, 410, 414, and 501 in
		this specification) can be reused by a cache with heuristic
		expiration unless otherwise indicated by the method definition or
		explicit cache controls [RFC7234]; all other status codes are not
		cacheable by default.
	*/
	switch statusCode {
	case 200:
		return true
	case 203:
		return true
	case 204:
		return true
	case 206:
		return true
	case 300:
		return true
	case 301:
		return true
	case 404:
		return true
	case 405:
		return true
	case 410:
		return true
	case 414:
		return true
	case 501:
		return true
	default:
		return false
	}
}
