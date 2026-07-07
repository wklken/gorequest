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
	Policy          RetryPolicy
	Enable          bool
}

// RetryPolicy decides whether a completed request attempt should be retried.
type RetryPolicy func(resp Response, body []byte, errs []error) bool

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
		Policy          RetryPolicy
		Enable          bool
	}{
		statusCode,
		retryerTime,
		retryerCount,
		0,
		s.Retryable.Policy,
		true,
	}
	return s
}

// SetRetryPolicy sets a custom policy for deciding whether an attempt should be retried.
func (s *SuperAgent) SetRetryPolicy(policy RetryPolicy) *SuperAgent {
	s.Retryable.Policy = policy
	return s
}

func (s *SuperAgent) shouldRetry(resp Response, body []byte, errs []error) bool {
	if !s.Retryable.Enable || s.Retryable.Attempt >= s.Retryable.RetryerCount {
		return false
	}

	if s.retryPolicyMatched(resp, body, errs) {
		time.Sleep(s.Retryable.RetryerTime)
		s.Retryable.Attempt++
		return true
	}
	return false
}

func (s *SuperAgent) retryPolicyMatched(resp Response, body []byte, errs []error) bool {
	if s.Retryable.Policy != nil {
		return s.Retryable.Policy(resp, body, errs)
	}
	return len(errs) > 0 || statusesContains(s.Retryable.RetryableStatus, resp.StatusCode)
}

func (s *SuperAgent) getResponseBytesWithRetry() (Response, []byte, []error) {
	var (
		errs []error
		resp Response
		body []byte
	)

	for {
		resp, body, errs = s.getResponseBytes()
		if !s.shouldRetry(resp, body, errs) {
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
