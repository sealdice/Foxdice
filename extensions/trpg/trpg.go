package trpg

import (
	"sync"
	"time"

	ds "github.com/sealdice/dicescript"
	"gorm.io/gorm"
)

type Modeler interface {
	GetData() map[string]ds.VMValue
}

type Model struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt        `gorm:"index"`
	Data      map[string]ds.VMValue `gorm:"serializer:json"`
}

func (m *Model) GetData() map[string]ds.VMValue {
	return m.Data
}

type Character struct {
	Model
	Type     string
	MasterId string
}

type Player struct {
	Model
	Uid    string
	Global uint
	PcList []Character
}

type Team struct {
	name string
	team map[string]Character
}

type Game struct {
	Model
	mu            sync.Mutex
	Uid           string
	teams         map[string]*Team
	ActiveExtList []string
	VM            *ds.Context
}
