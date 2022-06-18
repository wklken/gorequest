package gorequest

import "net/http"

// AddCookie adds a cookie to the request. The behavior is the same as AddCookie on Request from net/http
func (s *SuperAgent) AddCookie(c *http.Cookie) *SuperAgent {
	s.Cookies = append(s.Cookies, c)
	return s
}

// AddCookies is a convenient method to add multiple cookies
func (s *SuperAgent) AddCookies(cookies []*http.Cookie) *SuperAgent {
	s.Cookies = append(s.Cookies, cookies...)
	return s
}

func shallowCopyCookies(old []*http.Cookie) []*http.Cookie {
	if old == nil {
		return nil
	}
	newData := make([]*http.Cookie, len(old))
	copy(newData, old)
	return newData
}
