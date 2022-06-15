package gorequest

// UserAgent is used for setting User-Agent into headers
// Example. To set `User-Agent` as `Custom user agent`
//
//    gorequest.New().
//      Post("https://httpbin.org/post").
//      UserAgent("Custom user agent").
//      End()
func (s *SuperAgent) UserAgent(ua string) *SuperAgent {
	s.Header.Add("User-Agent", ua)
	return s
}
