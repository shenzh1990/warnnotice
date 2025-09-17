package util

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// SystemStatus 系统状态结构
type SystemStatus struct {
	CPUUsage   float64            `json:"cpu_usage"`   // CPU使用率
	MemUsage   float64            `json:"mem_usage"`   // 内存使用率
	DiskUsage  float64            `json:"disk_usage"`  // 磁盘使用率(平均值)
	DiskUsages map[string]float64 `json:"disk_usages"` // 各磁盘分区使用率
	Timestamp  int64              `json:"timestamp"`   // 时间戳
}

// MonitorConfig 监控配置结构
type MonitorConfig struct {
	Interval      int     `json:"interval"`       // 监控间隔(分钟)
	AvgCount      int     `json:"avg_count"`      // 平均值计算次数
	CPUThreshold  float64 `json:"cpu_threshold"`  // CPU阈值(%)
	MemThreshold  float64 `json:"mem_threshold"`  // 内存阈值(%)
	DiskThreshold float64 `json:"disk_threshold"` // 磁盘阈值(%)
}

// SystemMonitor 系统监控器结构
type SystemMonitor struct {
	Config        MonitorConfig
	StatusHistory []SystemStatus
	AlertFunc     func(string) error // 告警函数
}

// GetSystemStatus 获取当前系统状态
func GetSystemStatus() (*SystemStatus, error) {
	status := &SystemStatus{
		Timestamp:  time.Now().Unix(),
		DiskUsages: make(map[string]float64),
	}

	// 获取CPU使用率
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("获取CPU使用率失败: %v", err)
	}
	status.CPUUsage = cpuPercent[0]

	// 获取内存使用率
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("获取内存使用率失败: %v", err)
	}
	status.MemUsage = memStat.UsedPercent

	// 获取所有磁盘分区使用率
	parts, err := disk.Partitions(true)
	if err != nil {
		// 如果无法获取分区信息，使用原来的方法
		diskStat, err := disk.Usage("/")
		if err != nil {
			// 如果根目录获取失败，尝试当前目录
			pwd, _ := os.Getwd()
			diskStat, err = disk.Usage(pwd)
			if err != nil {
				return nil, fmt.Errorf("获取磁盘使用率失败: %v", err)
			}
		}
		status.DiskUsage = diskStat.UsedPercent
		status.DiskUsages["/"] = diskStat.UsedPercent
	} else {
		var totalUsage float64
		var validPartitions int

		// 遍历所有分区
		for _, part := range parts {
			// 跳过一些虚拟文件系统
			if strings.HasPrefix(part.Fstype, "tmpfs") ||
				strings.HasPrefix(part.Fstype, "sysfs") ||
				strings.HasPrefix(part.Fstype, "proc") ||
				strings.HasPrefix(part.Fstype, "devtmpfs") ||
				strings.HasPrefix(part.Fstype, "cgroup") ||
				part.Mountpoint == "/dev" ||
				part.Mountpoint == "/sys" ||
				part.Mountpoint == "/proc" {
				continue
			}

			diskStat, err := disk.Usage(part.Mountpoint)
			if err != nil {
				// 忽略单个分区错误
				continue
			}

			status.DiskUsages[part.Mountpoint] = diskStat.UsedPercent
			totalUsage += diskStat.UsedPercent
			validPartitions++
		}

		// 计算平均磁盘使用率
		if validPartitions > 0 {
			status.DiskUsage = totalUsage / float64(validPartitions)
		} else {
			// 如果没有有效的分区，使用默认值
			status.DiskUsage = 0
		}
	}

	return status, nil
}

// NewSystemMonitor 创建系统监控器
func NewSystemMonitor(config MonitorConfig, alertFunc func(string) error) *SystemMonitor {
	return &SystemMonitor{
		Config:        config,
		StatusHistory: make([]SystemStatus, 0),
		AlertFunc:     alertFunc,
	}
}

// AddStatus 添加状态记录
func (m *SystemMonitor) AddStatus(status SystemStatus) {
	m.StatusHistory = append(m.StatusHistory, status)

	// 只保留最近的AvgCount条记录
	if len(m.StatusHistory) > m.Config.AvgCount {
		m.StatusHistory = m.StatusHistory[1:]
	}
}

// CheckThreshold 检查阈值并触发告警
func (m *SystemMonitor) CheckThreshold() error {
	// 确保有足够的历史记录
	if len(m.StatusHistory) < m.Config.AvgCount {
		return nil
	}

	// 计算平均值
	var cpuSum, memSum, diskSum float64
	for _, status := range m.StatusHistory {
		cpuSum += status.CPUUsage
		memSum += status.MemUsage
		diskSum += status.DiskUsage
	}

	avgCPU := cpuSum / float64(len(m.StatusHistory))
	avgMem := memSum / float64(len(m.StatusHistory))
	avgDisk := diskSum / float64(len(m.StatusHistory))

	// 检查是否超过阈值
	alerts := make([]string, 0)
	if avgCPU > m.Config.CPUThreshold {
		alerts = append(alerts, fmt.Sprintf("CPU使用率%.2f%%超过阈值%.2f%%", avgCPU, m.Config.CPUThreshold))
	}

	if avgMem > m.Config.MemThreshold {
		alerts = append(alerts, fmt.Sprintf("内存使用率%.2f%%超过阈值%.2f%%", avgMem, m.Config.MemThreshold))
	}

	if avgDisk > m.Config.DiskThreshold {
		// 检查具体的磁盘分区使用情况
		latestStatus := m.StatusHistory[len(m.StatusHistory)-1]
		diskAlerts := make([]string, 0)
		for mountPoint, usage := range latestStatus.DiskUsages {
			if usage > m.Config.DiskThreshold {
				diskAlerts = append(diskAlerts, fmt.Sprintf("%s分区使用率%.2f%%", mountPoint, usage))
			}
		}

		if len(diskAlerts) > 0 {
			alerts = append(alerts, fmt.Sprintf("磁盘使用率超过阈值%.2f%%: %s", m.Config.DiskThreshold, strings.Join(diskAlerts, ", ")))
		} else {
			alerts = append(alerts, fmt.Sprintf("平均磁盘使用率%.2f%%超过阈值%.2f%%", avgDisk, m.Config.DiskThreshold))
		}
	}

	// 如果有超过阈值的情况，触发告警
	if len(alerts) > 0 {
		alertMsg := "系统监控告警:\n" + fmt.Sprintf("时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		for _, alert := range alerts {
			alertMsg += alert + "\n"
		}
		return m.AlertFunc(alertMsg)
	}

	return nil
}

// GetCPUCount 获取CPU核心数
func GetCPUCount() int {
	return runtime.NumCPU()
}

// FormatPercent 格式化百分比值
func FormatPercent(value float64) float64 {
	return math.Round(value*100) / 100
}
