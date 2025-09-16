package scheduler

import (
	"fmt"
	"github.com/huobirdcenter/huobi_golang/logging/applogger"
	"time"
	"warnnotice/database"
	"warnnotice/pkg/e"
	"warnnotice/util"
)

// 初始化监控器
func InitMonitor() {
	// 如果没有监控配置，使用默认配置
	config := util.MonitorConfig{
		Interval:      5,
		AvgCount:      3,
		CPUThreshold:  80.0,
		MemThreshold:  80.0,
		DiskThreshold: 85.0,
	}

	if e.MonitorConfig != nil {
		config = *e.MonitorConfig
	}

	e.Monitor = util.NewSystemMonitor(config, func(alertMsg string) error {
		// 发送告警邮件
		if e.EmailConfig != nil && e.EmailConfig.SMTPHost != "" {
			subject := fmt.Sprintf("[%s] 系统监控告警", e.SystemName)
			err := util.SendEmail(*e.EmailConfig, subject, alertMsg)
			// 保存发送记录
			sendStatus := true
			errorMessage := ""
			if err != nil {
				sendStatus = false
				errorMessage = err.Error()
			}
			// 保存到数据库
			saveErr := database.SaveAlertHistory(e.EmailConfig.To, subject, alertMsg, sendStatus, errorMessage)
			if saveErr != nil {
				applogger.Error("保存告警发送记录失败: %v", saveErr)
			}
			return err
		}
		return nil
	})
	// 初始化停止通道
	e.MonitorStopChan = make(chan bool, 1)
	// 启动定时监控任务
	go func() {
		ticker := time.NewTicker(time.Duration(config.Interval) * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 获取系统状态
				status, err := util.GetSystemStatus()
				if err != nil {
					applogger.Error("获取系统状态失败: %v", err)
					continue
				}

				// 保存系统状态到数据库
				err = database.SaveSystemStatus(*status)
				if err != nil {
					applogger.Error("保存系统状态失败: %v", err)
				}

				// 添加到历史记录
				e.Monitor.AddStatus(*status)

				// 检查阈值
				err = e.Monitor.CheckThreshold()
				if err != nil {
					applogger.Error("检查系统阈值失败: %v", err)
				}

			case <-e.MonitorStopChan:
				// 收到停止信号，退出循环
				return
			}
		}
	}()
}

// 重新启动监控任务
func RestartMonitor() {
	// 发送停止信号
	if e.MonitorStopChan != nil {
		select {
		case e.MonitorStopChan <- true:
		default:
		}
	}

	// 重新初始化监控器
	InitMonitor()
}
