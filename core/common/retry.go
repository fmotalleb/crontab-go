// Package common provides implementation of some of the basic functionalities to be used in application.
package common

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/sethvargo/go-retry"
)

type (
	RetryCount    = uint64
	DelayModifier string
)

const (
	RetryConstant    = DelayModifier("cons")
	RetryExponential = DelayModifier("expo")
	RetryFibonacci   = DelayModifier("fibo")
)

type RetryWrapper interface {
	Do(context.Context) error
}
type Retry struct {
	RetryWrapper

	// Backoff
	maxRetries RetryCount
	maxTimeout time.Duration
	// Retry timing
	retryDelay    time.Duration
	maxDelay      time.Duration
	jitter        time.Duration
	delayModifier DelayModifier
}

func (r *Retry) SetMaxRetry(retries uint64) {
	r.maxRetries = retries
}

func (r *Retry) SetRetryDelay(retryDelay time.Duration) {
	r.retryDelay = retryDelay
}

func (r *Retry) SetMaxTimeout(d time.Duration) {
	r.maxTimeout = d
}

func (r *Retry) SetMaxDelay(d time.Duration) {
	r.maxDelay = d
}

func (r *Retry) SetJitter(d time.Duration) {
	r.jitter = d
}

func (r *Retry) SetDelayModifierFromString(s string) {
	s = strings.ToLower(s)
	switch s {
	case "const", "cons", "constant":
		r.delayModifier = RetryConstant
	case "expo", "exponential":
		r.delayModifier = RetryExponential
	case "fibo", "fibonacci":
		r.delayModifier = RetryFibonacci
	default:
		// Maybe add some logging here
		r.delayModifier = RetryConstant
	}
}

func (r *Retry) ExecuteRetry(ctx context.Context) error {
	var backoff retry.Backoff
	switch r.delayModifier {
	case RetryConstant:
		backoff = retry.NewConstant(r.retryDelay)
	case RetryExponential:
		backoff = retry.NewExponential(r.retryDelay)
	case RetryFibonacci:
		backoff = retry.NewFibonacci(r.retryDelay)
	default:
		panic(errors.New("unknown retry delay modifier"))
	}
	if r.maxDelay != 0 {
		backoff = retry.WithCappedDuration(r.maxDelay, backoff)
	}
	if r.maxTimeout != 0 {
		backoff = retry.WithMaxDuration(r.maxTimeout, backoff)
	}
	if r.maxRetries != 0 {
		backoff = retry.WithMaxRetries(r.maxRetries, backoff)
	}
	if r.jitter != 0 {
		backoff = retry.WithJitter(r.jitter, backoff)
	}
	return retry.Do(ctx, backoff, r.Do)
}
