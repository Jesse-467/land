package main

import (
	"fmt"
	"land/dao/mysql"
	"land/dao/redis"
	"land/logger"
	"land/logic"
	"land/pkg/snowflake"
	"land/routers"
	"land/settings"
	"os"
	"time"
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

	if err := snowflake.Init("2024-06-07", 1); err != nil {
		fmt.Printf("init snowflake failed,err : %v\n", err)
		return
	}

	// 启动访问量同步服务
	syncService := logic.NewViewCountSyncService(5 * time.Minute) // 每5分钟同步一次
	syncService.Start()
	defer syncService.Stop()

	// 启动路由
	r := routers.SetRouter(settings.Conf.Mode)
	err := r.Run(fmt.Sprintf(":%d", settings.Conf.Port))
	if err != nil {
		fmt.Printf("run server failed,err : %v\n", err)
		return
	}
}
