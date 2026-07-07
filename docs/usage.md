# GoRequest Usage

This guide covers the common request patterns for
`github.com/wklken/gorequest`. For the complete API, see the
[pkg.go.dev reference](https://pkg.go.dev/github.com/wklken/gorequest).

## Basic Requests

```go
request := gorequest.New()
resp, body, errs := request.Get("https://example.com").End()
```

If you do not need to reuse the request builder:

```go
resp, body, errs := gorequest.New().
	Get("https://example.com").
	End()
```

The same pattern works for supported HTTP verbs:

```go
request := gorequest.New()

resp, body, errs := request.Post("https://example.com").End()
// request.Put("https://example.com").End()
// request.Delete("https://example.com").End()
// request.Head("https://example.com").End()
// request.Patch("https://example.com").End()
// request.Options("https://example.com").End()
// request.CustomMethod("TRACE", "https://example.com").End()
```

## Headers

Use `Set` to replace an existing header value:

```go
resp, body, errs := gorequest.New().
	Post("https://example.com/gamelist").
	Set("Accept", "application/json").
	End()
```

Use `AppendHeader` when the request needs repeated values for the same header:

```go
resp, body, errs := gorequest.New().
	Post("https://example.com/gamelist").
	AppendHeader("Accept", "application/json").
	AppendHeader("Accept", "text/plain").
	End()
```

Use `SetHeaders` to set multiple headers from a map or struct:

```go
headers := map[string]string{
	"Accept": "application/json",
	"Notes":  "GoRequest is coming!",
}

resp, body, errs := gorequest.New().
	Get("https://example.com").
	SetHeaders(headers).
	End()
```

## Query Parameters

Query values can be supplied as a raw query string, map, struct, or explicit
params:

```go
resp, body, errs := gorequest.New().
	Get("https://example.com/search").
	Query("q=gorequest&page=1").
	End()
```

```go
resp, body, errs := gorequest.New().
	Get("https://example.com/search").
	Query(map[string]string{"q": "gorequest", "page": "1"}).
	End()
```

```go
resp, body, errs := gorequest.New().
	Get("https://example.com/search").
	Param("q", "gorequest").
	Param("page", "1").
	End()
```

## JSON Bodies

`Send` accepts JSON strings, structs, maps, slices, and byte slices.

```go
resp, body, errs := gorequest.New().
	Post("https://example.com").
	Set("Notes", "gorequest is coming!").
	Send(`{"name":"backy","species":"dog"}`).
	End()
```

Struct values are marshaled into JSON:

```go
type BrowserVersionSupport struct {
	Chrome  string
	Firefox string
}

ver := BrowserVersionSupport{
	Chrome:  "37.0.2041.6",
	Firefox: "30.0",
}

resp, body, errs := gorequest.New().
	Post("https://version.example/update").
	Send(ver).
	Send(`{"Safari":"5.1.10"}`).
	End()
```

When you need explicit helpers, use `SendMap`, `SendStruct`, `SendSlice`,
`SendBytes`, or `SendString`.

By default, JSON request bodies follow `encoding/json` and escape HTML
characters. Use `SetJSONOptions` when the peer expects unescaped HTML
characters in JSON strings:

```go
resp, body, errs := gorequest.New().
	SetJSONOptions(gorequest.JSONOptions{DisableHTMLEscape: true}).
	Post("https://example.com").
	Send(map[string]string{"html": "<div>ok</div>"}).
	End()
```

## Form, Text, and XML Bodies

Use `Type` to choose a request content type:

```go
resp, body, errs := gorequest.New().
	Post("https://example.com/form").
	Type("form").
	Send(map[string]string{"name": "gorequest"}).
	End()
```

```go
resp, body, errs := gorequest.New().
	Post("https://example.com/text").
	Type("text").
	Send("plain text body").
	End()
```

```go
resp, body, errs := gorequest.New().
	Post("https://example.com/xml").
	Type("xml").
	Send("<name>gorequest</name>").
	End()
```

## Multipart Form Data

Use `Type("multipart")` for multipart form submissions:

```go
resp, body, errs := gorequest.New().
	Post("https://example.com/upload").
	Type("multipart").
	Send(`{"query1":"test"}`).
	End()
```

`SendFile` accepts a file path, a byte slice, or an `os.File`. Optional
arguments can set a custom file name and field name.

```go
path, _ := filepath.Abs("./file2.txt")
bytesOfFile, _ := os.ReadFile(path)

resp, body, errs := gorequest.New().
	Post("https://example.com/upload").
	Type("multipart").
	SendFile("./file1.txt").
	SendFile(bytesOfFile, "file2.txt", "my_file_fieldname").
	End()
```

Track request body upload progress with `SetUploadProgress`:

```go
resp, body, errs := gorequest.New().
	Post("https://example.com/upload").
	Type("multipart").
	SendFile("./file1.txt").
	SetUploadProgress(func(uploaded int64) {
		fmt.Println("uploaded bytes:", uploaded)
	}).
	End()
```

## Response Helpers

`End` returns the response body as a string:

```go
resp, body, errs := gorequest.New().
	Get("https://example.com").
	End()
```

`EndBytes` returns the response body as bytes:

```go
resp, bodyBytes, errs := gorequest.New().
	Get("https://example.com/file").
	EndBytes()
```

`EndStruct` decodes a JSON response into a struct:

```go
type HeyYou struct {
	Hey string `json:"hey"`
}

var heyYou HeyYou
resp, bodyBytes, errs := gorequest.New().
	Get("https://example.com").
	EndStruct(&heyYou)
```

All three helpers also accept callback functions.

```go
func printStatus(resp gorequest.Response, body string, errs []error) {
	fmt.Println(resp.Status)
}

gorequest.New().
	Get("https://example.com").
	End(printStatus)
```

## Authentication and Cookies

Add a basic authentication header:

```go
resp, body, errs := gorequest.New().
	SetBasicAuth("username", "password").
	Get("https://example.com/private").
	End()
```

Add explicit cookies:

```go
cookie := &http.Cookie{Name: "session", Value: "abc123"}

resp, body, errs := gorequest.New().
	AddCookie(cookie).
	Get("https://example.com").
	End()
```

## Proxy

Set a proxy URL explicitly:

```go
request := gorequest.New().Proxy("http://proxy:999")

resp, body, errs := request.
	Get("https://example-proxy.com").
	End()
```

Use an empty proxy string to clear the explicit proxy on a reused request:

```go
resp, body, errs = request.
	Proxy("").
	Get("https://example-no-proxy.com").
	End()
```

New requests also use the standard proxy environment variables handled by
`net/http`, including `HTTP_PROXY`, `HTTPS_PROXY`, and `NO_PROXY`.

## Timeouts

Set one timeout for the request:

```go
resp, body, errs := gorequest.New().
	Timeout(2 * time.Second).
	Get("https://example.com").
	End()
```

Set granular timeouts with `Timeouts`:

```go
timeouts := &gorequest.Timeouts{
	Dial:           5 * time.Second,
	TlsHandshake:   2 * time.Second,
	ResponseHeader: 2 * time.Second,
}

resp, body, errs := gorequest.New().
	Timeouts(timeouts).
	Get("https://example.com").
	End()
```

## Redirects

Redirects are controlled with `RedirectPolicy`, which behaves like
`net/http.Client.CheckRedirect`.

```go
resp, body, errs := gorequest.New().
	Get("http://example.com").
	RedirectPolicy(func(req gorequest.Request, via []gorequest.Request) error {
		if req.URL.Scheme != "https" {
			return http.ErrUseLastResponse
		}
		return nil
	}).
	End()
```

Disable redirects entirely:

```go
resp, body, errs := gorequest.New().
	DisableRedirect().
	Get("http://example.com").
	End()
```

## Retry

Retry a request for selected response status codes:

```go
resp, body, errs := gorequest.New().
	Get("https://example.com").
	Retry(3, 5*time.Second, http.StatusBadRequest, http.StatusInternalServerError).
	End()
```

## Clone and Reuse

Reuse request settings by cloning before making a request. Clones copy headers,
query params, cookies, the HTTP client settings, and other request settings.
Clones share the underlying transport by default so connection pooling can be
reused; changing transport settings such as proxy, TLS, compression, or
granular timeouts on a clone gives that clone its own transport.

```go
baseRequest := gorequest.New().
	Timeout(10 * time.Second).
	SetBasicAuth("user", "password")

resp, body, errs := baseRequest.Clone().
	Get("https://example.com").
	End()
```

## Context and Tracing

Attach a context:

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

resp, body, errs := gorequest.New().
	Context(ctx).
	Get("https://example.com").
	End()
```

Attach an HTTP client trace:

```go
trace := &httptrace.ClientTrace{
	GotConn: func(info httptrace.GotConnInfo) {
		fmt.Println("reused:", info.Reused)
	},
}

resp, body, errs := gorequest.New().
	HttpTrace(trace).
	Get("https://example.com").
	End()
```

## Debugging and Curl Output

Dump request and response details with debug logging:

```go
resp, body, errs := gorequest.New().
	SetDebug(true).
	Get("https://example.com").
	End()
```

The `GOREQUEST_DEBUG=1` environment variable also enables debug mode.

Generate a curl command for the current request:

```go
request := gorequest.New().
	SetCurlCommand(true).
	Post("https://example.com").
	Send(`{"name":"gorequest"}`)

curlCommand, err := request.AsCurlCommand()
resp, body, errs := request.End()
```

## Mocking

Use `Mock` with [gock](https://github.com/h2non/gock) in tests:

```go
func TestMock(t *testing.T) {
	defer gock.Off()

	gock.New("http://foo.com").
		Get("/bar").
		Reply(200).
		JSON(map[string]string{"foo": "bar"})

	resp, body, errs := gorequest.New().
		Mock().
		Get("http://foo.com/bar").
		End()
	if len(errs) != 0 {
		t.Fatalf("expected no error, got %v", errs)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", resp.StatusCode)
	}
	if strings.TrimSpace(body) != `{"foo":"bar"}` {
		t.Fatalf("unexpected body: %s", body)
	}
}
```

## Reuse Note

GoRequest is built on top of `http.Client` in most cases. Create and reuse a
base `SuperAgent` when settings should be shared, and call `Clone` before
modifying shared settings for an individual request.
