package envers

import (
	"reflect"
	"time"

	"gorm.io/gorm"
)

type AudData struct {
	Rev         uint
	Revtype     uint8
	Revend      *uint
	RevendTstmp *int64
}

type Revinfo struct {
	Rev      uint `gorm:"primaryKey;autoIncrement:true;column:rev"`
	Revtstmp int64
}

func AfterCreate(tx *gorm.DB, e interface{}) (err error) {
	revtstmp := time.Now().UnixMilli()
	revInfo := Revinfo{Revtstmp: revtstmp}
	result := tx.Create(&revInfo)
	if result.Error != nil {
		return result.Error
	}

	values := reflect.ValueOf(e)
	elem := values.Elem()

	revField := elem.FieldByName("Rev")
	revField.SetUint(uint64(revInfo.Rev))

	revtypeField := elem.FieldByName("Revtype")
	revtypeField.SetUint(uint64(Add))

	result = tx.Create(e)

	return result.Error
}

func AfterUpdate(tx *gorm.DB, emptyModel interface{}, columnIdName string, id interface{}, e interface{}) (err error) {
	revtstmp := time.Now().UnixMilli()
	revInfo := Revinfo{Revtstmp: revtstmp}
	result := tx.Create(&revInfo)
	if result.Error != nil {
		return result.Error
	}

	// TODO Remove column names hard code
	result = tx.Model(emptyModel).Where(columnIdName+"= ? and revend is null", id).Updates(map[string]interface{}{"revend": revInfo.Rev, "revend_tstmp": revtstmp})
	if result.Error != nil {
		return result.Error
	}

	values := reflect.ValueOf(e)
	elem := values.Elem()
	revField := elem.FieldByName("Rev")
	revField.SetUint(uint64(revInfo.Rev))

	revtypeField := elem.FieldByName("Revtype")
	revtypeField.SetUint(uint64(Update))

	result = tx.Create(e)

	return result.Error // TODO
}
func AfterDelete(tx *gorm.DB, emptyModel interface{}, columnIdName string, id interface{}, e interface{}) (err error) {
	revtstmp := time.Now().UnixMilli()
	revInfo := Revinfo{Revtstmp: revtstmp}
	result := tx.Create(&revInfo)
	if result.Error != nil {
		return result.Error
	}

	// TODO Remove column names hard code
	result = tx.Model(emptyModel).Where(columnIdName+"= ? and revend is null", id).Updates(map[string]interface{}{"revend": revInfo.Rev, "revend_tstmp": revtstmp})
	if result.Error != nil {
		return result.Error
	}

	values := reflect.ValueOf(e)
	elem := values.Elem()
	revField := elem.FieldByName("Rev")
	revField.SetUint(uint64(revInfo.Rev))

	revtypeField := elem.FieldByName("Revtype")
	revtypeField.SetUint(uint64(Del))

	result = tx.Create(e)

	return result.Error
}
