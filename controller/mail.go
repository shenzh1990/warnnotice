package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"warnnotice/database"
	"warnnotice/pkg/e"
	"warnnotice/util"
)

// 邮件配置处理函数
func SetEmailConfig(c *gin.Context) {
	var config util.EmailConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.INVALID_PARAMS,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	// 保存到数据库
	err := database.SaveEmailConfig(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "保存邮件配置失败: " + err.Error(),
		})
		return
	}

	// 更新全局变量
	e.EmailConfig = &config

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "邮件配置保存成功",
	})
}
func GetEmailConfig(c *gin.Context) {
	config, err := database.GetEmailConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取邮件配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "获取邮件配置成功",
		"data": config,
	})
}
func TestEmail(c *gin.Context) {
	if e.EmailConfig == nil || e.EmailConfig.SMTPHost == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.ERROR,
			"msg":  "请先配置邮件参数",
		})
		return
	}

	err := util.SendTestEmail(*e.EmailConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "邮件发送失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "测试邮件发送成功",
	})
}
