package connection_test

// import (
// 	"testing"

// 	"github.com/alecthomas/assert/v2"
// 	logger "github.com/sirupsen/logrus"

// 	"github.com/FMotalleb/crontab-go/config"
// 	connection "github.com/FMotalleb/crontab-go/core/cmd_connection"
// 	mocklogger "github.com/FMotalleb/crontab-go/logger/mock_logger"
// )

// func TestCompileConnection_NoValidConnectionType(t *testing.T) {
// 	// Arrange
// 	conn := &config.TaskConnection{}
// 	log, output := mocklogger.HijackOutput(logger.New())
// 	expectedError := "cannot compile given taskConnection"

// 	// var capturedError string

// 	// Act
// 	result := connection.Get(conn, log.WithField("test", "TestCompileConnection_NoValidConnectionType"))

// 	// Assert
// 	assert.Equal(t, nil, result, "Expected nil result when no valid connection type is found")
// 	assert.Contains(t, output.String(), expectedError, "Expected error message to be captured")
// }

// func TestCompileConnection_LocalConnection(t *testing.T) {
// 	// Arrange
// 	conn := &config.TaskConnection{Local: true}
// 	log, _ := mocklogger.HijackOutput(logger.New())

// 	// Act
// 	result := connection.Get(conn, log.WithField("test", "test"))
// 	_, ok := result.(*connection.Local)
// 	// Assert
// 	assert.True(t, ok, "Expected LocalCMDConn when Local connection type is found")
// }

// func TestCompileConnection_DockerAttachConnection(t *testing.T) {
// 	// Arrange
// 	conn := &config.TaskConnection{ContainerName: "testContainer", ImageName: ""}
// 	log, _ := mocklogger.HijackOutput(logger.New())

// 	// Act
// 	result := connection.Get(conn, log.WithField("test", "test"))
// 	_, ok := result.(*connection.DockerAttachConnection)
// 	// Assert
// 	assert.True(t, ok, "Expected DockerAttachConnection when ContainerName is provided and ImageName is empty")
// }

// func TestCompileConnection_DockerCreateConnection(t *testing.T) {
// 	// Arrange
// 	conn := &config.TaskConnection{ContainerName: "", ImageName: "TestImage"}
// 	log, _ := mocklogger.HijackOutput(logger.New())

// 	// Act
// 	result := connection.Get(conn, log.WithField("test", "test"))
// 	_, ok := result.(*connection.DockerCreateConnection)
// 	// Assert
// 	assert.True(t, ok, "Expected DockerAttachConnection when ContainerName is provided and ImageName is empty")
// }
