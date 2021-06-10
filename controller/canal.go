package controller

import (
	"github.com/buzhiyun/go-mysql-replication/config"
	"github.com/buzhiyun/go-mysql-replication/service"
	"github.com/buzhiyun/go-mysql-replication/utils"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
	"io/ioutil"
	"strings"
)

func GetCanalState(ctx iris.Context) {
	ctx.JSON(utils.ApiJson{
		Status: 0,
		Msg:    "",
		Data:   service.CanalInstance.GetCanelState(),
	})
}

type canalInfo struct {
	BinlogFile string `json:"binlog_file"`
	BinlogPos  uint32 `json:"binlog_pos"`
	GTIDSet    string `json:"gtid_set"`
}

func GetCanalInfo(ctx iris.Context) {

	ctx.JSON(utils.ApiJson{
		Status: 0,
		Msg:    "",
		Data: canalInfo{
			BinlogFile: service.Postion.Name,
			BinlogPos:  service.Postion.Pos,
			GTIDSet:    service.GTID,
		},
	})

}

func StopCanal(ctx iris.Context) {
	service.CanalInstance.Close()
	ctx.JSON(utils.ApiJson{
		Status: 0,
		Msg:    "stop success",
		Data:   "ok",
	})
}

func StartCanal(ctx iris.Context) {
	var gtid struct {
		Gtid string `json:"gtid"`
	}
	err := ctx.ReadJSON(&gtid)
	if err != nil {
		ctx.JSON(utils.ApiJson{Status: 500, Msg: "GTID 读取异常", Data: "failed"})
		return
	}
	gtidset := strings.TrimSpace(gtid.Gtid)
	if len(gtidset) > 0 {
		err = ioutil.WriteFile(config.GlobalConfig.FromGtidFile, []byte(gtidset), 0777)
		if err != nil {
			golog.Error("from_gtid_file 配置的文件权限异常： ", err.Error())
		}
	}
	msg := service.CanalInstance.StartManual()
	ctx.JSON(utils.ApiJson{
		Status: 0,
		Msg:    msg,
		Data:   "ok",
	})
}
