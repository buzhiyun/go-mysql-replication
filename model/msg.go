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
	Action        string `json:"action"`
	Schema        string `json:"schema"`
	Query         string `json:"query"`
	ErrorCode     uint16 `json:"error_code"`
	ExecutionTime uint32 `json:"execution_time"`
	StatusVars    []byte `json:"status_vars"`
}
