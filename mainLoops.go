package main

import (
	"RoleBot/cwlog"
	"RoleBot/glob"
	"os"
	"time"
)

func MainLoop() {

	/* Reconnect log descriptor */
	go func() {

		for glob.ServerRunning {
			time.Sleep(time.Second * 5)

			var err error
			if _, err = os.Stat(glob.LogName); err != nil {

				glob.LogDesc.Close()
				glob.LogDesc = nil
				cwlog.StartLog()
				cwlog.DoLog("Log file was deleted, recreated.")
			}
		}
	}()
}
