package abstraction

import (
	"context"

	"github.com/fmotalleb/crontab-go/config"
)

type CmdConnection interface {
	Prepare(context.Context, *config.Task) error
	Connect() error
	Execute() ([]byte, error)
	Disconnect() error
}
