package main

import (
	"fmt"
	"land/dao/mysql"
	"land/dao/redis"
	"land/logger"
	"land/settings"
	"os"
)

func main() {
	configPath := "./conf/config.yaml"

	// 可以使用命令行参数指定配置文件路径
	if len(os.Args) == 2 {
		configPath = os.Args[1]
	}

	if err := settings.Init(configPath); err != nil {
		fmt.Printf("load config failed,err : %v\n", err)
		return
	}

	if err := logger.Init(settings.Conf.LogConfig, settings.Conf.Mode); err != nil {
		fmt.Printf("init logger failed,err : %v\n", err)
		return
	}

	if err := mysql.Init(settings.Conf.MysqlConfig); err != nil {
		fmt.Printf("init mysql failed,err : %v\n", err)
		return
	}
	defer mysql.Close()

	if err := redis.Init(settings.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed,err : %v\n", err)
		return

	}
	defer redis.Close()
}
