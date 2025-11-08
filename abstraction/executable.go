// Package abstraction must contain only interfaces and abstract layers of modules
package abstraction

import (
	"context"

	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/config"
)

// Executable is an object that can be executed using a execute method and stopped using cancel method.
type Executable interface {
	Execute(context.Context) error
	SetMetaName(string)
	SetDoneHooks(context.Context, []Executable)
	SetFailHooks(context.Context, []Executable)
	Cancel()
}

type ExecutableMaker func(*zap.Logger, *config.Task) Executable
