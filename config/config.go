package config

import (
	"github.com/amocea/go-im-chat/defs"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

var db *xorm.Engine

func init() {
	dns := os.Getenv("MYSQL_DRIVER_URL")
	if dns == "" {
		dns = "root:123456@tcp(127.0.0.1:3306)/im_db?charset=utf8&parseTime=true&loc=Local"
	}
	var err error
	db, err = xorm.NewEngine("mysql", dns)
	if err != nil {
		log.Fatalf("Connecting database driver[mysql] failed: %v\n", err)
	}

	// 设置显示 sql 语句
	db.ShowSQL(true)
	// 设置数据库最大打开的数据库连接数
	db.SetMaxOpenConns(2)
	// 设置创建表结构时为复数形式
	db.SetMapper(names.SnakeMapper{})
	db.SetTableMapper(names.SnakeMapper{})
	db.SetColumnMapper(names.SnakeMapper{})
	_ = db.Sync2(&defs.User{}, &defs.Contact{}, &defs.Community{}) // 自动创建表结构

	log.Println("初始化数据库成功... ")
}

// DB 返回数据库
func DB() *xorm.Engine {
	return db
}
