package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"warnnotice/database"
	"warnnotice/pkg/e"
	"warnnotice/scheduler"
	"warnnotice/util"
)

// 监控配置处理函数
func SetMonitorConfig(c *gin.Context) {
	var config util.MonitorConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.INVALID_PARAMS,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	// 保存到数据库
	err := database.SaveMonitorConfig(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "保存监控配置失败: " + err.Error(),
		})
		return
	}

	// 更新全局变量
	e.MonitorConfig = &config

	// 更新监控器配置
	if e.Monitor != nil {
		e.Monitor.Config = config
	}
	scheduler.RestartMonitor()

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "监控配置保存成功",
	})
}
func GetMonitorConfig(c *gin.Context) {
	config, err := database.GetMonitorConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取监控配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "获取监控配置成功",
		"data": config,
	})
}
func GetSystemStatus(c *gin.Context) {
	status, err := util.GetSystemStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取系统状态失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "获取系统状态成功",
		"data": status,
	})
}
func GetSystemStatusHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	// 将字符串转换为整数
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10 // 如果转换失败，使用默认值10
	}

	statuses, err := database.GetSystemStatusHistory(limit) // 默认获取10条记录
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取系统状态历史失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "获取系统状态历史成功",
		"data": statuses,
	})
}
