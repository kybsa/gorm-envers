package envers

import (
	"fmt"

	"gorm.io/gorm"
)

type gormEnversPlugin struct {
}

func NewGormEnversPlugin() gorm.Plugin {
	return &gormEnversPlugin{}
}

func (_self *gormEnversPlugin) Initialize(db *gorm.DB) error {
	fmt.Println("Start: Initialize")

	// TODO Check if a entity support aud table

	db.Callback().Create().After("gorm:create").Register("create_at", _self.create)
	db.Callback().Update().After("gorm:update").Register("update_at", _self.update)
	db.Callback().Delete().After("gorm:delete").Register("delete_at", _self.delete)

	fmt.Println("End  : Initialize")

	return nil
}

func (_self *gormEnversPlugin) Name() string {
	return "GORM-ENVERS"
}
func (_self *gormEnversPlugin) create(db *gorm.DB) {
	createSQLAud(db, Add)
}

func (_self *gormEnversPlugin) update(db *gorm.DB) {
	createSQLAud(db, Update)
}

func (_self *gormEnversPlugin) delete(db *gorm.DB) {
	createSQLAud(db, Del)
}
