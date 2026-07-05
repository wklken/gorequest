package gorequest

type basicAuth struct {
	Username string
	Password string
}

// SetBasicAuth sets the basic authentication header
// Example. To set the header for username "myuser" and password "mypass"
//
//	gorequest.New()
//	  Post("https://httpbin.org/post").
//	  SetBasicAuth("myuser", "mypass").
//	  End()
func (s *SuperAgent) SetBasicAuth(username string, password string) *SuperAgent {
	s.BasicAuth = basicAuth{username, password}
	return s
}
