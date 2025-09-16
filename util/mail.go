package util

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"
)

// EmailConfig 邮件配置结构
type EmailConfig struct {
	SMTPHost string `json:"smtp_host"`
	SMTPPort int    `json:"smtp_port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	To       string `json:"to"`
}

// SendTestEmail 发送测试邮件
func SendTestEmail(config EmailConfig) error {
	// 构建邮件内容
	subject := "测试邮件"
	body := "这是一封测试邮件，用于验证SMTP配置是否正确。"

	return SendEmail(config, subject, body)
}

// SendEmail 发送邮件
func SendEmail(config EmailConfig, subject, body string) error {
	// 构造邮件内容
	message := fmt.Sprintf(
		"To: %s\r\n"+
			"From: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n"+
			"\r\n%s",
		config.To,
		config.From,
		subject,
		body,
	)

	// 连接SMTP服务器
	host := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)
	auth := smtp.PlainAuth("", config.Username, config.Password, config.SMTPHost)

	// TLS配置
	tlsConfig := &tls.Config{
		ServerName: config.SMTPHost,
	}

	conn, err := net.DialTimeout("tcp", host, 10*time.Second)
	if err != nil {
		return fmt.Errorf("连接SMTP服务器失败: %v", err)
	}
	defer conn.Close()

	// 建立TLS连接
	tlsConn := tls.Client(conn, tlsConfig)
	err = tlsConn.Handshake()
	if err != nil {
		return fmt.Errorf("TLS握手失败: %v", err)
	}
	defer tlsConn.Close()

	// 创建SMTP客户端
	client, err := smtp.NewClient(tlsConn, config.SMTPHost)
	if err != nil {
		return fmt.Errorf("创建SMTP客户端失败: %v", err)
	}
	defer client.Quit()

	// 启用Auth
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP认证失败: %v", err)
	}

	// 设置发件人
	if err = client.Mail(config.From); err != nil {
		return fmt.Errorf("设置发件人失败: %v", err)
	}

	// 设置收件人
	if err = client.Rcpt(config.To); err != nil {
		return fmt.Errorf("设置收件人失败: %v", err)
	}

	// 发送邮件内容
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("准备发送邮件内容失败: %v", err)
	}

	_, err = writer.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("发送邮件内容失败: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("关闭邮件内容写入器失败: %v", err)
	}

	return nil
}
