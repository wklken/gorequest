# GoRequest

GoRequest is a chainable Go HTTP client inspired by SuperAgent for Node.js.
It wraps the standard `net/http` package with a compact request builder for
common client-side HTTP work.

This repository is maintained as `github.com/wklken/gorequest`. It started as a
fork of [parnurzeal/gorequest](https://github.com/parnurzeal/gorequest), but is
now maintained independently with bug fixes, compatibility updates, and new
features.

## Features

- HTTP verbs: `GET`, `POST`, `PUT`, `HEAD`, `DELETE`, `PATCH`, `OPTIONS`, and custom methods
- Query building from strings, maps, structs, and explicit key/value params
- JSON, form, XML, text, and raw byte request bodies
- Multipart form data and file uploads
- Header replacement and repeated header appends
- Cookies, cookie jars, and basic authentication
- Proxy support, including environment proxy settings and SOCKS5 proxies
- Request timeout, granular timeout, TLS, redirect, compression, and context controls
- Retry support for selected HTTP status codes and custom retry policies
- Request body upload progress callbacks
- String, byte slice, and JSON-decoded response helpers
- Request/response debug logging, curl command output, HTTP tracing, and gock-based mocks

## Installation

```sh
go get github.com/wklken/gorequest
```

The module declares Go 1.21 and is tested in CI across Go 1.21.x through
1.26.x.

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/wklken/gorequest"
)

func main() {
	resp, body, errs := gorequest.New().
		Get("https://example.com").
		End()
	if len(errs) > 0 {
		panic(errs[0])
	}

	fmt.Println(resp.StatusCode)
	fmt.Println(body)
}
```

Most methods return `*SuperAgent`, so request setup can be chained. `End`
returns the response, response body as a string, and a slice of errors.

## Usage

Detailed usage examples live in [docs/usage.md](docs/usage.md).

API reference is available at
[pkg.go.dev/github.com/wklken/gorequest](https://pkg.go.dev/github.com/wklken/gorequest).

## Development

This is a single-package Go library. The usual local checks are:

```sh
go test ./...
go test -v ./...
make test
```

The [Makefile](Makefile) captures the standard local commands.

## Lineage and Credits

GoRequest was originally created by
[parnurzeal/gorequest](https://github.com/parnurzeal/gorequest).

The GoRequest gopher image used by the original project was credited to Wisi
Mongkhonsrisawat, and the Go gopher mascot was created by Renee French.

## License

GoRequest is released under the [MIT License](LICENSE).
