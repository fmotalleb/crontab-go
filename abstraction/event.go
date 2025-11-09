// Package abstraction must contain only interfaces and abstract layers of modules
package abstraction

import "github.com/maniartech/signals"

type EventGenerator interface {
	BuildTickChannel(EventDispatcher)
}

type (
	EventDispatcher = signals.Signal[Event]
)

type Event interface {
	GetData() map[string]any
}
