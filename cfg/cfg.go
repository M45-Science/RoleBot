package cfg

import (
	"RoleBot/cons"
	"RoleBot/cwlog"
	"encoding/json"
	"os"
)

var Config serverConfig

type serverConfig struct {
	Token  string
	App    string
	Domain string
}

func WriteCfg() bool {
	tempPath := cons.ConfigFile + ".tmp"
	finalPath := cons.ConfigFile

	encBuf, err := json.MarshalIndent(Config, "", "\t")
	if err != nil {
		cwlog.DoLog("WriteCfg: marshal failure")
		return false
	}
	outbuf := append(encBuf, '\n')

	err = os.WriteFile(tempPath, outbuf, 0600)

	if err != nil {
		cwlog.DoLog("WriteCfg: WriteFile failure")
	}

	err = os.Rename(tempPath, finalPath)

	if err != nil {
		cwlog.DoLog("WriteCfg: Couldn't rename cfg file.")
		return false
	}

	return true
}

func ReadCfg() bool {

	_, err := os.Stat(cons.ConfigFile)
	notfound := os.IsNotExist(err)

	if notfound {
		cwlog.DoLog("ReadCfg: os.Stat failed, empty config generated.")
		return true
	} else { /* Otherwise just read in the config */
		file, err := os.ReadFile(cons.ConfigFile)

		if file != nil && err == nil {
			newcfg := serverConfig{}

			err := json.Unmarshal([]byte(file), &newcfg)
			if err != nil {
				cwlog.DoLog("ReadCfg: Unmarshal failure")
				cwlog.DoLog(err.Error())
				return false
			}

			Config = newcfg
			return true
		} else {
			cwlog.DoLog("ReadCfg: ReadFile failure")
			return false
		}
	}
}
