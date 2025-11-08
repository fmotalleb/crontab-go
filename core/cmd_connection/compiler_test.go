package connection_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"go.uber.org/zap"

	"github.com/FMotalleb/crontab-go/config"
	connection "github.com/FMotalleb/crontab-go/core/cmd_connection"
)

func TestCompileConnection_NoValidConnectionType(t *testing.T) {
	// Arrange
	conn := &config.TaskConnection{}

	// Act
	result := connection.Get(conn, zap.NewNop())

	// Assert
	assert.Equal(t, nil, result, "Expected nil result when no valid connection type is found")
}

func TestCompileConnection_LocalConnection(t *testing.T) {
	// Arrange
	conn := &config.TaskConnection{Local: true}
	// Act
	result := connection.Get(conn, zap.NewNop())
	_, ok := result.(*connection.Local)
	// Assert
	assert.True(t, ok, "Expected LocalCMDConn when Local connection type is found")
}

func TestCompileConnection_DockerAttachConnection(t *testing.T) {
	// Arrange
	conn := &config.TaskConnection{ContainerName: "testContainer", ImageName: ""}

	// Act
	result := connection.Get(conn, zap.NewNop())
	_, ok := result.(*connection.DockerAttachConnection)
	// Assert
	assert.True(t, ok, "Expected DockerAttachConnection when ContainerName is provided and ImageName is empty")
}

func TestCompileConnection_DockerCreateConnection(t *testing.T) {
	// Arrange
	conn := &config.TaskConnection{ContainerName: "", ImageName: "TestImage"}

	// Act
	result := connection.Get(conn, zap.NewNop())
	_, ok := result.(*connection.DockerCreateConnection)
	// Assert
	assert.True(t, ok, "Expected DockerAttachConnection when ContainerName is provided and ImageName is empty")
}
