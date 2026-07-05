package gorequest

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestConfigurationSettersAndCloneIsolation(t *testing.T) {
	base := New()
	base.Client.Transport = base.Transport
	base.DisableCompression().
		SetCurlCommand(true).
		SetDoNotClearSuperAgent(true)

	if !base.Transport.DisableCompression {
		t.Fatal("Expected DisableCompression to update the transport")
	}
	if !base.CurlCommand {
		t.Fatal("Expected SetCurlCommand to enable curl logging")
	}

	base.Get("http://example.com")
	base.ClearSuperAgent()
	if base.Method != GET {
		t.Fatalf("Expected DoNotClearSuperAgent to preserve method %q, got %q", GET, base.Method)
	}

	clone := base.Clone()
	clone.SetDoNotClearSuperAgent(false)
	clone.Timeout(2 * time.Second)
	if clone.Client == base.Client {
		t.Fatal("Expected cloned agent timeout changes to allocate a separate client")
	}
	if base.Client.Timeout != 0 {
		t.Fatalf("Expected base client timeout to stay unchanged, got %s", base.Client.Timeout)
	}
	if clone.Client.Timeout != 2*time.Second {
		t.Fatalf("Expected clone timeout to be set, got %s", clone.Client.Timeout)
	}

	cfg := &tls.Config{ServerName: "example.com"}
	clone.TLSClientConfig(cfg)
	if clone.Transport == base.Transport {
		t.Fatal("Expected cloned agent TLS changes to allocate a separate transport")
	}
	if clone.Transport.TLSClientConfig != cfg {
		t.Fatal("Expected clone transport to receive TLS config")
	}
	if base.Transport.TLSClientConfig != nil {
		t.Fatal("Expected base transport TLS config to stay unchanged")
	}

	clone.Proxy("")
	if clone.Transport.Proxy != nil {
		t.Fatal("Expected empty proxy URL to clear the proxy function")
	}

	withBadProxy := New().Proxy("http://[::1")
	if len(withBadProxy.Errors) == 0 {
		t.Fatal("Expected invalid proxy URL to be recorded as an error")
	}
}

func TestTimeoutsUpdatesTransportAndIgnoresNonTransport(t *testing.T) {
	agent := New()
	agent.Client.Transport = agent.Transport
	agent.Timeouts(&Timeouts{
		Dial:           time.Second,
		KeepAlive:      2 * time.Second,
		TlsHandshake:   3 * time.Second,
		ResponseHeader: 4 * time.Second,
		ExpectContinue: 5 * time.Second,
		IdleConn:       6 * time.Second,
	})

	transport, ok := agent.Client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Expected *http.Transport, got %T", agent.Client.Transport)
	}
	if transport.DialContext == nil {
		t.Fatal("Expected Timeouts to configure DialContext")
	}
	if transport.TLSHandshakeTimeout != 3*time.Second {
		t.Fatalf("Expected TLSHandshakeTimeout=3s, got %s", transport.TLSHandshakeTimeout)
	}
	if transport.ResponseHeaderTimeout != 4*time.Second {
		t.Fatalf("Expected ResponseHeaderTimeout=4s, got %s", transport.ResponseHeaderTimeout)
	}

	nonTransport := New()
	nonTransport.Client.Transport = roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return nil, errors.New("not used")
	})
	if nonTransport.Timeouts(&Timeouts{}) != nonTransport {
		t.Fatal("Expected Timeouts to return the same agent when transport is not *http.Transport")
	}
}

func TestCurlCommandLoggingAndAsCurlErrors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := io.WriteString(w, "ok"); err != nil {
			t.Fatalf("Unexpected write error: %s", err)
		}
	}))
	defer ts.Close()

	curl, err := New().Post(ts.URL).Send(`{"hello":"world"}`).AsCurlCommand()
	if err != nil {
		t.Fatalf("Unexpected AsCurlCommand error: %s", err)
	}
	if !strings.Contains(curl, "curl") || !strings.Contains(curl, "hello") {
		t.Fatalf("Expected curl command to include curl and request body, got %q", curl)
	}

	if _, err := New().AsCurlCommand(); err == nil {
		t.Fatal("Expected AsCurlCommand without a method to return an error")
	}

	var buf bytes.Buffer
	logger := log.New(&buf, "[gorequest]", 0)
	resp, _, errs := New().SetLogger(logger).SetCurlCommand(true).Get(ts.URL).End()
	if len(errs) > 0 {
		t.Fatalf("Unexpected request errors: %s", errs)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
	if !strings.Contains(buf.String(), "CURL command line") {
		t.Fatalf("Expected curl command log, got %q", buf.String())
	}
}

func TestEndStructDecodeErrors(t *testing.T) {
	textServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := io.WriteString(w, "not json"); err != nil {
			t.Fatalf("Unexpected write error: %s", err)
		}
	}))
	defer textServer.Close()

	var textResult heyYou
	resp, body, errs := New().Get(textServer.URL).EndStruct(&textResult)
	if len(errs) == 0 {
		t.Fatal("Expected non-json response to return a decode error")
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
	if string(body) != "not json" {
		t.Fatalf("Expected original response body, got %q", string(body))
	}
	if !strings.Contains(errs[0].Error(), "not application/json") {
		t.Fatalf("Expected content-type decode error, got %q", errs[0].Error())
	}

	jsonServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", MIMEJSON+"; charset=utf-8")
		if _, err := io.WriteString(w, "{"); err != nil {
			t.Fatalf("Unexpected write error: %s", err)
		}
	}))
	defer jsonServer.Close()

	var jsonResult heyYou
	_, _, errs = New().Get(jsonServer.URL).EndStruct(&jsonResult)
	if len(errs) == 0 {
		t.Fatal("Expected invalid JSON response to return a decode error")
	}
	if !strings.Contains(errs[0].Error(), "response body json decode fail") {
		t.Fatalf("Expected JSON decode error, got %q", errs[0].Error())
	}

	agent := New()
	agent.Errors = []error{errors.New("preflight")}
	if resp, body, errs := agent.EndStruct(&jsonResult); resp != nil || body != nil || len(errs) != 1 {
		t.Fatalf("Expected preflight error to short-circuit EndStruct, got resp=%v body=%v errs=%v", resp, body, errs)
	}
}

func TestCustomMethodAndMakeRequestEdges(t *testing.T) {
	for _, method := range []string{GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS, "TRACE"} {
		req, err := New().CustomMethod(method, "http://example.com/path").MakeRequest()
		if err != nil {
			t.Fatalf("Unexpected MakeRequest error for %s: %s", method, err)
		}
		if req.Method != method {
			t.Fatalf("Expected method %q, got %q", method, req.Method)
		}
	}

	req, err := New().Get("http://example.com").Set("Host", "api.example.com").MakeRequest()
	if err != nil {
		t.Fatalf("Unexpected MakeRequest error with Host header: %s", err)
	}
	if req.Host != "api.example.com" {
		t.Fatalf("Expected request Host to be overridden, got %q", req.Host)
	}

	agent := New().Get("http://example.com")
	agent.TargetType = "bad-type"
	if _, err := agent.MakeRequest(); err == nil {
		t.Fatal("Expected unsupported target type to return an error")
	}
}

func TestDisableRedirectReturnsRedirectResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/target", http.StatusFound)
			return
		}
		if _, err := io.WriteString(w, "target"); err != nil {
			t.Fatalf("Unexpected write error: %s", err)
		}
	}))
	defer ts.Close()

	resp, _, errs := New().DisableRedirect().Get(ts.URL + "/redirect").End()
	if len(errs) > 0 {
		t.Fatalf("Unexpected request errors: %s", errs)
	}
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("Expected redirect response to be returned, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Location") != "/target" {
		t.Fatalf("Expected Location header to be preserved, got %q", resp.Header.Get("Location"))
	}
}

func TestHttpTraceAppliesClientTrace(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := io.WriteString(w, "ok"); err != nil {
			t.Fatalf("Unexpected write error: %s", err)
		}
	}))
	defer ts.Close()

	gotFirstResponseByte := false
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			gotFirstResponseByte = true
		},
	}

	resp, _, errs := New().HttpTrace(trace).Get(ts.URL).End()
	if len(errs) > 0 {
		t.Fatalf("Unexpected request errors: %s", errs)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
	if !gotFirstResponseByte {
		t.Fatal("Expected client trace callback to run")
	}
}

func TestQuerySendAndURLValueConversions(t *testing.T) {
	type queryPayload struct {
		Name   string
		Count  int64
		Price  float64
		Active bool
		Nested map[string]string
	}

	agent := New().
		Query(`{"json":"yes"}`).
		Query("a=1&a=2").
		Query(queryPayload{
			Name:   "gorequest",
			Count:  6673221165400540161,
			Price:  3.5,
			Active: true,
			Nested: map[string]string{"k": "v"},
		}).
		Query(map[string]any{"map": "value"}).
		Query("%zz")

	if agent.QueryData.Get("json") != "yes" {
		t.Fatalf("Expected JSON query value, got %q", agent.QueryData.Get("json"))
	}
	if got := agent.QueryData["a"]; !reflect.DeepEqual(got, []string{"1", "2"}) {
		t.Fatalf("Expected duplicate query values, got %v", got)
	}
	if agent.QueryData.Get("Count") != "6673221165400540161" {
		t.Fatalf("Expected int64 query to preserve precision, got %q", agent.QueryData.Get("Count"))
	}
	if agent.QueryData.Get("Nested") != `{"k":"v"}` {
		t.Fatalf("Expected nested query to be encoded as JSON, got %q", agent.QueryData.Get("Nested"))
	}
	if len(agent.Errors) == 0 {
		t.Fatal("Expected invalid query string to append an error")
	}

	sendAgent := New().
		Send(map[string]any{"name": "gorequest"}).
		Send([]int{1, 2}).
		Send(complex64(1 + 2i))
	if sendAgent.Data["name"] != "gorequest" {
		t.Fatalf("Expected Send(map) to fill Data, got %v", sendAgent.Data)
	}
	if !reflect.DeepEqual(sendAgent.SliceData, []any{1, 2}) {
		t.Fatalf("Expected Send(slice) to fill SliceData, got %v", sendAgent.SliceData)
	}

	formAgent := New().Send("a=1&a=2")
	if got := formAgent.Data["a"]; !reflect.DeepEqual(got, []string{"2", "1"}) {
		t.Fatalf("Expected duplicate form values to be preserved, got %v", got)
	}

	values := changeMapToURLValues(map[string]any{
		"float32": []float32{1.5, 2.5},
		"strings": []any{
			"one",
			"two",
		},
		"bools": []any{
			true,
			false,
		},
		"numbers": []any{
			json.Number("9"),
			json.Number("10"),
		},
		"empty": []any{},
	})
	if got := values["float32"]; !reflect.DeepEqual(got, []string{"1.5", "2.5"}) {
		t.Fatalf("Expected float32 values, got %v", got)
	}
	if got := values["strings"]; !reflect.DeepEqual(got, []string{"one", "two"}) {
		t.Fatalf("Expected string interface values, got %v", got)
	}
	if got := values["bools"]; !reflect.DeepEqual(got, []string{"true", "false"}) {
		t.Fatalf("Expected bool interface values, got %v", got)
	}
	if got := values["numbers"]; !reflect.DeepEqual(got, []string{"9", "10"}) {
		t.Fatalf("Expected json.Number interface values, got %v", got)
	}
	if _, ok := values["empty"]; ok {
		t.Fatal("Expected empty interface slice to be skipped")
	}

	if got := makeSliceOfReflectValue(reflect.ValueOf("not a slice")); got != nil {
		t.Fatalf("Expected non-slice input to return nil, got %v", got)
	}
}

func TestCopyHelpersNilAndIsolation(t *testing.T) {
	if shallowCopyData(nil) != nil {
		t.Fatal("Expected nil map copy to stay nil")
	}
	if shallowCopyDataSlice(nil) != nil {
		t.Fatal("Expected nil data slice copy to stay nil")
	}
	if shallowCopyFileArray(nil) != nil {
		t.Fatal("Expected nil file slice copy to stay nil")
	}
	if shallowCopyCookies(nil) != nil {
		t.Fatal("Expected nil cookie slice copy to stay nil")
	}
	if shallowCopyErrors(nil) != nil {
		t.Fatal("Expected nil error slice copy to stay nil")
	}

	retryable := superAgentRetryable{RetryableStatus: []int{http.StatusTooManyRequests}}
	retryableCopy := copyRetryable(retryable)
	retryableCopy.RetryableStatus[0] = http.StatusInternalServerError
	if retryable.RetryableStatus[0] != http.StatusTooManyRequests {
		t.Fatal("Expected copyRetryable to clone RetryableStatus slice")
	}

	errs := []error{errors.New("first")}
	errsCopy := shallowCopyErrors(errs)
	errsCopy[0] = errors.New("second")
	if errs[0].Error() != "first" {
		t.Fatal("Expected shallowCopyErrors to clone the slice")
	}

	if filterFlags("application/json; charset=utf-8") != MIMEJSON {
		t.Fatal("Expected filterFlags to remove content-type flags")
	}
	if filterFlags("text/plain charset=utf-8") != "text/plain" {
		t.Fatal("Expected filterFlags to stop at spaces")
	}
	if filterFlags(MIMEJSON) != MIMEJSON {
		t.Fatal("Expected filterFlags to return content without flags unchanged")
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
