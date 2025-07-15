package email

import (
	"errors"
	"fmt"
	"github/JustGopher/Gotaxy/pkg/utils"

	"gopkg.in/gomail.v2"
)

func SendEmail(to string, subject string, body string) error {
	from := "1927836028@qq.com"
	host := "smtp.qq.com"
	port := 465
	username := "1927836028@qq.com"
	password := "rznfvafkkqqzjddj"

	// 邮箱验证
	matchEmail := utils.IsValidateEmail(to)
	if !matchEmail {
		err := errors.New("邮箱格式错误")
		return err
	}
	m := gomail.NewMessage()
	// 设置邮件消息的头部字段
	m.SetHeader("From", from)       // 发送人
	m.SetHeader("To", to)           // 接收人
	m.SetHeader("Subject", subject) // 主题
	m.SetBody("text/html", body)    // 正文内容
	// 创建一个新的邮件拨号器对象，用于通过指定的 SMTP 服务器发送邮件
	d := gomail.NewDialer(host, port, username, password)
	// 通过拨号器对象发送指定的邮件消息
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}
	return nil
}
