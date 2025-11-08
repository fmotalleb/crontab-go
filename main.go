/*
Copyright © 2024 Motalleb Fallahnezhad

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	"log"
	"os"

	"github.com/robfig/cron/v3"

	"github.com/FMotalleb/crontab-go/cmd"
	"github.com/FMotalleb/crontab-go/core/global"
	"github.com/FMotalleb/crontab-go/core/jobs"
	"github.com/FMotalleb/crontab-go/core/webserver"
	"github.com/FMotalleb/crontab-go/meta"
)

func initializeGlobalState() {
	cronInstance := cron.New(cron.WithSeconds())
	global.Put(cronInstance)
	cronInstance.Start()
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf(
				"an error stopped application from working, if you think this is an error in application side please report to %s\nError: %v",
				meta.Issues(),
				err,
			)
			os.Exit(1)
		}
	}()

	cmd.Execute()

	initializeGlobalState()
	l := global.Logger("cron")
	l.Info("Booting up")
	jobs.InitializeJobs()
	if cmd.CFG.WebServerAddress != "" {
		go webserver.
			NewWebServer(
				global.CTX(),
				cmd.CFG.WebServerAddress,
				cmd.CFG.WebServerPort,

				cmd.CFG.WebServerMetrics,
				&webserver.AuthConfig{
					Username: cmd.CFG.WebserverUsername,
					Password: cmd.CFG.WebServerPassword,
				},
			).
			Serve()
	}
	<-global.CTX().Done()
}
