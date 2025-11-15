// Package ctxutils provides utility functions for working with context.Context.
package ctxutils

type ContextKey string

var (
	ScopeKey       = ContextKey("scope")
	LoggerKey      = ContextKey("logger")
	JobKey         = ContextKey("job")
	TaskKey        = ContextKey("task")
	EventListeners = ContextKey("event-listeners")
	EventData      = ContextKey("event-data")
	Environments   = ContextKey("cmd-environments")
	Vars           = ContextKey("cmd-vars")
)
