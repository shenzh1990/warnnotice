package router

import (
	"github.com/gin-gonic/gin"
	"github.com/huobirdcenter/huobi_golang/logging/applogger"
	"net/http"
	"time"
	"warnnotice/controller"
	"warnnotice/middleware/cors"
	"warnnotice/pkg/settings"
)

func InitRouter() *gin.Engine {
	r := gin.New()

	r.Use(Logger())
	r.Use(gin.Recovery())
	r.Use(cors.Cors())
	gin.SetMode(settings.RunMode)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	//r.LoadHTMLFiles("view/index.html")
	r.Static("/static", "./view/static")
	r.StaticFile("/index", "view/index.html")

	// 添加API路由组
	api := r.Group("/api/v1")
	{
		// 系统配置相关路由
		api.POST("/system/name", controller.SetSystemName)
		api.GET("/system/name", controller.GetSystemName)
		// 邮件配置相关路由
		api.POST("/email/config", controller.SetEmailConfig)
		api.POST("/email/test", controller.TestEmail)
		api.GET("/email/config", controller.GetEmailConfig)

		// 脚本配置相关路由
		api.POST("/script/config", controller.SetScriptConfig)
		api.POST("/script/test", controller.TestScript)
		api.GET("/script/config", controller.GetScriptConfig)
		api.GET("/script/history", controller.GetScriptHistory)

		// 脚本返回值配置相关字典给
		api.POST("/script/return-config", controller.SetScriptReturnConfig)
		api.GET("/script/return-configs", controller.GetScriptReturnConfigs)
		// 系统监控相关路由
		api.POST("/monitor/config", controller.SetMonitorConfig)
		api.GET("/monitor/config", controller.GetMonitorConfig)
		api.GET("/monitor/status", controller.GetSystemStatus)
		api.GET("/monitor/status/history", controller.GetSystemStatusHistory)
	}

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "index")
	})
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"status": 404,
			"error":  "404, page not exists!",
		})
	})
	return r
}
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方式
		reqMethod := c.Request.Method

		// 请求路由
		reqUri := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 请求IP
		clientIP := c.ClientIP()

		//日志格式
		applogger.Info("| %3d | %13v | %15s | %s | %s |",
			statusCode,
			latencyTime,
			clientIP,
			reqMethod,
			reqUri,
		)
	}
}
