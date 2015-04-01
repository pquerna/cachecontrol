# cachecontrol: HTTP Caching Parser and Interpretation

[![GoDoc][1]][2][![Build Status](https://travis-ci.org/pquerna/cachecontrol.svg?branch=master)](https://travis-ci.org/pquerna/cachecontrol)
[1]: https://godoc.org/github.com/pquerna/cachecontrol?status.svg
[2]: https://godoc.org/github.com/pquerna/cachecontrol
 

`cachecontrol` implements [RFC 7234](http://tools.ietf.org/html/rfc7234) __Hypertext Transfer Protocol (HTTP/1.1): Caching__.  It does this by parsing the `Cache-Control` and other headers, providing information about requests and responses -- but `cachecontrol` does not implement an actual cache backend, just the control plane to make decisions about if a particular response is cachable.

# Usage

## Can you cache Example.com?

```go
package main

import (
	"github.com/pquerna/cachecontrol"

	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	req, _ := http.NewRequest("GET", "http://www.example.com/", nil)

	res, _ := http.DefaultClient.Do(req)
	_, _ = ioutil.ReadAll(res.Body)

	reasons, expires, _ := cachecontrol.CachableResponse(req, res, cachecontrol.Options{})

	fmt.Println("Reasons to not cache: ", reasons)
	fmt.Println("Expiration: ", expires.String())
}
```

# License

[Apache 2.0](./LICENSE)
