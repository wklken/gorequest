package gorequest

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/http/httpproxy"
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
	return httpproxy.FromEnvironment().ProxyFunc()(req.URL)
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
	if contextDialer, ok := dialer.(proxy.ContextDialer); ok {
		return contextDialer.DialContext, nil
	}

	return func(ctx context.Context, network string, address string) (net.Conn, error) {
		type dialResult struct {
			conn net.Conn
			err  error
		}
		result := make(chan dialResult, 1)
		go func() {
			conn, err := dialer.Dial(network, address)
			select {
			case result <- dialResult{conn: conn, err: err}:
			case <-ctx.Done():
				if conn != nil {
					conn.Close()
				}
			}
		}()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case res := <-result:
			return res.conn, res.err
		}
	}, nil
}
