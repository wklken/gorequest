package gorequest

import (
	"net/http"
	"net/http/httputil"
)

// SetDebug enable the debug mode which logs request/response detail.
func (s *SuperAgent) SetDebug(enable bool) *SuperAgent {
	s.Debug = enable
	return s
}

func (s *SuperAgent) debuggingRequest(req *http.Request) {
	if s.Debug {
		dump, err := httputil.DumpRequest(req, true)
		s.logger.SetPrefix("[http] ")
		if err != nil {
			s.logger.Println("Error:", err)
		} else {
			s.logger.Printf("HTTP Request: %s", BytesToString(dump))
		}
	}
}

func (s *SuperAgent) debuggingResponse(resp *http.Response) {
	if s.Debug {
		dump, err := httputil.DumpResponse(resp, true)
		if nil != err {
			s.logger.Println("Error:", err)
		} else {
			s.logger.Printf("HTTP Response: %s", BytesToString(dump))
		}
	}
}
