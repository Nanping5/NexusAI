package myemail

import (
	myredis "NexusAi/common/redis"
	"NexusAi/config"
	mylogger "NexusAi/pkg/logger"
	"NexusAi/pkg/utils"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strconv"

	"github.com/jordan-wright/email"
	"go.uber.org/zap"
)

func SendCaptcha(emailAddr string) (msg string, err error) {
	// 检查发送频率限制
	canSend, remainingSeconds, err := myredis.CheckCaptchaRateLimit(emailAddr)
	if err != nil {
		mylogger.Logger.Error("检查频率限制失败", zap.String("email", emailAddr), zap.Error(err))
		return "系统繁忙，请稍后再试", err
	}
	if !canSend {
		return fmt.Sprintf("发送太频繁，请 %d 秒后再试", remainingSeconds), nil
	}

	smtpServer := config.GetConfig().Smtp.SmtpServer
	senderEmail := config.GetConfig().Smtp.EmailAddr
	smtpKey := config.GetConfig().Smtp.SmtpKey

	captcha := strconv.Itoa(utils.GenerateRandomCode(6))

	// 发送邮件
	em := email.NewEmail()
	em.From = fmt.Sprintf("NexusAi <%s>", senderEmail)
	em.To = []string{emailAddr}
	em.Subject = "NexusAi 验证码"
	em.Text = []byte(fmt.Sprintf("您的验证码是: %s, 有效期为3分钟", captcha))

	//这里需要根据你的SMTP服务器配置调整端口
	err = em.SendWithTLS(
		smtpServer+":465",
		smtp.PlainAuth("", senderEmail, smtpKey, smtpServer),
		&tls.Config{ServerName: smtpServer},
	)
	if err != nil {
		mylogger.Logger.Error("发送邮件失败", zap.String("email", emailAddr), zap.Error(err))
		return "发送邮件失败，请稍后再试", err
	}

	// 邮件发送成功后，再设置验证码到 Redis
	if err := myredis.SetCaptchaForEmail(emailAddr, captcha); err != nil {
		mylogger.Logger.Error("设置验证码失败", zap.String("email", emailAddr), zap.Error(err))
		return "设置验证码失败，请稍后再试", err
	}

	// 设置发送频率限制
	if err := myredis.SetCaptchaRateLimit(emailAddr, 60); err != nil {
		mylogger.Logger.Warn("设置频率限制失败", zap.String("email", emailAddr), zap.Error(err))
	}

	msg = "验证码已发送，请注意查收"
	mylogger.Logger.Info(msg, zap.String("email", emailAddr))
	return msg, nil
}
