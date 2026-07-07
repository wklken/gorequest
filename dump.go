package gorequest

import "net/http/httputil"

// Dump returns the outgoing HTTP request bytes without sending the request.
func (s *SuperAgent) Dump() (string, error) {
	req, err := s.MakeRequest()
	if err != nil {
		return "", err
	}
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return "", err
	}
	return BytesToString(dump), nil
}
