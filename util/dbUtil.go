package util

import (
	"Gooj/config"
	"Gooj/logger"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)
var database *gorm.DB
func ConnectDB() {
	conf := config.GetDBSettings()
	connPath := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		conf.DatabaseSettings.User, conf.DatabaseSettings.Password,
		conf.DatabaseSettings.Address, conf.DatabaseSettings.Port,
		conf.DatabaseSettings.DBName)
	db, err := gorm.Open(mysql.Open(connPath), &gorm.Config{})
	if err != nil {
		logger.Fatal("数据库出错："+err.Error())
	}
	database = db
	logger.Info("数据库已连接！")
}
func GetConn()*gorm.DB{
	return database
}
