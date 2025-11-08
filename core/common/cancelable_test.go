package common_test

import (
	"testing"

	"github.com/fmotalleb/crontab-go/core/common"
)

func TestSetCancel(t *testing.T) {
	c := &common.Cancelable{}
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetCancel() panicked: %v", r)
		}
	}()
	c.SetCancel(func() {})
}

func TestCancelWithoutSetting(t *testing.T) {
	c := &common.Cancelable{}
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Cancel() panicked: %v", r)
		}
	}()
	c.Cancel() // Should not panic even if cancel is not set
}

func TestCancelAfterSetting(t *testing.T) {
	c := &common.Cancelable{}
	called := false
	c.SetCancel(func() {
		called = true
	})
	c.Cancel()
	if !called {
		t.Errorf("Cancel() did not call the set function")
	}
}

func TestCancelMultipleTimes(t *testing.T) {
	c := &common.Cancelable{}
	callCount := 0
	c.SetCancel(func() {
		callCount++
	})
	c.Cancel()
	c.Cancel()
	if callCount != 2 {
		t.Errorf("Cancel() called the function %d times, expected 2", callCount)
	}
}

func TestSetCancelOverwrite(t *testing.T) {
	c := &common.Cancelable{}
	firstCalled := false
	secondCalled := false
	c.SetCancel(func() {
		firstCalled = true
	})
	c.SetCancel(func() {
		secondCalled = true
	})
	c.Cancel()
	if firstCalled {
		t.Errorf("First cancel function was called, but it should have been overwritten")
	}
	if !secondCalled {
		t.Errorf("Second cancel function was not called")
	}
}
