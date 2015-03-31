# cachecontrol: HTTP Caching Parser and Interpretation

[![GoDoc][1]][2][![Build Status](https://travis-ci.org/pquerna/cachecontrol.svg?branch=master)](https://travis-ci.org/pquerna/cachecontrol)
[1]: https://godoc.org/github.com/pquerna/cachecontrol?status.svg
[2]: https://godoc.org/github.com/pquerna/cachecontrol
 

`cachecontrol` implements [RFC 7234](http://tools.ietf.org/html/rfc7234) __Hypertext Transfer Protocol (HTTP/1.1): Caching__.  It does this by parsing the `Cache-Control` and other headers, providing information about requests and responses -- but `cachecontrol` does not implement an actual cache backend, just the control plane to make decisions about if a particular response is cachable.

# License

[Apache 2.0](./LICENSE)
