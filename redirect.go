package gorequest

import (
	"errors"
	"net/http"
	"strings"
)

// RedirectPolicy accepts a function to define how to handle redirects. If the
// policy function returns an error, the next Request is not made and the previous
// request is returned.
//
// The policy function's arguments are the Request about to be made and the
// past requests in order of oldest first.
func (s *SuperAgent) RedirectPolicy(policy func(req Request, via []Request) error) *SuperAgent {
	s.safeModifyHttpClient()
	s.Client.CheckRedirect = func(r *http.Request, v []*http.Request) error {
		stripCustomSensitiveHeadersOnRedirect(r, v)
		vv := make([]Request, len(v))
		for i, r := range v {
			vv[i] = Request(r)
		}
		return policy(Request(r), vv)
	}
	return s
}

func defaultRedirectPolicy(req *http.Request, via []*http.Request) error {
	stripCustomSensitiveHeadersOnRedirect(req, via)
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	return nil
}

func stripCustomSensitiveHeadersOnRedirect(req *http.Request, via []*http.Request) {
	if len(via) == 0 || req.URL == nil || via[0].URL == nil {
		return
	}
	if sameRedirectHost(req.URL.Host, via[0].URL.Host) {
		return
	}

	for _, header := range []string{
		"API-Key",
		"X-API-Key",
		"X-Auth-Key",
		"X-Auth-Token",
		"X-API-Token",
		"X-Access-Token",
	} {
		req.Header.Del(header)
	}
}

func sameRedirectHost(a, b string) bool {
	return strings.EqualFold(a, b)
}

// DisableRedirect will disable the redirect of status code 3xx.
func (s *SuperAgent) DisableRedirect() *SuperAgent {
	s.Client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return s
}
