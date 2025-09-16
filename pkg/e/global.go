package e

import "warnnotice/util"

var (
	EmailConfig        *util.EmailConfig
	MonitorConfig      *util.MonitorConfig
	Monitor            *util.SystemMonitor
	ScriptConfig       *util.ScriptConfig
	ScriptReturnConfig map[int]string // 脚本返回值配置
	SystemName         string
	MonitorStopChan    chan bool
	ScriptStopChan     chan bool
)
