package service

import (
	"github.com/buzhiyun/go-mysql-replication/config"
	"github.com/buzhiyun/go-mysql-replication/model"
	"github.com/buzhiyun/go-mysql-replication/prometheus"
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/kataras/golog"
)

//type handler struct {
//	queue chan interface{}
//	stop  chan struct{}
//}

type canalHandler struct {
}

func newCanalHandler() *canalHandler {
	return &canalHandler{}
}

func updateTableInfo(schema string, table string) (err error) {
	if hitTable(schema, table) {
		_tableInfo, err := CanalInstance.canal.GetTable(schema, table)
		if err != nil {
			golog.Error(err)
			return err
		}
		tableName := schema + "." + table
		model.TableInfo[tableName] = model.TableInfomation{
			TableInfo:       _tableInfo,
			TableColumnSize: len(_tableInfo.Columns),
		}
	}

	return
}

/*
* 检查表是否命中筛选规则
* 当canal.ExecutionPath 为空时候 ，会读取出所有表的binlog ，需要筛选
 */
func hitTable(schema string, table string) (hit bool) {
	// 默认情况直接命中
	hit = true
	if len(config.SchemaMap) > 0 {
		if _, ok := config.SchemaMap[schema]; ok != true {
			hit = false
		}
	}

	if len(config.TableMap) > 0 {
		if _, ok := config.TableMap[table]; ok != true {
			hit = false
		}
	}

	return
}

func (h *canalHandler) OnRotate(e *replication.RotateEvent) error {
	Msg.queue <- model.PosMsg{
		Name:  string(e.NextLogName),
		Pos:   uint32(e.Position),
		Force: true,
	}
	prometheus.AddTransferMsg()
	return nil
}

func (h *canalHandler) OnTableChanged(schema, table string) error {
	updateTableInfo(schema, table)
	return nil
}

func (h *canalHandler) OnDDL(nextPos mysql.Position, q *replication.QueryEvent) error {
	if !hitTable(string(q.Schema), "") {
		return nil
	}

	Msg.queue <- model.PosMsg{
		Name:  nextPos.Name,
		Pos:   nextPos.Pos,
		Force: true,
	}
	prometheus.AddTransferMsg()
	//golog.Infof("ddl: schema : %s" ,q.Schema)
	//golog.Infof("ddl: query : %s" ,q.Query)

	Msg.queue <- model.DdlMsg{
		Action:        "ddl",
		Schema:        string(q.Schema),
		Query:         string(q.Query),
		ErrorCode:     q.ErrorCode,
		ExecutionTime: q.ExecutionTime,
		StatusVars:    q.StatusVars,
	}
	prometheus.AddTransferMsg()

	Msg.queue <- model.GtidMsg{
		GtidSet: q.GSet.String(),
		Force:   false,
	}
	prometheus.AddTransferMsg()
	return nil
}

func (h *canalHandler) OnXID(nextPos mysql.Position) error {
	Msg.queue <- model.PosMsg{
		Name:  nextPos.Name,
		Pos:   nextPos.Pos,
		Force: true,
	}
	prometheus.AddTransferMsg()
	return nil
}

func (h *canalHandler) OnRow(e *canal.RowsEvent) error {

	if !hitTable(e.Table.Schema, e.Table.Name) {
		return nil
	}

	//golog.Infof("rowevent: ",e.Table.Schema,e.Table.Name,e.Action,e.Rows)
	var requests []*model.RowMsg
	if e.Action != canal.UpdateAction {
		// 定长分配
		requests = make([]*model.RowMsg, 0, len(e.Rows))
	}


	if e.Action == canal.UpdateAction {
		for i := 0; i < len(e.Rows); i++ {
			v := &model.RowMsg{}
			if (i+1)%2 == 0 {
				v.Table = e.Table.Name
				v.Schema = e.Table.Schema
				v.Key = e.Table.Schema + "." + e.Table.Name
				v.Action = e.Action
				v.Timestamp = e.Header.Timestamp
				v.Old = e.Rows[i-1]
				v.Row = e.Rows[i]
				requests = append(requests, v)
			}
		}
	} else {
		for _, row := range e.Rows {
			v := &model.RowMsg{}
			v.Table = e.Table.Name
			v.Schema = e.Table.Schema
			v.Key = e.Table.Schema + "." + e.Table.Name
			v.Action = e.Action
			v.Timestamp = e.Header.Timestamp
			v.Row = row
			//golog.Infof("%s", row)
			requests = append(requests, v)
		}
	}
	Msg.queue <- requests
	prometheus.AddTransferMsg()

	return nil
}

func (h *canalHandler) OnGTID(gtid mysql.GTIDSet) error {
	Msg.queue <- model.GtidMsg{
		GtidSet: gtid.String(),
		Force:   false,
	}
	prometheus.AddTransferMsg()
	return nil
}

func (h *canalHandler) OnPosSynced(pos mysql.Position, set mysql.GTIDSet, force bool) error {
	return nil
}

func (h *canalHandler) String() string {
	return "canal Handler"
}
