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

package cacheobject

import (
	"github.com/stretchr/testify/require"

	"testing"
)

func TestCachableStatusCode(t *testing.T) {
	ok := []int{200, 203, 204, 206, 300, 301, 404, 405, 410, 414, 501}
	for _, v := range ok {
		require.True(t, cachableStatusCode(v), "status code should be cacheable: %d", v)
	}

	notok := []int{201, 429, 500, 504}
	for _, v := range notok {
		require.False(t, cachableStatusCode(v), "status code should not be cachable: %d", v)
	}
}
