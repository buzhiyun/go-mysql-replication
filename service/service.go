package service

import (
	"github.com/buzhiyun/go-mysql-replication/message"
	"github.com/kataras/golog"
	"time"
)

var CanalInstance = &Canal{}

type msg struct {
	queue chan interface{}
	stop  chan interface{}
}

var Msg = msg{
	queue: make(chan interface{}, 4096),
	stop:  make(chan interface{}, 1),
}

var TransferInstance = transfer{
	loopStopSignal: make(chan struct{}, 1),
}

//func Initialize()  {
//
//}

func Start() (err error) {
	// 开启传输线程
	TransferInstance.Start()

	for !TransferInstance.transferEnable {
		golog.Info("等待传输线程启动")
		time.Sleep(1 * time.Second)
		if TransferInstance.transferEnable {
			break
		}
	}

	// 开启canal
	err = CanalInstance.initialize()
	CanalInstance.firstsStart = true
	CanalInstance.StartUpFromGtidSet()

	// 初始化告警通知通道
	message.InitChannels()

	TransferInstance.HealthCheck()
	return
}

func Close() {
	// 保存位点代码
	//Msg.stop <- struct{}{}
	TransferInstance.loopStopSignal <- struct{}{}

	CanalInstance.Close()
}
