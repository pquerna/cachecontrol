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
	"github.com/stretchr/testify/require"

	"testing"
)

func TestMaxAge(t *testing.T) {
	cd, err := ParseResponseCacheControl("")
	require.NoError(t, err)
	require.Equal(t, cd.MaxAge, -1)

	cd, err = ParseResponseCacheControl("max-age")
	require.Error(t, err)

	cd, err = ParseResponseCacheControl("max-age=20")
	require.NoError(t, err)
	require.Equal(t, cd.MaxAge, 20)

	cd, err = ParseResponseCacheControl("max-age=0")
	require.NoError(t, err)
	require.Equal(t, cd.MaxAge, 0)

	cd, err = ParseResponseCacheControl("max-age=-1")
	require.Error(t, err)
}

func TestSMaxAge(t *testing.T) {
	cd, err := ParseResponseCacheControl("")
	require.NoError(t, err)
	require.Equal(t, cd.SMaxAge, -1)

	cd, err = ParseResponseCacheControl("s-maxage")
	require.Error(t, err)

	cd, err = ParseResponseCacheControl("s-maxage=20")
	require.NoError(t, err)
	require.Equal(t, cd.SMaxAge, 20)

	cd, err = ParseResponseCacheControl("s-maxage=0")
	require.NoError(t, err)
	require.Equal(t, cd.SMaxAge, 0)

	cd, err = ParseResponseCacheControl("s-maxage=-1")
	require.Error(t, err)
}
