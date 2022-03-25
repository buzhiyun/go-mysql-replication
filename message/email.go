package message

import (
	"crypto/tls"
	"errors"
	"github.com/buzhiyun/go-mysql-replication/config"
	"github.com/go-gomail/gomail"
	"github.com/kataras/golog"
	"strings"
	"time"
)

type emailChannel struct {
	ToUser       string // 收件人
	CcUser       string // 抄送
	ServerConfig mailConf
	LastSendTime time.Time
}

type mailConf struct {
	Host     string
	port     int
	Username string
	Passwd   string
}

func (e *emailChannel) String() string {
	return "email"
}

func (e *emailChannel) InitChannel() (err error) {

	if len(config.GlobalConfig.Notice.Email.Host) == 0 || len(config.GlobalConfig.Notice.Email.ToUser) == 0 {
		golog.Warnf("email 初始化参数不足 : %v ", config.GlobalConfig.Notice.Email)
		return errors.New("不能初始化 email 通道")
	}
	e.ServerConfig = mailConf{
		Host:     config.GlobalConfig.Notice.Email.Host,
		port:     config.GlobalConfig.Notice.Email.Port,
		Username: config.GlobalConfig.Notice.Email.LoginUser,
		Passwd:   config.GlobalConfig.Notice.Email.LoginPasswd,
	}
	e.ToUser = strings.Join(config.GlobalConfig.Notice.Email.ToUser, ",")
	e.LastSendTime = time.Now().Add(-1 * time.Hour)
	//e.Subject = "【告警】binlog 发送出现问题"

	return nil
}

func (e *emailChannel) Send2UserContext(toUser string, title string, context string) (err error) {
	// 通道静默 1800s
	if time.Now().Sub(e.LastSendTime) < 1800*time.Second {
		return
	}
	return e.SendMail(e.ToUser, "", title, context, "")
}

func (e *emailChannel) Send(title string, content string) (err error) {
	// 通道静默 1800s
	if time.Now().Sub(e.LastSendTime) < 1800*time.Second {
		golog.Info("email 通道静默中，跳过发送")
		return
	}
	return e.SendMail(e.ToUser, e.CcUser, title, GetHtmlstr("", content), "")
}

func FormatAddress(addressStr string, m *gomail.Message) []string {
	addresses := strings.Split(addressStr, ",")

	addr := []string{}
	//strconv.Atoi(strconv.FormatInt(config.Conf.Get("main.port").(int64), 10))
	for _, value := range addresses {
		//log.Println("value:",value)
		thisAdd := strings.Split(value, "<")
		if len(thisAdd) > 1 {
			//log.Println("addr: ", strings.TrimRight(thisAdd[1], ">"))
			addr = append(addr, m.FormatAddress(strings.TrimRight(thisAdd[1], ">"), thisAdd[0]))
		} else {
			addr = append(addr, value)
		}
	}
	return addr
}

func GetHtmlstr(toUsername string, content string) (htmlStr string) {
	content = strings.ReplaceAll(content, "\n", "<br>")
	mailstr1 := `<div><div style="font-family: &quot;Microsoft YaHei UI&quot;; line-height: 21px;"><span style="font-size: 10.5pt; line-height: 1.5; background-color: window;">`

	mailstr2 := `，您好：</span></div><div style="font-family: &quot;Microsoft YaHei UI&quot;; line-height: 21px;"><span style="font-size: 10.5pt; line-height: 1.5; background-color: window;"><br></span></div><div style="font-family: &quot;Microsoft YaHei UI&quot;; line-height: 21px;"><span style="font-family: ''; font-size: 10.5pt; line-height: 1.5; background-color: window;">&nbsp; &nbsp;&nbsp;</span><span style="font-family: ''; font-size: 10.5pt; line-height: 1.5; background-color: window;">&nbsp; &nbsp;&nbsp;</span><span style="background-color: window; font-size: 10.5pt; line-height: 1.5;">`

	mailstr3 := `</span></div><div style="font-family: &quot;Microsoft YaHei UI&quot;; line-height: 21px;"><br></div><hr color="#b5c4df" size="1" align="left" style="box-sizing: border-box; font-family: &quot;Microsoft YaHei UI&quot;; line-height: 21px; width: 210px; height: 1px;"><div style="font-family: &quot;Microsoft YaHei UI&quot;; line-height: 21px;"><div style="position: static !important; margin: 10px; font-size: 10pt;"><div class="MsoNormal" align="left" style="font-size: 10.5pt; line-height: normal; font-family: Calibri, sans-serif; text-align: justify; margin: 0cm 0cm 0.0001pt;"><span style="font-size: 10.5pt; line-height: 1.5; background-color: window;">Best regards,</span></div><p class="MsoNormal" style="margin: 0cm 0cm 0.0001pt; font-size: 10.5pt; line-height: normal; font-family: Calibri, sans-serif; text-align: justify;"><span lang="EN-US" style="color: rgb(31, 73, 125);">&nbsp;</span><span lang="EN-US"><o:p></o:p></span></p><p class="MsoNormal" style="margin: 0cm 0cm 0.0001pt; font-size: 10.5pt; line-height: normal; font-family: Calibri, sans-serif; text-align: justify;"><b><span style="font-size: 14pt; font-family: 华文新魏; color: rgb(64, 64, 64);">服务管理</span></b><span lang="EN-US"><o:p></o:p></span></p><p class="MsoNormal" style="margin: 0cm 0cm 0.0001pt; font-family: Verdana; font-size: 14px; line-height: normal; text-align: justify;"><font color="#ee9a1e" face="微软雅黑, sans-serif" size="2"><span style="line-height: 19px;">监控中心</span></font></p></div></div></div><div><sign signid="0"><div style="font-size:14px;font-family:Verdana;color:#000;">
	</div></sign></div><div>&nbsp;</div><div><includetail><!--<![endif]--></includetail></div>`

	if len(toUsername) > 0 {
		return mailstr1 + toUsername + mailstr2 + content + mailstr3
	} else {
		return mailstr1 + content + mailstr3
	}
}

func (e *emailChannel) SendMail(toUsers string, ccUsers string, subject string, htmlBody string, fileAttach string) error {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", e.ServerConfig.Username /*"发件人地址"*/, "发件人") // 发件人

	m.SetHeader("To", FormatAddress(toUsers, m)...) // 收件人

	//m.SetHeader("Cc",
	//	m.FormatAddress("xxxx@7net.cc", "收件人")) //抄送
	if len(ccUsers) > 0 {
		m.SetHeader("Cc", FormatAddress(ccUsers, m)...) //抄送
	}

	//m.SetHeader("Bcc",
	//	m.FormatAddress("xxxx@7net.cc", "收件人")) // 暗送

	m.SetHeader("Subject", subject) // 主题

	//m.SetBody("text/html",xxxxx ") // 可以放html..还有其他的
	m.SetBody("text/html", htmlBody) // 正文

	if len(fileAttach) > 0 {
		m.Attach(fileAttach) //添加附件
	}

	d := gomail.NewDialer(e.ServerConfig.Host, e.ServerConfig.port, e.ServerConfig.Username, e.ServerConfig.Passwd) // 发送邮件服务器、端口、发件人账号、发件人密码
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}                                                             // 这里忽略了tls 验证，防止在某些docker 容器里发生证书失败
	if err := d.DialAndSend(m); err != nil {
		golog.Errorf("发送失败: %v", err.Error())
		return err
	}
	e.LastSendTime = time.Now()
	golog.Info("email 发送成功")
	return nil

}

//func TestMail(t *testing.T)  {0
//	SendMail("天才周<zhouyang@7net.cc>,研发部<dev@7net.cc>", "", "测试邮件", GetHtmlstr("","就是测试邮件"), "")
//}
