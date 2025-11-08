package concurrency_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/fmotalleb/crontab-go/core/concurrency"
)

func TestLockedValue(t *testing.T) {
	// Create a new LockedValue with an initial value of 10
	lv := concurrency.NewLockedValue[int](10)

	// Test the Get method
	assert.Equal(t, 10, lv.Get())

	// Test the Set method
	lv.Set(20)
	assert.Equal(t, 20, lv.Get())

	// Test the Operate method
	operator := func(x int) int { return x * 2 }
	lv.Operate(operator)
	assert.Equal(t, 40, lv.Get())
}
