// Package connection provides implementation of the abstraction.CmdConnection interface for command tasks.
package connection

import (
	"go.uber.org/zap"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/generator"
)

var cg = generator.New[*config.TaskConnection, abstraction.CmdConnection]()

// Get compiles the task connection based on the provided configuration and logger.
// It returns an abstraction.CmdConnection interface based on the type of connection specified in the configuration.
// If the connection type is not recognized or invalid, it logs a fatal error and returns nil.
func Get(conn *config.TaskConnection, logger *zap.Logger) abstraction.CmdConnection {
	con, ok := cg.Get(logger, conn)
	if ok {
		return con
	}
	logger.Error("cannot compile given taskConnection", zap.Any("connection", conn))
	return nil
}
