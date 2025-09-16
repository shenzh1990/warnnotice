package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"warnnotice/database"
	"warnnotice/pkg/e"
)

// GetAlertHistory 获取告警消息发送历史记录
func GetAlertHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	var histories []database.AlertHistory
	if limit > 0 {
		histories, err = database.GetAlertHistory(limit)
	} else {
		histories, err = database.GetAllAlertHistory()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取告警消息发送历史失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "获取告警消息发送历史成功",
		"data": histories,
	})
}
