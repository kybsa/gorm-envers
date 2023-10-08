// Package envers with config utils
package envers

func NewConfig() config {
	return config{
		ShowSQL:               false,
		RevinfoTableName:      "revinfo",
		ShowDebugInfo:         false,
		RevtstmpColumnName:    "revtstmp",
		RevColumnName:         "rev",
		RevtypeColumnName:     "revtype",
		RevendColumnName:      "revend",
		RevendTstmpColumnName: "revend_tstmp",
		AuditTableSuffix:      "_aud",
	}
}
