package main

import (
	"context"
	"fmt"
	"github.com/huobirdcenter/huobi_golang/logging/applogger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"warnnotice/database"
	"warnnotice/pkg/e"
	"warnnotice/pkg/settings"
	"warnnotice/router"
	"warnnotice/util"
)

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	initDb()
	initConfig()
	initMonitor()
	initScriptScheduler()
	webServer()
}

func initDb() {
	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatal("初始化数据库失败:", err)
	}
}

func webServer() {
	r := router.InitRouter()
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", settings.HTTPPort),
		Handler:        r,
		ReadTimeout:    settings.ReadTimeout,
		WriteTimeout:   settings.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Println("Listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	log.Println("Server exiting")
}

// 初始化配置
func initConfig() {
	// 加载系统名称
	systemName, err := database.GetSystemName()
	if err != nil {
		applogger.Error("加载系统名称失败: %v", err)
		e.SystemName = "告警通知系统"
	} else {
		e.SystemName = systemName
	}

	// 加载邮件配置
	emailCfg, err := database.GetEmailConfig()
	if err != nil {
		applogger.Error("加载邮件配置失败: %v", err)
	} else if emailCfg != nil {
		e.EmailConfig = emailCfg
	}

	// 加载脚本配置
	scriptCfg, err := database.GetScriptConfig()
	if err != nil {
		applogger.Error("加载脚本配置失败: %v", err)
	} else if scriptCfg != nil {
		e.ScriptConfig = scriptCfg
	}

	// 加载脚本返回值配置
	scriptReturnCfg, err := database.GetAllScriptReturnConfigs()
	if err != nil {
		applogger.Error("加载脚本返回值配置失败: %v", err)
	} else {
		e.ScriptReturnConfig = scriptReturnCfg
	}

	// 加载监控配置
	monitorCfg, err := database.GetMonitorConfig()
	if err != nil {
		applogger.Error("加载监控配置失败: %v", err)
	} else if monitorCfg != nil {
		e.MonitorConfig = monitorCfg
	}
}

// 初始化监控器
func initMonitor() {
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
			return util.SendEmail(*e.EmailConfig, subject, alertMsg)
		}
		return nil
	})

	// 启动定时监控任务
	go func() {
		for {
			interval := config.Interval
			if interval <= 0 {
				interval = 5 // 默认5分钟
			}

			// 等待指定的时间间隔
			time.Sleep(time.Duration(interval) * time.Minute)

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
		}
	}()
}

// 初始化脚本定时执行任务
func initScriptScheduler() {
	go func() {
		for {
			// 检查是否有脚本配置且设置了执行间隔
			if e.ScriptConfig != nil && e.ScriptConfig.Interval > 0 {
				// 等待指定的时间间隔
				time.Sleep(time.Duration(e.ScriptConfig.Interval) * time.Minute)

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
							if err != nil {
								applogger.Error("发送脚本告警邮件失败: %v", err)
							}
						}
					}
				}
			} else {
				// 如果没有配置脚本或未设置间隔，等待1分钟再检查
				time.Sleep(1 * time.Minute)
			}
		}
	}()
}
