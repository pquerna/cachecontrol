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
)

type Options struct {
	// Set to True for a prviate cache, which is not shared amoung users (eg, in a browser)
	// Set to False for a "shared" cache, which is more common in a server context.
	PrivateCache bool
}

// calculate if a freshness directive is present: http://tools.ietf.org/html/rfc7234#section-4.2.1
func hasFreshness(reqDir *RequestCacheDirectives, respDir *ResponseCacheDirectives, resp *http.Response, opts Options) bool {
	if !opts.PrivateCache && respDir.SMaxAge != -1 {
		return true
	}

	if respDir.MaxAge != -1 {
		return true
	}

	if resp.Header.Get("Expires") != "" {
		return true
	}

	return false
}

func Cachable(req *http.Request, resp *http.Response, opts Options) ([]Reason, error) {
	var rv []Reason

	respDir, err := ParseResponseCacheControl(resp.Header.Get("Cache-Control"))
	if err != nil {
		return nil, err
	}

	if req != nil {
		reqDir, err := ParseRequestCacheControl(req.Header.Get("Cache-Control"))
		if err != nil {
			return nil, err
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
			if !hasFreshness(reqDir, respDir, resp, opts) {
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

	// TODO(pquerna): response status code & the response either clauses.

	return rv, nil
}
