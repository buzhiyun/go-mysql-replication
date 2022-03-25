package message

import (
	"errors"
	"github.com/buzhiyun/go-mysql-replication/config"
	"github.com/buzhiyun/go-mysql-replication/utils"
	"github.com/kataras/golog"
	"strings"
	"time"
)

type wechatWorkMsg struct {
	Msgtype  string `json:"msgtype"` // 固定用markdown 类型的吧，我也不想写了
	Markdown struct {
		Content string `json:"content"`
	} `json:"markdown"`
}

type wechatWorkChannel struct {
	webhookUrl   string
	LastSendTime time.Time
}

func (w *wechatWorkChannel) String() string {
	return "企业微信机器人"
}

func (w *wechatWorkChannel) InitChannel() (err error) {
	if len(config.GlobalConfig.Notice.WechatWorkRobot.WebhookUrl) == 0 {
		golog.Warnf("企业微信机器人 初始化参数不足  %v ", config.GlobalConfig.Notice.WechatWorkRobot.WebhookUrl)
		return errors.New("不能初始化 企业微信机器人 通道")
	}
	w.webhookUrl = config.GlobalConfig.Notice.WechatWorkRobot.WebhookUrl
	w.LastSendTime = time.Now().Add(-1 * time.Hour)

	return
}

func (w *wechatWorkChannel) Send(title string, content string) (err error) {

	if time.Now().Sub(w.LastSendTime) < 1800*time.Second {
		golog.Info("企业微信机器人 通道静默中，跳过发送")
		return
	}

	var contentLines []string

	for _, str := range strings.Split(content, "\n") {
		str = strings.ReplaceAll(str, "failed", "<font color=\"warning\">**failed**</font>")
		line := strings.SplitN(str, ":", 2)
		if len(line) == 2 {
			contentLines = append(contentLines, "> "+line[0]+":<font color=\"comment\">"+line[1]+"</font>")
		} else {
			contentLines = append(contentLines, "> "+str)
		}
	}
	content = "### " + title + "\n" + strings.Join(contentLines, "\n")
	wechatMag := wechatWorkMsg{
		Msgtype: "markdown",
		Markdown: struct {
			Content string `json:"content"`
		}{
			content,
		},
	}

	res, err := utils.HttpPostJson(w.webhookUrl, wechatMag)
	if err == nil {
		golog.Infof("企业微信机器人 通知成功 , 接口返回 %s", res)
		w.LastSendTime = time.Now()
	}
	return
}
