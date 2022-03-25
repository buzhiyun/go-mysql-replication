package message

import (
	"github.com/buzhiyun/go-mysql-replication/config"
	"github.com/kataras/golog"
)

type channel interface {
	InitChannel() error
	Send(title string, content string) error // 快速从通道发送报警
	String() string
}

var Channels []channel

func InitChannels() {
	golog.Info("初始化发送通道")
	Channels = GetChannal()
	golog.Infof("channels: %v", Channels)
}

func GetChannal() (channels []channel) {
	//if len(config.GlobalConfig.Notice.Email.Host) > 0 && len(config.GlobalConfig.Notice.Email.ToUser) > 0 &&
	//	len(config.GlobalConfig.Notice.Email.LoginUser) > 0 {
	//	email := new(emailChannel)
	//	if err := email.InitChannel(); err == nil {
	//		channels = append(channels, email)
	//	}
	//}
	//if len(config.GlobalConfig.Notice.DingtalkRobot.WebhookUrl) > 0 {
	//	dingtalk := new(dingtalkChannel)
	//	if err := dingtalk.InitChannel(); err == nil {
	//		channels = append(channels, dingtalk)
	//	}
	//}
	//
	//if len(config.GlobalConfig.Notice.WechatWorkRobot.WebhookUrl) > 0 {
	//	wechatWork := new(wechatWorkChannel)
	//	if err := wechatWork.InitChannel(); err == nil {
	//		channels = append(channels, wechatWork)
	//	}
	//}

	if len(config.GlobalConfig.Notice.SeptnetMsg.Api) >= 0 {
		golog.Info("初始化 七天企业微信发送通道")
		septnetMesage := new(septnetMsgChannel)
		if err := septnetMesage.InitChannel(); err == nil {
			channels = append(channels, septnetMesage)
		}
	}

	return
}

func SendAllChannel(title, context string) {
	if len(Channels) == 0 {
		golog.Warnf("无法发送告警：无可用告警通道 ！")
		return
	}
	//golog.Infof("告警通道数： %v" ,)
	for _, c := range Channels {
		go func(c channel) {
			if err := c.Send(title, context); err != nil {
				golog.Errorf("%s 通道发送失败, %v", c.String(), err)
			}
		}(c)
	}
}
