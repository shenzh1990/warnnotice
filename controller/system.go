package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"warnnotice/database"
	"warnnotice/pkg/e"
)

// 系统名称处理函数
func SetSystemName(c *gin.Context) {
	type SystemNameRequest struct {
		Name string `json:"name"`
	}

	var req SystemNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "系统名称不能为空",
		})
		return
	}

	// 保存到数据库
	err := database.SaveSystemName(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "保存系统名称失败: " + err.Error(),
		})
		return
	}

	// 更新全局变量
	e.SystemName = req.Name

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "系统名称保存成功",
	})
}

func GetSystemName(c *gin.Context) {
	name, err := database.GetSystemName()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "获取系统名称失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取系统名称成功",
		"data": gin.H{
			"name": name,
		},
	})
}
