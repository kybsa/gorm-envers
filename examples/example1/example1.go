package main

import (
	"github.com/kybsa/gorm-envers/envers"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type A struct {
	ID   uint
	Name string
}

func (A) IsAudit() bool {
	return true
}

type AsAud struct {
	A
	envers.AudData
}

func (AsAud) TableName() string {
	return "as_aud"
}

func main() {
	gormEnversPlugin := envers.NewGormEnversPlugin(envers.NewConfig())
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{
		Plugins: map[string]gorm.Plugin{gormEnversPlugin.Name(): gormEnversPlugin},
	})

	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&A{}, &envers.Revinfo{}, &AsAud{})

	a := &A{Name: "A-1"}
	db.Create(a)
	a.Name = "A-2"
	db.Save(a)
	db.Delete(a)
}
