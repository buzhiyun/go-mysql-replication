package service

import (
	"fmt"
	"github.com/buzhiyun/go-mysql-replication/config"
	"github.com/buzhiyun/go-mysql-replication/message"
	"github.com/buzhiyun/go-mysql-replication/model"
	"github.com/buzhiyun/go-mysql-replication/prometheus"
	"github.com/buzhiyun/go-mysql-replication/service/endpoint"
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/schema"
	"github.com/json-iterator/go"
	"net"
	"os"
	"strings"

	//"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/kataras/golog"
	"time"
)

/*
* 把 msg.queue 中的消息往 mq 里去塞
 */

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var GTID string
var Postion model.PosMsg

type transfer struct {
	ticker         *time.Ticker
	endpoint       endpoint.Endpoint
	endpointEnable bool
	loopStopSignal chan struct{} // 健康检查
	transferEnable bool
	hostname       string // 监控告警用的信息
	ipAddress      string
}

func (t *transfer) GetEndpointInfo() string {
	return t.endpoint.String()
}

func (t *transfer) GetEndpointState() bool {
	return t.endpointEnable
}

func (t *transfer) GetTransferState() bool {
	return t.transferEnable
}

func rowMap(req *model.RowMsg, tableinfo *schema.Table) map[string]interface{} {
	kv := make(map[string]interface{}, len(tableinfo.Columns))

	for i, column := range tableinfo.Columns {
		switch t := req.Row[i].(type) {
		case uint64:
			kv[column.Name] = t
		case string:
			kv[column.Name] = t
		case []byte:
			kv[column.Name] = string(t)
		default:
			kv[column.Name] = t
		}
	}
	return kv
}

func oldRowMap(req *model.RowMsg, tableinfo *schema.Table) map[string]interface{} {
	kv := make(map[string]interface{}, len(tableinfo.Columns))

	for i, column := range tableinfo.Columns {
		//golog.Infof("old: %v" ,req.Old)
		//golog.Infof("table clolumes: %v" ,tableinfo.Columns)
		switch t := req.Old[i].(type) {
		case uint64:
			kv[column.Name] = t
		case string:
			kv[column.Name] = t
		case []byte:
			kv[column.Name] = string(t)
		default:
			kv[column.Name] = t
		}
	}
	return kv
}

func getIpAddr() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	var addrList []string
	for _, addr := range addrs {
		address := addr.String()
		if strings.HasPrefix(address, "10.") || strings.HasPrefix(address, "192.") ||
			strings.HasPrefix(address, "172.") || strings.HasPrefix(address, "100.") {
			addrList = append(addrList, address)
		}
	}
	return strings.Join(addrList, ",")
}

func (t *transfer) Start() (msg string) {
	t.hostname, _ = os.Hostname()
	t.ipAddress = getIpAddr()
	if t.transferEnable {
		return "已经有transfer存活，无需再次启动"
	}
	//var schema string  // 给报警的时候提示是哪一个库用
	// 开启传输线程
	go func() {
		golog.Infof("transfer start")
		t.endpoint = endpoint.NewEndpoint()
		if err := t.endpoint.Connect(); err != nil {
			golog.Error("endpoint 连接失败 :", err)
			return
		}
		t.endpointEnable = true
		prometheus.SetEndpointState(1)

		interval := time.Duration(config.GlobalConfig.FlushBulkInterval)
		bulkSize := config.GlobalConfig.BulkSize
		t.ticker = time.NewTicker(time.Millisecond * interval)
		defer t.Stop()
		var needFlush, needSavePos bool

		lastSavedTime := time.Now()
		rowMsgs := make([]*model.RowMsg, 0, bulkSize)
		t.transferEnable = true
		prometheus.SetTransferState(1)
		//var current mysql.Position
		for {
			needFlush = false
			needSavePos = false
			select {
			case v := <-Msg.queue:
				prometheus.DecTransferMsg()
				switch v := v.(type) {
				// 这里 RDS高可用集群如果发生主备切换会造成binlog pos 位点漂移，程序保存的时候保存gtid set
				case model.PosMsg:
					//now := time.Now()
					//if v.Force || now.Sub(lastSavedTime) > 3*time.Second {
					//	lastSavedTime = now
					//	needFlush = true
					//	needSavePos = true
					//	current = mysql.Position{
					//		Name: v.Name,
					//		Pos:  v.Pos,
					//	}
					//}
					Postion = v
				case model.GtidMsg:
					now := time.Now()
					if v.Force || now.Sub(lastSavedTime) > 3*time.Second {
						//保存gtid位点
						lastSavedTime = now
						//s := strings.Split(v.GtidSet,":")
						//GtidSet = s[0] +":1-" + s[1]
						GTID = v.GtidSet
						needFlush = true
						needSavePos = true

					}
				case model.DdlMsg:
					// 表结构变更通知

					body, err := json.Marshal(v)
					if err != nil {
						golog.Warnf("json 格式化失败 %v", v)
						continue
					}
					err = t.endpoint.Produce(body)
					if err != nil {
						golog.Error("endpoint 写入binlog 失败 %s", err.Error())
						prometheus.SetEndpointState(0)
						return

					}
					message.SendAllChannel("发生数据库DDL事件", fmt.Sprintf("db: %s\nSQL:%s", v.Schema, v.Query))
					//message.SendAllChannel("发生表结构变更", v.Schema)
					prometheus.UpdateActionNum("ddl", string(v.Schema))

				case []*model.RowMsg:
					rowMsgs = append(rowMsgs, v...)
					needFlush = int64(len(rowMsgs)) >= config.GlobalConfig.BulkSize
				}
			case <-t.ticker.C:
				needFlush = true
			case <-Msg.stop:
				return
			}

			if needFlush && len(rowMsgs) > 0 {

				// 写消息
				for _, row := range rowMsgs {
					meta, ok := model.TableInfo[row.Key]
					if !ok {
						golog.Warnf("未发现表结构信息 %s ", row.Key)
						continue
					}
					if meta.TableColumnSize != len(row.Row) {
						golog.Warnf("%s.%s schema mismatching", row.Schema, row.Table)
						continue
					}
					kvm := rowMap(row, meta.TableInfo)
					resp := model.MQResponse{
						Action:    row.Action,
						Timestamp: row.Timestamp,
						Schema:    row.Schema,
						Table:     row.Table,
						Values:    kvm,
					}
					//resp := new(model.MQResponse)
					//resp.Action = row.Action
					//resp.Timestamp = row.Timestamp
					//resp.Table = row.Table
					//resp.Schema = row.Schema
					//resp.Values = kvm

					if canal.UpdateAction == row.Action {
						resp.OldValues = oldRowMap(row, meta.TableInfo)
					}
					//golog.Infof("resp.Values: %s", resp.Values)
					//time.Sleep(5 * time.Second)
					//golog.Infof("resp: %#v", resp)

					body, err := json.Marshal(resp)
					//golog.Infof("%s",resp.Values)
					if err != nil {
						golog.Errorf("binlog json 异常： %v", resp)
						continue
					}
					// 对具体操作放到具体的endpoint里去 ,要用阻塞方法

					err = t.endpoint.Produce(body)
					if err != nil {
						golog.Error("endpoint 写入binlog 失败 %s", err.Error())
						prometheus.SetEndpointState(0)
						return
					}

					prometheus.UpdateActionNum(row.Action, row.Key)
					rowMsgs = rowMsgs[0:0]

				}
			}

			if needSavePos {

				// 保存gtid的代码
				golog.Info("save GTIDSet: ", GTID)
				if err := config.SaveGtidSet(GTID); err != nil {
					message.SendAllChannel("【告警】保存GTID错误", "gocanal实例: "+config.GlobalConfig.Addr+"\n"+err.Error())
					golog.Errorf("gtidset 保存错误: %v", err.Error())
				}

			}

		}
	}()
	return "已经触发 transfer 启动"
}

var statusMap = map[bool]string{
	true:  "ok",
	false: "failed",
}

func (t *transfer) HealthCheck() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				//golog.Info("健康检查开始")
				//检查 endpoint 是否健康
				if t.endpointEnable == false {
					if config.GlobalConfig.IsRabbitmq() {
						t.endpoint.Connect()
					}
					golog.Info("endpoint 已经停止，尝试重新启动")

					// 这地方可能有问题 ，后续验证
					t.Start()
				} else {
					err := t.endpoint.Ping()
					if err != nil {
						golog.Error("endpoint not available,see the log file for details")
						golog.Error(err.Error())

						// endpoint 连接失败
						t.endpointEnable = false
						prometheus.SetEndpointState(0)

					} else if t.endpointEnable == false { // 如果是在endpoint 挂掉以后，endpoint 可以连接
						t.endpointEnable = true
						prometheus.SetEndpointState(1)
					}
				}
				// 检查canal运行
				if CanalInstance.canalEnable == false {
					golog.Warn("canal 已经停止")
					prometheus.SetCanalState(0)
					// 考虑尝试启动启动 canal

				}

				// 通知
				if CanalInstance.canalEnable == false || t.endpointEnable == false {

					message.SendAllChannel("go-canal 异常", "主机 : "+t.hostname+"  \nIP : "+t.ipAddress+
						"  \n数据库 : "+config.GlobalConfig.Addr+"   \ngo-canal  状态 : "+statusMap[CanalInstance.canalEnable]+
						".   \nendpoint : "+t.endpoint.String()+"  状态 : "+statusMap[t.endpointEnable])
				}
				//golog.Info("健康检查结束")

			case <-t.loopStopSignal:
				return
			}
		}
	}()
}

func (t *transfer) Stop() {
	Msg.stop <- struct{}{}
	t.transferEnable = false
	prometheus.SetTransferState(0)

	t.ticker.Stop()
	t.endpoint.Close()
	t.endpointEnable = false
	prometheus.SetEndpointState(0)

	golog.Println("transfer stop")
}
