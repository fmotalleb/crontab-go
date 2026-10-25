package connection

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestCheckDockerExitStatus(t *testing.T) {
	assert.NoError(t, checkDockerExitStatus("container", 0))
	assert.Equal(t, "container exited with status 7", checkDockerExitStatus("container", 7).Error())
}
