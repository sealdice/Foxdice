package trpg

import (
	"foxdice/core"

	"gorm.io/gorm"
)

type GameLog struct {
	gorm.Model
	Name    string
	GroupId string
	Size    int
}

type Paragraph struct {
	gorm.Model
	Nickname  string
	IMUserId  string
	Time      int64
	Message   string
	IsDice    bool
	CommandId int64
	UniformId string
	Channel   string
}

func BuildReport(m *core.Manager, rdb *gorm.DB) {
	var err error
	err = rdb.AutoMigrate(&GameLog{})
	err = rdb.AutoMigrate(&Paragraph{})
	if err != nil {

	}
}
