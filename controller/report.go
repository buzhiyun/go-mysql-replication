package controller

import (
	"github.com/buzhiyun/go-mysql-replication/prometheus"
	"github.com/buzhiyun/go-mysql-replication/utils"
	"github.com/kataras/iris/v12"
)

func GetTableReport(ctx iris.Context) {

	ctx.JSON(utils.ApiJson{
		Status: 0,
		Msg:    "",
		Data:   prometheus.GetTableReport(),
	})
}

func GetSchemaReport(ctx iris.Context) {
	ctx.JSON(utils.ApiJson{
		Status: 0,
		Msg:    "",
		Data:   prometheus.GetSchmaDdlReport(),
	})
}
