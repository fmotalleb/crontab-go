// Package meta contains meta data information about this program.
package meta

import "fmt"

func GHUserName() string {
	return "fmotalleb"
}

func GHProjectName() string {
	return "crontab-go"
}

func Project() string {
	return fmt.Sprintf("https://github.com/%s/%s", GHUserName(), GHProjectName())
}

func Issues() string {
	return Project() + "/issues"
}
