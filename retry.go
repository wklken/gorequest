package gorequest

import (
	"fmt"
	"net/http"
	"time"
)

type superAgentRetryable struct {
	RetryStatus []int
	RetryTime   time.Duration
	RetryCount  int
	Attempt     int
	Enable      bool
}

// Retry is used for setting a Retry policy
// Example. To set Retry policy with 5 seconds between each attempt.
//          3 max attempt.
//          And StatusBadRequest and StatusInternalServerError as RetryStatus
//
//    gorequest.New().
//      Post("https://httpbin.org/post").
//      Retry(3, 5 * time.Second, http.StatusBadRequest, http.StatusInternalServerError).
//      End()
func (s *SuperAgent) Retry(retryCount int, retryTime time.Duration, statusCode ...int) *SuperAgent {
	for _, code := range statusCode {
		statusText := http.StatusText(code)
		if len(statusText) == 0 {
			s.Errors = append(s.Errors, fmt.Errorf("StatusCode '%d' doesn't exist in http package", code))
		}
	}

	s.Retryable = struct {
		RetryStatus []int
		RetryTime   time.Duration
		RetryCount  int
		Attempt     int
		Enable      bool
	}{
		statusCode,
		retryTime,
		retryCount,
		0,
		true,
	}
	return s
}

func (s *SuperAgent) shouldRetry(resp Response, hasError bool) bool {
	if s.Retryable.Enable && s.Retryable.Attempt < s.Retryable.RetryCount &&
		(hasError || statusesContains(s.Retryable.RetryStatus, resp.StatusCode)) {
		time.Sleep(s.Retryable.RetryTime)
		s.Retryable.Attempt++
		return true
	}
	return false
}

// just need to change the array pointer?
func copyRetryable(old superAgentRetryable) superAgentRetryable {
	newRetryable := old
	newRetryable.RetryStatus = make([]int, len(old.RetryStatus))
	for i := range old.RetryStatus {
		newRetryable.RetryStatus[i] = old.RetryStatus[i]
	}
	return newRetryable
}
