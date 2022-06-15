package gorequest

type basicAuth struct {
	Username string
	Password string
}

// SetBasicAuth sets the basic authentication header
// Example. To set the header for username "my_user" and password "my_pass"
//
//    gorequest.New()
//      Post("https://httpbin.org/post").
//      SetBasicAuth("my_user", "my_pass").
//      End()
func (s *SuperAgent) SetBasicAuth(username string, password string) *SuperAgent {
	s.BasicAuth = basicAuth{username, password}
	return s
}
