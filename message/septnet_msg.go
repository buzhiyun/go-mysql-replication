package message

import (
	"errors"
	"github.com/buzhiyun/go-mysql-replication/config"
	"github.com/buzhiyun/go-mysql-replication/utils"
	"github.com/kataras/golog"
	"strings"
	"time"
)

type septnetMsgChannel struct {
	api      string
	toUsers  string
	sendTime time.Time
}

type septnetMsg struct {
	Content string `json:"content"`
	Touser  string `json:"touser"`
}

func (s *septnetMsgChannel) Send(title string, content string) (err error) {
	// 通知静默期 5分钟
	if time.Now().Sub(s.sendTime) < time.Minute*5 {
		golog.Info("处在通知静默期，暂不发送")
		return
	}
	body := strings.ReplaceAll(content, "\n", "\n> ")
	_, err = utils.HttpPostJson(s.api+"/api/wechatwork/msg/markdown", septnetMsg{
		Content: "### " + title + "\n> " + body,
		Touser:  s.toUsers,
	})
	if err != nil {
		golog.Errorf("发送企业微信消息错误，%s", err.Error())
	}
	s.sendTime = time.Now()
	return
}

func (s *septnetMsgChannel) InitChannel() (err error) {
	if len(config.GlobalConfig.Notice.SeptnetMsg.Api) == 0 || len(config.GlobalConfig.Notice.SeptnetMsg.ToUsers) == 0 {
		golog.Warnf("七天微信 初始化参数不足  %v ", config.GlobalConfig.Notice.SeptnetMsg.Api)
		return errors.New("不能初始化 septnetMsg 通道")
	}
	s.toUsers = config.GlobalConfig.Notice.SeptnetMsg.ToUsers
	s.api = config.GlobalConfig.Notice.SeptnetMsg.Api
	return
}

func (s *septnetMsgChannel) String() string {
	return "七天网络通知"
}
