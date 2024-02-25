package main

import (
	"context"
	"fmt"
	"foxdice/core"
	"foxdice/extensions/admin"
	"foxdice/extensions/draw"
	"foxdice/extensions/trpg"
	"foxdice/serve"
	"foxdice/utils"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"gorm.io/gorm"
)

func main() {
	k := koanf.New(".")
	err := k.Load(file.Provider("mock/mock.json"), toml.Parser())
	if err != nil {
		fmt.Println()
		panic(err)
		return
	}
	log := NewLogger(&LoggerOption{})
	db, _ := Connect(k.MustString("db.driver"), k.String("dsn"))

	// 连接机器人
	m := core.InitManager(koanf.New("core"), log)
	core.DefaultHandler()
	admin.Build(m, db)
	trpg.BuildCOC7(m)
	trpg.BuildDND5e(m)
	trpg.BuildSR5e(m)
	draw.Build(m)
	{
		var rdb *gorm.DB
		driver := m.Config.String("db.game.log.driver")
		if driver == "sqlite" {
			dsn := fmt.Sprintf("%s?_pragma=auto_vacuum(1)", utils.DataDir("gameLog.db"))
			log.Infof("使用 sqlite driver, 将使用独立数据库, DSN: %s", dsn)
			rdb, err = Connect(driver, dsn)
		} else {
			rdb = db
		}
		trpg.BuildReport(m, rdb)
	}

	go core.Serve(context.TODO())

	// 服务启动
	serve.Serve(koanf.New("serve"), log)
}
