package settings

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	Conf = new(AppConfig)
)

type AppConfig struct {
	Name      string `mapstructure:"name"`       // 应用名称
	Mode      string `mapstructure:"mode"`       // 运行模式
	Version   string `mapstructure:"version"`    // 版本号
	StartTime string `mapstructure:"start_time"` // 启动时间
	MachineID int64  `mapstructure:"machine_id"` // 机器ID
	Port      int    `mapstructure:"port"`       // 端口号

	*LogConfig   `mapstructure:"log"`   // 日志配置
	*MysqlConfig `mapstructure:"mysql"` // mysql配置
	*RedisConfig `mapstructure:"redis"` // redis配置
}

type LogConfig struct {
	Level      string `mapstructure:"level"`       // 日志级别
	Filename   string `mapstructure:"filename"`    // 日志文件名
	MaxSize    int    `mapstructure:"max_size"`    // 日志文件最大大小
	MaxAge     int    `mapstructure:"max_age"`     // 日志文件最大保存时间
	MaxBackups int    `mapstructure:"max_backups"` // 日志文件最大备份数

}

type MysqlConfig struct {
	User         string `mapstructure:"user"`           // mysql用户名
	PassWord     string `mapstructure:"password"`       // mysql密码
	Host         string `mapstructure:"host"`           // mysql主机地址
	DBName       string `mapstructure:"dbname"`         // mysql数据库名
	Port         int    `mapstructure:"port"`           // mysql主机端口
	MaxOpenConns int    `mapstructure:"max_open_conns"` // 最大连接数
	MaxIdleConns int    `mapstructure:"max_idle_conns"` // 最大空闲连接数
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`           // redis主机地址
	PassWord     string `mapstructure:"password"`       // redis密码
	Port         int    `mapstructure:"port"`           // redis主机端口
	DB           int    `mapstructure:"db"`             // redis数据库
	PoolSize     int    `mapstructure:"pool_size"`      // redis连接池大小
	MinIdleConns int    `mapstructure:"min_idle_conns"` // 最小空闲连接数
}

func Init(filePath string) error {
	// 读取并初始化配置
	viper.SetConfigFile(filePath)

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("viper.ReadInConfig() failed, err:%v\n", err)
		return err
	}

	if err := viper.Unmarshal(Conf); err != nil {
		log.Printf("vip.Unmarshal() failed, err:%v\n", err)
		return err
	}

	// 监听配置文件变化
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		log.Println("配置文件发生变化")
		if err := viper.Unmarshal(Conf); err != nil {
			log.Printf("vip.Unmarshal() failed, err:%v\n", err)
		}
	})

	fmt.Println("settings init success")

	return nil
}
