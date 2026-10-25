package global

import "testing"

func TestEventListenersReturnsSnapshot(t *testing.T) {
	const eventName = "test-listener-snapshot"
	CTX().AddEventListener(eventName, func(map[string]any) {})

	snapshot := CTX().EventListeners()
	delete(snapshot, eventName)
	if len(CTX().EventListeners()[eventName]) != 1 {
		t.Fatal("mutating the returned listener map changed registered listeners")
	}
}
