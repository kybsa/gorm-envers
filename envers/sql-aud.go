package envers

import (
	"fmt"
	"reflect"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var auditedType = reflect.TypeOf((*Audited)(nil)).Elem()

type sqlAud struct {
	config config
}

func NewSQLAud(config config) *sqlAud {
	return &sqlAud{config: config}
}

func (_self *sqlAud) createSQLAud(db *gorm.DB, audType int) {
	if db.Statement.Schema == nil {
		return
	}

	modelType := db.Statement.Schema.ModelType
	if !modelType.Implements(auditedType) {
		return
	}

	if _self.config.ShowDebugInfo {
		db.Config.Logger.Info(db.Statement.Context, "Start creating audit data to table:{%s} model:{%s.%s}", db.Statement.Schema.Table, db.Statement.Schema.ModelType.PkgPath(), db.Statement.Schema.ModelType.Name())
	}

	revtstmp := time.Now().UnixMilli()
	var rev int64
	ndb, _ := db.DB()
	sqlInsertRevInfo := "insert into " + _self.config.RevinfoTableName + " (" + _self.config.RevtstmpColumnName + ") VALUES(" + fmt.Sprint(revtstmp) + ") RETURNING " + _self.config.RevColumnName
	if _self.config.ShowSQL {
		db.Config.Logger.Info(db.Statement.Context, sqlInsertRevInfo)
	}
	result, errSQL := ndb.Query(sqlInsertRevInfo)
	if errSQL != nil {
		db.Config.Logger.Error(db.Statement.Context, "fail to insert revinfo data", errSQL)
		return
	}
	result.Next()
	result.Scan(&rev)

	tableName := db.Statement.Schema.Table + _self.config.AuditTableSuffix

	if audType == Del || audType == Update {
		_self.updateRevEndFields(db, tableName, rev, revtstmp)
	}

	sqlTable := "insert into " + tableName
	sqlColumns := _self.createSQLColumns(db.Statement.Schema.Fields)
	sqlValues := _self.createSQLValues(db)
	values := _self.createValues(db, rev, audType)
	sql := sqlTable + sqlColumns + sqlValues
	if _self.config.ShowSQL {
		db.Config.Logger.Info(db.Statement.Context, sql)
	}
	_, errAud := ndb.Query(sql, values...)
	if errAud != nil {
		db.Config.Logger.Error(db.Statement.Context, "fail to insert audit data", errAud.Error())
	}
}

func (_self *sqlAud) getNumberOfTuples(db *gorm.DB) int {
	if db.Statement.ReflectValue.Kind() == reflect.Array || db.Statement.ReflectValue.Kind() == reflect.Slice {
		return db.Statement.ReflectValue.Len()
	}
	return 1
}

func (_self *sqlAud) createSQLColumns(fields []*schema.Field) string {
	sqlColumns := " ("
	for _, field := range fields {
		sqlColumns += field.DBName
		sqlColumns += ","
	}
	sqlColumns += _self.config.RevColumnName + "," + _self.config.RevtypeColumnName + ") "
	return sqlColumns
}

func (_self *sqlAud) createSQLValues(db *gorm.DB) string {
	tuplasSize := _self.getNumberOfTuples(db)
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

func (_self *sqlAud) createValues(db *gorm.DB, rev int64, audType int) []interface{} {
	tuplasSize := _self.getNumberOfTuples(db)
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

func (_self *sqlAud) updateRevEndFields(db *gorm.DB, tableName string, revend int64, revtstmp int64) {
	ndb, errDB := db.DB()
	if errDB != nil {
		db.Config.Logger.Error(db.Statement.Context, "fail to get Db instance", errDB.Error())
		return
	}

	values := make([]interface{}, 0, 2)
	values = append(values, revend, revtstmp)

	// Get where pk
	indexID := 3
	wherePk := ""
	tuplasSize := _self.getNumberOfTuples(db)

	if tuplasSize == 1 {
		for _, field := range db.Statement.Schema.Fields {
			if field.PrimaryKey {
				if len(wherePk) > 0 {
					wherePk += " and "
				}
				wherePk += field.DBName
				wherePk += "=$" + fmt.Sprint(indexID)
				indexID++

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
					wherePkEntity += "=$" + fmt.Sprint(indexID)
					indexID++

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
	sqlUpdate := fmt.Sprintf("update %s set %s=$1, %s=$2 where %s and %s is null",
		tableName,
		_self.config.RevendColumnName,
		_self.config.RevendTstmpColumnName,
		wherePk,
		_self.config.RevendColumnName)
	_, errUpdate := ndb.Query(sqlUpdate, values...)
	if errUpdate != nil {
		db.Config.Logger.Error(db.Statement.Context, "fail to update revend fields", errDB.Error())
	}
}
