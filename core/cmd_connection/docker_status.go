package connection

import "fmt"

func checkDockerExitStatus(subject string, status int64) error {
	if status != 0 {
		return fmt.Errorf("%s exited with status %d", subject, status)
	}
	return nil
}
