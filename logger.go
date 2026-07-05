package gorequest

type Logger interface {
	SetPrefix(string)
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

// SetLogger set the logger which is the default logger to the SuperAgent instance.
func (s *SuperAgent) SetLogger(logger Logger) *SuperAgent {
	s.logger = logger
	return s
}
