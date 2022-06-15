package gorequest

import "crypto/tls"

// TLSClientConfig set TLSClientConfig for underling Transport.
// One example is you can use it to disable security check (https):
//
//      gorequest.New().TLSClientConfig(&tls.Config{ InsecureSkipVerify: true}).
//        Get("https://disable-security-check.com").
//        End()
//
func (s *SuperAgent) TLSClientConfig(config *tls.Config) *SuperAgent {
	s.safeModifyTransport()
	s.Transport.TLSClientConfig = config
	return s
}
