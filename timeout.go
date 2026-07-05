package gorequest

import (
	"net"
	"net/http"
	"time"
)

func (s *SuperAgent) Timeout(timeout time.Duration) *SuperAgent {
	s.safeModifyHttpClient()
	s.Client.Timeout = timeout
	return s
}

type Timeouts struct {
	Dial      time.Duration
	KeepAlive time.Duration

	TlsHandshake   time.Duration
	ResponseHeader time.Duration
	ExpectContinue time.Duration
	IdleConn       time.Duration
}

func (s *SuperAgent) Timeouts(timeouts *Timeouts) *SuperAgent {
	s.safeModifyHttpClient()

	transport, ok := s.Client.Transport.(*http.Transport)
	if !ok {
		return s
	}

	transport.DialContext = (&net.Dialer{
		Timeout:   timeouts.Dial,
		KeepAlive: timeouts.KeepAlive,
	}).DialContext

	transport.TLSHandshakeTimeout = timeouts.TlsHandshake
	transport.ResponseHeaderTimeout = timeouts.ResponseHeader
	transport.ExpectContinueTimeout = timeouts.ExpectContinue
	transport.ExpectContinueTimeout = timeouts.IdleConn

	s.Client.Transport = transport

	return s
}
