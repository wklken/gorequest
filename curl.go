package gorequest

import "moul.io/http2curl"

// SetCurlCommand enable the curlcommand mode which display a CURL command line.
func (s *SuperAgent) SetCurlCommand(enable bool) *SuperAgent {
	s.CurlCommand = enable
	return s
}

// AsCurlCommand returns a string representing the runnable `curl' command
// version of the request.
func (s *SuperAgent) AsCurlCommand() (string, error) {
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
