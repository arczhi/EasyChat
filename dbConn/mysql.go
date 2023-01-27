package dbConn

import (
	"EasyChat/config"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewMySQLConn() *gorm.DB {
	//连接mysql
	dsn := config.Cfg.MYSQL.Username + ":" + config.Cfg.MYSQL.Password + "@tcp(" + config.Cfg.MYSQL.Addr + ")/" + config.Cfg.MYSQL.Db + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	//维护连接池
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	//设置连接可复用的最大时间，与mysql的默认设置时间保持一致
	sqlDB.SetConnMaxLifetime(8 * time.Hour)

	return db
}
