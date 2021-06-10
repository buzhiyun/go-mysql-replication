package model

//var RowRequestPool = sync.Pool{
//	New: func() interface{} {
//		return new(RowRequest)
//	},
//}

type RowMsg struct {
	Schema    string
	Table     string
	Key       string //  db.table
	Action    string
	Timestamp uint32
	Old       []interface{}
	Row       []interface{}
}

type PosMsg struct {
	Name  string
	Pos   uint32
	Force bool
}

type GtidMsg struct {
	GtidSet string
	Force   bool
}

type DdlMsg struct {
	Action        string
	Schema        string
	Query         string
	ErrorCode     uint16
	ExecutionTime uint32
	StatusVars    []byte
}
