package gorequest

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/proxy"
)

// Proxy function accepts a proxy url string to setup proxy url for any request.
// It provides a convenience way to setup proxy which have advantages over usual old ways.
// One example is you might try to set `http_proxy` environment. This means you are setting proxy up for all the requests.
// You will not be able to send different request with different proxy unless you change your `http_proxy` environment again.
// Another example is using Golang proxy setting. This is normal prefer way to do but too verbase compared to GoRequest's Proxy:
//
//	gorequest.New().Proxy("http://myproxy:9999").
//	  Post("http://www.google.com").
//	  End()
//
// To set no_proxy, just put empty string to Proxy func:
//
//	gorequest.New().Proxy("").
//	  Post("http://www.google.com").
//	  End()
func (s *SuperAgent) Proxy(proxyUrl string) *SuperAgent {
	parsedProxyUrl, err := url.Parse(proxyUrl)
	if err != nil {
		s.Errors = append(s.Errors, err)
	} else if proxyUrl == "" {
		s.safeModifyTransport()
		s.Transport.Proxy = nil
		s.Transport.DialContext = nil
	} else if isSocks5Proxy(parsedProxyUrl) {
		s.safeModifyTransport()
		dialContext, err := socks5DialContext(parsedProxyUrl)
		if err != nil {
			s.Errors = append(s.Errors, err)
			return s
		}
		s.Transport.Proxy = nil
		s.Transport.DialContext = dialContext
	} else {
		s.safeModifyTransport()
		s.Transport.Proxy = http.ProxyURL(parsedProxyUrl)
		s.Transport.DialContext = nil
	}
	return s
}

func proxyFromEnvironment(req *http.Request) (*url.URL, error) {
	if req == nil || req.URL == nil {
		return nil, nil
	}
	if bypassProxy(req.URL.Hostname(), req.URL.Host) {
		return nil, nil
	}

	proxyURL := proxyEnvValue(req.URL.Scheme)
	if proxyURL == "" {
		return nil, nil
	}
	if !strings.Contains(proxyURL, "://") {
		proxyURL = req.URL.Scheme + "://" + proxyURL
	}
	return url.Parse(proxyURL)
}

func proxyEnvValue(scheme string) string {
	switch strings.ToLower(scheme) {
	case "https":
		if proxyURL := firstEnv("HTTPS_PROXY", "https_proxy"); proxyURL != "" {
			return proxyURL
		}
	case "http":
		if proxyURL := firstEnv("HTTP_PROXY", "http_proxy"); proxyURL != "" {
			return proxyURL
		}
	}
	return firstEnv("ALL_PROXY", "all_proxy")
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}

func bypassProxy(hostname, host string) bool {
	if hostname == "localhost" {
		return true
	}
	if ip := net.ParseIP(hostname); ip != nil && ip.IsLoopback() {
		return true
	}

	noProxy := firstEnv("NO_PROXY", "no_proxy")
	if noProxy == "" {
		return false
	}
	for _, entry := range strings.Split(noProxy, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if entry == "*" || strings.EqualFold(entry, host) || strings.EqualFold(entry, hostname) {
			return true
		}
		if strings.HasPrefix(entry, ".") && strings.HasSuffix(hostname, entry) {
			return true
		}
	}
	return false
}

func isSocks5Proxy(proxyURL *url.URL) bool {
	scheme := strings.ToLower(proxyURL.Scheme)
	return scheme == "socks5" || scheme == "socks5h"
}

func socks5DialContext(proxyURL *url.URL) (func(context.Context, string, string) (net.Conn, error), error) {
	var auth *proxy.Auth
	if proxyURL.User != nil {
		password, _ := proxyURL.User.Password()
		auth = &proxy.Auth{
			User:     proxyURL.User.Username(),
			Password: password,
		}
	}

	dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, network string, address string) (net.Conn, error) {
		type dialResult struct {
			conn net.Conn
			err  error
		}
		result := make(chan dialResult, 1)
		go func() {
			conn, err := dialer.Dial(network, address)
			result <- dialResult{conn: conn, err: err}
		}()

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("socks5 dial canceled: %w", ctx.Err())
		case res := <-result:
			return res.conn, res.err
		}
	}, nil
}
