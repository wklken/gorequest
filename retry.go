package gorequest

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type superAgentRetryable struct {
	RetryableStatus []int
	RetryerTime     time.Duration
	RetryerCount    int
	Attempt         int
	Enable          bool
}

// Retry is used for setting a Retryer policy
// Example. To set Retryer policy with 5 seconds between each attempt.
//
//	      3 max attempt.
//	      And StatusBadRequest and StatusInternalServerError as RetryableStatus
//
//	gorequest.New().
//	  Post("https://httpbin.org/post").
//	  Retry(3, 5 * time.Second, http.StatusBadRequest, http.StatusInternalServerError).
//	  End()
func (s *SuperAgent) Retry(retryerCount int, retryerTime time.Duration, statusCode ...int) *SuperAgent {
	for _, code := range statusCode {
		statusText := http.StatusText(code)
		if len(statusText) == 0 {
			s.Errors = append(s.Errors, fmt.Errorf("StatusCode '%d' doesn't exist in http package", code))
		}
	}

	s.Retryable = struct {
		RetryableStatus []int
		RetryerTime     time.Duration
		RetryerCount    int
		Attempt         int
		Enable          bool
	}{
		statusCode,
		retryerTime,
		retryerCount,
		0,
		true,
	}
	return s
}

func (s *SuperAgent) shouldRetry(resp Response, hasError bool) bool {
	if s.Retryable.Enable && s.Retryable.Attempt < s.Retryable.RetryerCount &&
		(hasError || statusesContains(s.Retryable.RetryableStatus, resp.StatusCode)) {
		time.Sleep(s.Retryable.RetryerTime)
		s.Retryable.Attempt++
		return true
	}
	return false
}

func (s *SuperAgent) getResponseBytesWithRetry() (Response, []byte, []error) {
	var (
		errs []error
		resp Response
		body []byte
	)

	for {
		resp, body, errs = s.getResponseBytes()
		if !s.shouldRetry(resp, len(errs) > 0) {
			s.setRetryCountHeader(resp)
			break
		}

		s.Errors = nil
	}

	return resp, body, errs
}

func (s *SuperAgent) setRetryCountHeader(resp Response) {
	if resp != nil {
		resp.Header.Set("Retry-Count", strconv.Itoa(s.Retryable.Attempt))
	}
}

// just need to change the array pointer?
func copyRetryable(old superAgentRetryable) superAgentRetryable {
	newRetryable := old
	newRetryable.RetryableStatus = make([]int, len(old.RetryableStatus))
	copy(newRetryable.RetryableStatus, old.RetryableStatus)
	return newRetryable
}
