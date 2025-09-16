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

// 脚本配置处理函数
func SetScriptConfig(c *gin.Context) {
	var config util.ScriptConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.INVALID_PARAMS,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	// 保存到数据库
	err := database.SaveScriptConfig(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "保存脚本配置失败: " + err.Error(),
		})
		return
	}

	// 更新全局变量
	e.ScriptConfig = &config

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "脚本配置保存成功",
	})
}
func GetScriptConfig(c *gin.Context) {
	config, err := database.GetScriptConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取脚本配置失败: " + err.Error(),
		})
		return
	}
	scheduler.RestartScriptScheduler()
	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "获取脚本配置成功",
		"data": config,
	})
}
func TestScript(c *gin.Context) {
	if e.ScriptConfig == nil || e.ScriptConfig.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.ERROR,
			"msg":  "请先配置脚本路径",
		})
		return
	}

	result, output, err := util.ExecuteScript(*e.ScriptConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "脚本执行失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "脚本执行成功,结果：" + output,
		"data": gin.H{
			"result": result,
		},
	})
}

func GetScriptHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")

	// 将字符串转换为整数
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10 // 如果转换失败，使用默认值10
	}

	histories, err := database.GetScriptHistory(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取脚本执行历史失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "获取脚本执行历史成功",
		"data": histories,
	})
}

// 脚本返回值配置处理函数
func SetScriptReturnConfig(c *gin.Context) {
	type ScriptReturnConfigRequest struct {
		ReturnValue int    `json:"return_value"`
		AlertText   string `json:"alert_text"`
	}

	var req ScriptReturnConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.INVALID_PARAMS,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	// 保存到数据库
	err := database.SaveScriptReturnConfig(req.ReturnValue, req.AlertText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "保存脚本返回值配置失败: " + err.Error(),
		})
		return
	}

	// 更新全局变量
	if e.ScriptReturnConfig == nil {
		e.ScriptReturnConfig = make(map[int]string)
	}
	e.ScriptReturnConfig[req.ReturnValue] = req.AlertText

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "脚本返回值配置保存成功",
	})
}

func GetScriptReturnConfigs(c *gin.Context) {
	configs, err := database.GetAllScriptReturnConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取脚本返回值配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "获取脚本返回值配置成功",
		"data": configs,
	})
}
