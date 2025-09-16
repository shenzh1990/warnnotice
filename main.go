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
	"warnnotice/scheduler"
)

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	initDb()
	initConfig()
	scheduler.InitMonitor()
	scheduler.InitScriptScheduler()
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
		log.Printf("Server is running on http://localhost:%d", settings.HTTPPort)
		if err := s.ListenAndServe(); err != nil {
			log.Printf("Listen: %v\n", err)
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
