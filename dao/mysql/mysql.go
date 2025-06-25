package mysql

import (
	"fmt"
	"land/settings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func Init(cfg *settings.MysqlConfig) (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local", cfg.User, cfg.PassWord, cfg.Host, cfg.Port, cfg.DBName)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("mysql connect error:%v\n", err)
		return err
	}

	mysqlDB, err := db.DB()
	if err != nil {
		fmt.Printf("get mysql db error:%v\n", err)
		return err
	}

	mysqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	mysqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	fmt.Println("mysql connect success: ", db == nil)

	return nil
}

func Close() {
	if db == nil {
		return
	}
	mysqlDB, err := db.DB()
	if err != nil {
		fmt.Printf("get mysql db error:%v\n", err)
		return
	}
	_ = mysqlDB.Close()
}
