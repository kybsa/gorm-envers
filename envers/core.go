package envers

const Add = 0
const Update = 1
const Del = 2

type Audited interface {
	IsAudit() bool
}

type config struct {
	ShowSQL               bool
	ShowDebugInfo         bool
	RevinfoTableName      string
	RevtstmpColumnName    string
	RevColumnName         string
	RevtypeColumnName     string
	RevendColumnName      string
	RevendTstmpColumnName string
	AuditTableSuffix      string
}
