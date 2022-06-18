package gorequest

import "gopkg.in/h2non/gock.v1"

// Mock will enable gock, http mocking for net/http
func (s *SuperAgent) Mock() *SuperAgent {
	gock.InterceptClient(s.Client)
	s.isMock = true
	return s
}
