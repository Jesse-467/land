package main

import (
	"fmt"
	"land/logger"
	"land/settings"
)

func main() {
	configPath := "./conf/config.yaml"

	if err := settings.Init(configPath); err != nil {
		fmt.Printf("load config failed,err : %v\n", err)
		return
	}

	if err := logger.Init(settings.Conf.LogConfig, settings.Conf.Mode); err != nil {
		fmt.Printf("init logger failed,err : %v\n", err)
		return
	}

}
