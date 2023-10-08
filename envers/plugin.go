package envers

import (
	"gorm.io/gorm"
)

type gormEnversPlugin struct {
	Config config
	sqlAud sqlAud
}

func NewGormEnversPlugin(config config) gorm.Plugin {
	return &gormEnversPlugin{Config: config, sqlAud: *NewSQLAud(config)}
}

func (_self *gormEnversPlugin) Initialize(db *gorm.DB) error {

	if _self.Config.ShowDebugInfo {
		db.Config.Logger.Info(db.Statement.Context, "Start Initialize:"+_self.Name())
	}

	err := db.Callback().Create().After("gorm:create").Register(_self.Name()+".create_at", _self.create)
	if err != nil {
		db.Config.Logger.Error(db.Statement.Context, _self.Name()+" fail to register create callback")
		return err
	}

	err = db.Callback().Update().After("gorm:update").Register(_self.Name()+".update_at", _self.update)
	if err != nil {
		db.Config.Logger.Error(db.Statement.Context, _self.Name()+" fail to register update callback")
		return err
	}

	err = db.Callback().Delete().After("gorm:delete").Register(_self.Name()+".delete_at", _self.delete)
	if err != nil {
		db.Config.Logger.Error(db.Statement.Context, _self.Name()+" fail to register delete callback")
		return err
	}

	if _self.Config.ShowDebugInfo {
		db.Config.Logger.Info(db.Statement.Context, "Success Initialize:"+_self.Name())
	}

	return nil
}

func (_self *gormEnversPlugin) Name() string {
	return "GORM-ENVERS"
}
func (_self *gormEnversPlugin) create(db *gorm.DB) {
	_self.sqlAud.createSQLAud(db, Add)
}

func (_self *gormEnversPlugin) update(db *gorm.DB) {
	_self.sqlAud.createSQLAud(db, Update)
}

func (_self *gormEnversPlugin) delete(db *gorm.DB) {
	_self.sqlAud.createSQLAud(db, Del)
}
