package gorequest

import "net/http/httptrace"

func (s *SuperAgent) HttpTrace(trace *httptrace.ClientTrace) *SuperAgent {
	s.trace = trace
	return s
}
