package delivery

import (
	"strings"
	"time"
)

// retryDelays defines the exponential backoff schedule.
// Index 0 corresponds to the initial dispatch (delay unused),
// index 1 is the delay after the 1st failure, and so on.
var retryDelays = []time.Duration{
	0,                  // initial dispatch — not used as a retry delay
	1 * time.Minute,    // after 1st delivery failure
	5 * time.Minute,    // after 2nd delivery failure
	30 * time.Minute,   // after 3rd delivery failure
	2 * time.Hour,      // after 4th delivery failure
	24 * time.Hour,     // after 5th delivery failure and beyond
}

// RetryDelay returns the wait duration before the next retry attempt.
// For "fixed" policy, always returns 1 minute. For "exponential" (default),
// returns the delay corresponding to the given attempt count.
func RetryDelay(attempt int, policy string) time.Duration {
	if policy == "fixed" {
		return 1 * time.Minute
	}

	if attempt >= len(retryDelays) {
		return retryDelays[len(retryDelays)-1]
	}

	if attempt < 0 {
		return 0
	}

	return retryDelays[attempt]
}

// nonRetryableKeywords are error message fragments that indicate the error
// is permanent and should not be retried. If any of these appear (case-
// insensitive) in a provider error, the notification is marked "failed"
// instead of being scheduled for retry.
var nonRetryableKeywords = []string{
	"recipient email is required",
	"recipient phone is required",
	"phone number is required",
	"sms body is required",
	"invalid phone",
	"unsubscribed",
	"missing consent",
	"preference disabled",
	"suppressed",
	"template",
	"not found",
}

// isRetryableError classifies a provider send error as retryable or not.
// Errors containing known permanent-failure fragments are not retried.
// All other errors are considered transient and eligible for retry.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())

	for _, keyword := range nonRetryableKeywords {
		if strings.Contains(msg, keyword) {
			return false
		}
	}

	return true
}
