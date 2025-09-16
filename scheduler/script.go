package scheduler

import (
	"fmt"
	"github.com/huobirdcenter/huobi_golang/logging/applogger"
	"time"
	"warnnotice/database"
	"warnnotice/pkg/e"
	"warnnotice/util"
)

// 初始化脚本定时执行任务
func InitScriptScheduler() {
	// 初始化停止通道
	e.ScriptStopChan = make(chan bool, 1)
	go func() {
		var ticker *time.Ticker
		var interval int

		// 如果有脚本配置且设置了执行间隔
		if e.ScriptConfig != nil && e.ScriptConfig.Interval > 0 {
			interval = e.ScriptConfig.Interval
			ticker = time.NewTicker(time.Duration(interval) * time.Minute)
			defer ticker.Stop()
		}

		for {
			// 检查配置是否发生变化
			if e.ScriptConfig != nil && e.ScriptConfig.Interval > 0 {
				if ticker == nil || interval != e.ScriptConfig.Interval {
					// 配置发生变化，重新创建 ticker
					if ticker != nil {
						ticker.Stop()
					}
					interval = e.ScriptConfig.Interval
					ticker = time.NewTicker(time.Duration(interval) * time.Minute)
				}
			} else {
				// 没有配置或间隔为0，停止 ticker
				if ticker != nil {
					ticker.Stop()
					ticker = nil
				}
			}

			if ticker != nil {
				select {
				case <-ticker.C:
					// 执行脚本
					result, output, err := util.ExecuteScript(*e.ScriptConfig)

					// 保存执行历史
					err = database.SaveScriptHistory(result, output)
					if err != nil {
						applogger.Error("保存脚本执行历史失败: %v", err)
					}

					// 根据返回值发送不同的告警邮件
					// 返回值为0时正常，不发送邮件
					if result != 0 {
						// 查找是否有对应的告警文本配置
						if alertText, exists := e.ScriptReturnConfig[result]; exists && alertText != "" {
							// 发送对应返回值的告警邮件
							if e.EmailConfig != nil && e.EmailConfig.SMTPHost != "" {
								subject := fmt.Sprintf("[%s] 脚本执行告警", e.SystemName)
								err = util.SendEmail(*e.EmailConfig, subject, alertText)
								// 保存发送记录
								sendStatus := true
								errorMessage := ""
								if err != nil {
									sendStatus = false
									errorMessage = err.Error()
								}
								// 保存到数据库
								saveErr := database.SaveAlertHistory(e.EmailConfig.To, subject, alertText, sendStatus, errorMessage)
								if saveErr != nil {
									applogger.Error("保存脚本告警发送记录失败: %v", saveErr)
								}
								if err != nil {
									applogger.Error("发送脚本告警邮件失败: %v", err)
								}
							}
						}
					}

				case <-e.ScriptStopChan:
					// 收到停止信号，退出循环
					return
				}
			} else {
				// 没有配置时等待一段时间再检查
				select {
				case <-time.After(1 * time.Minute):
				case <-e.ScriptStopChan:
					// 收到停止信号，退出循环
					return
				}
			}
		}
	}()
}

// 重新启动脚本调度器
func RestartScriptScheduler() {
	// 发送停止信号
	if e.ScriptStopChan != nil {
		select {
		case e.ScriptStopChan <- true:
		default:
		}
	}

	// 重新初始化脚本调度器
	InitScriptScheduler()
}
