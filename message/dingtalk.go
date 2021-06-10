package message

import (
	"errors"
	"github.com/buzhiyun/go-mysql-replication/config"
	"github.com/buzhiyun/go-mysql-replication/utils"
	"github.com/kataras/golog"
	"strconv"
	"strings"
	"time"
)

type dingtalkChannel struct {
	webhookUrl   string // 带有accesstoken 的webhook_url
	secret       string
	LastSendTime time.Time
}

func sign(timestamp string, secret string) (sign string) {
	string_to_sign := timestamp + "\n" + secret
	return utils.Base64UrlSafeEncode(utils.GethmacSha256(string_to_sign, secret))

}

type dingtalkMsg struct {
	Msgtype string `json:"msgtype"`
	//Text		struct{
	//	Content			string		`json:"content"`
	//}			`json:"text"`			// text 类型的消息
	Markdown struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	} `json:"markdown"` // markdown 类型的消息
	//Link		struct{
	//	Text			string		`json:"text"`
	//	Title			string		`json:"text"`
	//	PicUrl			string		`json:"picUrl"`
	//	MessageUrl		string		`json:"messageUrl"`
	//}			`json:"link"`			// linke 类型的消息
	//At			struct{
	//	AtMobiles	[]string		`json:"atMobiles"`
	//	AtUserIds	[]string		`json:"atUserIds"`
	//	IsAtAll		bool			`json:"isAtAll"`
	//}		`json:"at"`					// 是否 @xxx
} // 我就一个告警通知消息， 其他类型的消息类型以后再实现

func (d *dingtalkChannel) InitChannel() (err error) {
	if len(config.GlobalConfig.Notice.DingtalkRobot.WebhookUrl) == 0 {
		golog.Warnf("dingtalk 初始化参数不足  %v ", config.GlobalConfig.Notice.DingtalkRobot.WebhookUrl)
		return errors.New("不能初始化 email 通道")
	}
	d.secret = config.GlobalConfig.Notice.DingtalkRobot.Secret
	d.webhookUrl = config.GlobalConfig.Notice.DingtalkRobot.WebhookUrl
	d.LastSendTime = time.Now().Add(-1 * time.Hour)

	return
}

func (d *dingtalkChannel) SendContext(toUser string, title string, context string) (err error) {
	// 没实现， 暂时也不想实现
	return
}

func (d *dingtalkChannel) Send(title string, content string) (err error) {
	if time.Now().Sub(d.LastSendTime) < 1800*time.Second {
		golog.Info("企业微信机器人 通道静默中，跳过发送")
		return
	}

	var signUrl string
	if len(d.secret) > 0 {
		timestamp := time.Now().Unix() * 1000
		//golog.Debug(timestamp, sign(strconv.FormatInt(timestamp,10),d.secret))
		sign := sign(strconv.FormatInt(timestamp, 10), d.secret)

		signUrl = "&timestamp=" + strconv.FormatInt(timestamp, 10) + "&sign=" + sign
	}
	url := d.webhookUrl + signUrl

	content = strings.ReplaceAll(content, "\n", "\n\n> ")
	content = "### " + title + "\n> " + strings.ReplaceAll(content, "failed", "**failed**")

	markdownMsg := dingtalkMsg{
		Msgtype: "markdown",
		Markdown: struct {
			Title string `json:"title"`
			Text  string `json:"text"`
		}{
			title,
			content,
		},
	}

	res, err := utils.HttpPostJson(url, markdownMsg)

	if err == nil {
		golog.Infof("钉钉机器人 通知成功 , 接口返回 %s", res)
		d.LastSendTime = time.Now()
	}
	return
}

func (d *dingtalkChannel) String() string {
	return "钉钉机器人"
}
