package connection

import "fmt"

// checkDockerExitStatus returns nil for status zero, otherwise an error naming
// subject and its exit status.
func checkDockerExitStatus(subject string, status int64) error {
	if status != 0 {
		return fmt.Errorf("%s exited with status %d", subject, status)
	}
	return nil
}
