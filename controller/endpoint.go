package controller

import (
	"github.com/buzhiyun/go-mysql-replication/service"
	"github.com/buzhiyun/go-mysql-replication/utils"
	"github.com/kataras/iris/v12"
)

func GetEndpointType(ctx iris.Context) {
	ctx.JSON(utils.ApiJson{
		Status: 0,
		Msg:    "",
		Data:   service.TransferInstance.GetEndpointInfo(),
	})
}

func GetEndpointState(ctx iris.Context) {
	ctx.JSON(utils.ApiJson{
		Status: 0,
		Msg:    "",
		Data:   service.TransferInstance.GetEndpointState(),
	})
}
