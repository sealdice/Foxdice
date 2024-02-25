package main

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(driver, dsn string) (*gorm.DB, error) {
	var dia gorm.Dialector
	switch driver {
	case "sqlite":
		dia = sqlite.Open(dsn)
	case "postgres":
		dia = postgres.Open(dsn)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动:%s", driver)
	}
	return gorm.Open(dia)
}

var s = sessionStore{}

type sessionStore struct{}

func (s *sessionStore) Load(id string) map[string]any {
	type VarData struct {
		Gid  string         `gorm:"primarykey"`
		Data map[string]any `gorm:"serializer:json"`
	}
	return map[string]any{}
}

func (s *sessionStore) Store(id string, m map[string]any) {

}
