package envers

const Add = 0
const Update = 1
const Del = 2

type Audited interface {
	IsAudit() bool
}
