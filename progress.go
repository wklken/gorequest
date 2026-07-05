package gorequest

import (
	"io"
	"net/http"
)

// UploadProgress receives the cumulative number of request body bytes uploaded.
type UploadProgress func(uploaded int64)

// SetUploadProgress sets a callback for request body upload progress.
func (s *SuperAgent) SetUploadProgress(progress UploadProgress) *SuperAgent {
	s.uploadProgress = progress
	return s
}

func (s *SuperAgent) wrapUploadProgress(req *http.Request) {
	if s.uploadProgress == nil || req.Body == nil {
		return
	}

	var uploaded int64
	report := func(n int) {
		uploaded += int64(n)
		s.uploadProgress(uploaded)
	}

	req.Body = &uploadProgressReadCloser{
		ReadCloser: req.Body,
		report:     report,
	}

	if req.GetBody == nil {
		return
	}
	getBody := req.GetBody
	req.GetBody = func() (io.ReadCloser, error) {
		body, err := getBody()
		if err != nil {
			return nil, err
		}
		return &uploadProgressReadCloser{
			ReadCloser: body,
			report:     report,
		}, nil
	}
}

type uploadProgressReadCloser struct {
	io.ReadCloser
	report func(int)
}

func (r *uploadProgressReadCloser) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	if n > 0 {
		r.report(n)
	}
	return n, err
}
