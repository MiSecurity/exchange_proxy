package settings

import (
	"exchange_proxy/logger"
	"gopkg.in/ini.v1"
)

var (
	Cfg *ini.File
)

func init() {
	var err error
	source := "conf/app.ini"
	Cfg, err = ini.Load(source)

	if err != nil {
		logger.Log.Panicln(err)
	}
}
