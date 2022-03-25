package prometheus

import (
	"github.com/buzhiyun/go-mysql-replication/model"
)

var (
	insertRecord map[string]uint64
	updateRecord map[string]uint64
	deleteRecord map[string]uint64
	ddlRecord    map[string]uint64
)

type TableReport struct {
	InsertCount uint64 `json:"insert_count"`
	UpdateCount uint64 `json:"update_count"`
	DeleteCount uint64 `json:"delete_count"`
}

type DdlReport map[string]struct {
	DdlCount uint64 `json:"ddl_count"`
}

func GetTableReport() interface{} {
	tableReport := make(map[string]TableReport, len(model.TableInfo))
	for _, infomation := range model.TableInfo {
		key := infomation.TableInfo.Schema + "." + infomation.TableInfo.Name
		tableReport[key] = TableReport{
			InsertCount: insertRecord[key],
			UpdateCount: updateRecord[key],
			DeleteCount: deleteRecord[key],
		}
	}
	return tableReport
}

func GetSchmaDdlReport() interface{} {
	return ddlRecord
}
