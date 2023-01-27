package models

import (
	"EasyChat/dbConn"

	"gorm.io/gorm"
)

var DB *gorm.DB

func init() {
	DB = dbConn.NewMySQLConn()
}
