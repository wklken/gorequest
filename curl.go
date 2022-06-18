package gorequest

import (
	"net/http"

	"moul.io/http2curl"
)

// SetCurlCommand enable the curlcommand mode which display a CURL command line.
func (s *SuperAgent) SetCurlCommand(enable bool) *SuperAgent {
	s.CurlCommand = enable
	return s
}

// AsCurlCommand returns a string representing the runnable `curl' command
// version of the request.
func (s *SuperAgent) AsCurlCommand() (string, error) {
	// FIXME: why here MakeRequest again?
	req, err := s.MakeRequest()
	if err != nil {
		return "", err
	}
	cmd, err := http2curl.GetCurlCommand(req)
	if err != nil {
		return "", err
	}
	return cmd.String(), nil
}

func (s *SuperAgent) logCurlCommand(req *http.Request) {
	if s.CurlCommand {
		curl, err := http2curl.GetCurlCommand(req)
		s.logger.SetPrefix("[curl] ")
		if err != nil {
			s.logger.Println("Error:", err)
		} else {
			s.logger.Printf("CURL command line: %s", curl)
		}
	}
}
