package envers

import (
	"fmt"
	"reflect"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const Add = 0
const Update = 1
const Del = 2

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

type GormEnversPlugin struct {
}

func (_self *GormEnversPlugin) Name() string {
	return "GORM-ENVERS"
}
func (_self *GormEnversPlugin) Create(db *gorm.DB) {
	_self.CreateAud(db, Add)
}

func (_self *GormEnversPlugin) Update(db *gorm.DB) {
	_self.CreateAud(db, Update)
}

func (_self *GormEnversPlugin) Delete(db *gorm.DB) {
	_self.CreateAud(db, Del)
}

func (_self *GormEnversPlugin) CreateAud(db *gorm.DB, audType int) {
	if db.Statement.Schema == nil {
		return
	}

	fmt.Println("Create Table:", db.Statement.Schema.Table)

	revtstmp := time.Now().UnixMilli()
	var rev int64
	ndb, _ := db.DB()
	result, errSql := ndb.Query("insert into revinfo (revtstmp) VALUES(" + fmt.Sprint(revtstmp) + ") RETURNING rev") // tabla quemada
	if errSql != nil {
		fmt.Println(errSql)
		return
	}
	result.Next()
	result.Scan(&rev)

	fmt.Println("Table:", db.Statement.Schema.Table, " ModelType:", db.Statement.Schema.ModelType)

	tableName := db.Statement.Schema.Table + "_aud" // TODO Remove suffix hard code

	if audType == Del || audType == Update {
		updateRevEndFields(db, tableName, rev, revtstmp)
	}

	sqlTable := "insert into " + tableName
	sqlColumns := createSqlColumns(db.Statement.Schema.Fields)
	sqlValues := createSqlValues(db)
	values := createValues(db, rev, audType)
	sql := sqlTable + sqlColumns + sqlValues
	fmt.Println(sql)
	result, errAud := ndb.Query(sql, values...)
	fmt.Println(errAud)
	fmt.Println(result)
}

func getNumberOfTuples(db *gorm.DB) int {
	if db.Statement.ReflectValue.Kind() == reflect.Array || db.Statement.ReflectValue.Kind() == reflect.Slice {
		return db.Statement.ReflectValue.Len()
	}
	return 1
}

func createSqlColumns(fields []*schema.Field) string {
	sqlColumns := " ("
	for _, field := range fields {
		sqlColumns += field.DBName
		sqlColumns += ","
	}
	sqlColumns += "rev,revtype) " // Remove column names hard code
	return sqlColumns
}

func createSqlValues(db *gorm.DB) string {
	tuplasSize := getNumberOfTuples(db)
	numFields := len(db.Statement.Schema.Fields)
	sqlValues := " VALUES"

	for i := 0; i < tuplasSize; i++ {
		sqlValues += "("
		indexField := i*(numFields+2) + 1
		limit := indexField + numFields
		for ; indexField < limit; indexField++ {
			sqlValues += "$" + fmt.Sprint(indexField) + ","
		}
		sqlValues += "$" + fmt.Sprint(indexField) + ",$" + fmt.Sprint(indexField+1) + ")"
		if i < tuplasSize-1 {
			sqlValues += ","
		}
	}
	return sqlValues
}

func createValues(db *gorm.DB, rev int64, audType int) []interface{} {
	tuplasSize := getNumberOfTuples(db)
	numFields := len(db.Statement.Schema.Fields) + 2
	values := make([]interface{}, 0, tuplasSize*numFields)

	if db.Statement.ReflectValue.Kind() == reflect.Array || db.Statement.ReflectValue.Kind() == reflect.Slice {
		for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
			for _, field := range db.Statement.Schema.Fields {
				if fieldValue, isZero := field.ValueOf(db.Statement.Context, db.Statement.ReflectValue.Index(i)); !isZero {
					values = append(values, fieldValue)
				}
			}
			values = append(values, rev)
			values = append(values, audType)
		}
	} else {
		for _, field := range db.Statement.Schema.Fields {
			if fieldValue, isZero := field.ValueOf(db.Statement.Context, db.Statement.ReflectValue); !isZero {
				values = append(values, fieldValue)
			}
		}
		values = append(values, rev)
		values = append(values, audType)
	}

	return values
}

func updateRevEndFields(db *gorm.DB, tableName string, revend int64, revtstmp int64) {
	ndb, errDb := db.DB()
	if errDb != nil {
		fmt.Println(errDb)
		return
	}

	values := make([]interface{}, 0, 2)
	values = append(values, revend, revtstmp)

	// Get where pk
	indexId := 3
	wherePk := ""
	tuplasSize := getNumberOfTuples(db)

	if tuplasSize == 1 {
		for _, field := range db.Statement.Schema.Fields {
			if field.PrimaryKey {
				if len(wherePk) > 0 {
					wherePk += " and "
				}
				wherePk += field.DBName
				wherePk += "=$" + fmt.Sprint(indexId)
				indexId++

				if fieldValue, isZero := field.ValueOf(db.Statement.Context, db.Statement.ReflectValue); !isZero {
					values = append(values, fieldValue)
				}
			}
		}
	} else {
		for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
			wherePk += "("
			for _, field := range db.Statement.Schema.Fields {
				wherePkEntity := ""
				if field.PrimaryKey {
					if len(wherePkEntity) > 0 {
						wherePkEntity += " and "
					}
					wherePkEntity += field.DBName
					wherePkEntity += "=$" + fmt.Sprint(indexId)
					indexId++

					if fieldValue, isZero := field.ValueOf(db.Statement.Context, db.Statement.ReflectValue.Index(i)); !isZero {
						values = append(values, fieldValue)
					}
				}
				wherePk += wherePkEntity
			}
			wherePk += ")"
			if i < db.Statement.ReflectValue.Len()-1 {
				wherePk += " or "
			}
		}
	}
	_, errUpdate := ndb.Query("update "+tableName+" set revend=$1, revend_tstmp=$2 where "+wherePk+" and revend is null", values...) // Remove columns hardcode
	if errUpdate != nil {
		fmt.Println(errUpdate)
	}
}

func (_self *GormEnversPlugin) Initialize(db *gorm.DB) error {
	fmt.Println("Start: Initialize")

	db.Callback().Create().After("gorm:create").Register("create_at", _self.Create)
	db.Callback().Update().After("gorm:update").Register("update_at", _self.Update)
	db.Callback().Delete().After("gorm:delete").Register("delete_at", _self.Delete)

	fmt.Println("End  : Initialize")

	return nil
}
