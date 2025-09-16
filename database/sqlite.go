package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"warnnotice/util"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// InitDB 初始化数据库
func InitDB() error {
	// 创建数据库目录
	dbDir := "./db"
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		err := os.MkdirAll(dbDir, 0755)
		if err != nil {
			return fmt.Errorf("创建数据库目录失败: %v", err)
		}
	}

	// 打开数据库，设置连接参数以减少锁定问题
	db, err := sql.Open("sqlite", filepath.Join(dbDir, "warnnotice.db")+"?_busy_timeout=10000&_journal_mode=WAL&_sync=1")
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(1) // SQLite是文件数据库，限制为1个连接可减少锁定问题
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	DB = db

	// 创建表
	err = createTables()
	if err != nil {
		return fmt.Errorf("创建表失败: %v", err)
	}

	// 初始化系统名称
	err = initSystemName()
	if err != nil {
		return fmt.Errorf("初始化系统名称失败: %v", err)
	}

	return nil
}

// createTables 创建数据表
func createTables() error {
	// 系统配置表
	systemConfigSQL := `
	CREATE TABLE IF NOT EXISTS system_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		system_name TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// 邮件配置表
	emailConfigSQL := `
	CREATE TABLE IF NOT EXISTS email_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		smtp_host TEXT NOT NULL,
		smtp_port INTEGER NOT NULL,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		from_email TEXT NOT NULL,
		to_email TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// 脚本配置表
	scriptConfigSQL := `
	CREATE TABLE IF NOT EXISTS script_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL,
		parameters TEXT,
		timeout INTEGER,
		interval INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// 脚本返回值配置表（支持无限扩展）
	scriptReturnConfigSQL := `
	CREATE TABLE IF NOT EXISTS script_return_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		return_value INTEGER NOT NULL,
		alert_text TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(return_value)
	);`

	// 监控配置表
	monitorConfigSQL := `
	CREATE TABLE IF NOT EXISTS monitor_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		interval INTEGER NOT NULL,
		avg_count INTEGER NOT NULL,
		cpu_threshold REAL NOT NULL,
		mem_threshold REAL NOT NULL,
		disk_threshold REAL NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// 系统状态历史表
	systemStatusSQL := `
	CREATE TABLE IF NOT EXISTS system_status (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cpu_usage REAL NOT NULL,
		mem_usage REAL NOT NULL,
		disk_usage REAL NOT NULL,
		timestamp INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// 脚本执行历史表
	scriptHistorySQL := `
	CREATE TABLE IF NOT EXISTS script_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		result INTEGER NOT NULL,
		output TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	tables := []string{systemConfigSQL, emailConfigSQL, scriptConfigSQL, scriptReturnConfigSQL, monitorConfigSQL, systemStatusSQL, scriptHistorySQL}

	for _, sql := range tables {
		_, err := DB.Exec(sql)
		if err != nil {
			return fmt.Errorf("执行SQL语句失败: %v", err)
		}
	}

	return nil
}

// initSystemName 初始化系统名称
func initSystemName() error {
	// 检查是否已存在系统名称
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM system_config").Scan(&count)
	if err != nil {
		return fmt.Errorf("查询系统配置失败: %v", err)
	}

	// 如果不存在，则插入默认系统名称
	if count == 0 {
		_, err = DB.Exec("INSERT INTO system_config (system_name) VALUES (?)", "告警通知系统")
		if err != nil {
			return fmt.Errorf("插入默认系统名称失败: %v", err)
		}
	}

	return nil
}

// SaveSystemName 保存系统名称
func SaveSystemName(name string) error {
	// 使用事务确保操作原子性
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 先删除旧配置
	_, err = tx.Exec("DELETE FROM system_config")
	if err != nil {
		return fmt.Errorf("删除旧系统配置失败: %v", err)
	}

	// 插入新配置
	stmt, err := tx.Prepare("INSERT INTO system_config (system_name) VALUES (?)")
	if err != nil {
		return fmt.Errorf("准备插入语句失败: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(name)
	if err != nil {
		return fmt.Errorf("插入系统名称失败: %v", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// GetSystemName 获取系统名称
func GetSystemName() (string, error) {
	row := DB.QueryRow("SELECT system_name FROM system_config ORDER BY id DESC LIMIT 1")

	var name string
	err := row.Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return "告警通知系统", nil // 默认名称
		}
		return "", fmt.Errorf("查询系统名称失败: %v", err)
	}

	return name, nil
}

// SaveEmailConfig 保存邮件配置
func SaveEmailConfig(config util.EmailConfig) error {
	// 使用事务确保操作原子性
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 先删除旧配置
	_, err = tx.Exec("DELETE FROM email_config")
	if err != nil {
		return fmt.Errorf("删除旧邮件配置失败: %v", err)
	}

	// 插入新配置
	stmt, err := tx.Prepare("INSERT INTO email_config (smtp_host, smtp_port, username, password, from_email, to_email) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("准备插入语句失败: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(config.SMTPHost, config.SMTPPort, config.Username, config.Password, config.From, config.To)
	if err != nil {
		return fmt.Errorf("插入邮件配置失败: %v", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// GetEmailConfig 获取邮件配置
func GetEmailConfig() (*util.EmailConfig, error) {
	row := DB.QueryRow("SELECT smtp_host, smtp_port, username, password, from_email, to_email FROM email_config ORDER BY id DESC LIMIT 1")

	var config util.EmailConfig
	err := row.Scan(&config.SMTPHost, &config.SMTPPort, &config.Username, &config.Password, &config.From, &config.To)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有配置
		}
		return nil, fmt.Errorf("查询邮件配置失败: %v", err)
	}

	return &config, nil
}

// SaveScriptConfig 保存脚本配置（包括定时任务）
func SaveScriptConfig(config util.ScriptConfig) error {
	// 使用事务确保操作原子性
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 先删除旧配置
	_, err = tx.Exec("DELETE FROM script_config")
	if err != nil {
		return fmt.Errorf("删除旧脚本配置失败: %v", err)
	}

	// 插入新配置
	stmt, err := tx.Prepare("INSERT INTO script_config (path, parameters, timeout, interval) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("准备插入语句失败: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(config.Path, config.Parameters, config.Timeout, config.Interval)
	if err != nil {
		return fmt.Errorf("插入脚本配置失败: %v", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// GetScriptConfig 获取脚本配置
func GetScriptConfig() (*util.ScriptConfig, error) {
	row := DB.QueryRow("SELECT path, parameters, timeout, interval FROM script_config ORDER BY id DESC LIMIT 1")

	var config util.ScriptConfig
	err := row.Scan(&config.Path, &config.Parameters, &config.Timeout, &config.Interval)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有配置
		}
		return nil, fmt.Errorf("查询脚本配置失败: %v", err)
	}

	return &config, nil
}

// SaveScriptReturnConfig 保存脚本返回值配置
func SaveScriptReturnConfig(returnValue int, alertText string) error {
	// 如果告警文本为空，则删除配置
	if alertText == "" {
		_, err := DB.Exec("DELETE FROM script_return_config WHERE return_value = ?", returnValue)
		if err != nil {
			return fmt.Errorf("删除脚本返回值配置失败: %v", err)
		}
		return nil
	}

	// 使用事务确保操作原子性
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 使用UPSERT操作（INSERT OR REPLACE）
	stmt, err := tx.Prepare("INSERT OR REPLACE INTO script_return_config (return_value, alert_text) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("准备插入语句失败: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(returnValue, alertText)
	if err != nil {
		return fmt.Errorf("插入脚本返回值配置失败: %v", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// GetAllScriptReturnConfigs 获取所有脚本返回值配置
func GetAllScriptReturnConfigs() (map[int]string, error) {
	rows, err := DB.Query("SELECT return_value, alert_text FROM script_return_config")
	if err != nil {
		return nil, fmt.Errorf("查询脚本返回值配置失败: %v", err)
	}
	defer rows.Close()

	configs := make(map[int]string)
	for rows.Next() {
		var returnValue int
		var alertText string
		err := rows.Scan(&returnValue, &alertText)
		if err != nil {
			return nil, fmt.Errorf("扫描脚本返回值配置失败: %v", err)
		}
		configs[returnValue] = alertText
	}

	// 检查迭代过程中是否有错误
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历结果时出错: %v", err)
	}

	return configs, nil
}

// SaveMonitorConfig 保存监控配置
func SaveMonitorConfig(config util.MonitorConfig) error {
	// 使用事务确保操作原子性
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 先删除旧配置
	_, err = tx.Exec("DELETE FROM monitor_config")
	if err != nil {
		return fmt.Errorf("删除旧监控配置失败: %v", err)
	}

	// 插入新配置
	stmt, err := tx.Prepare("INSERT INTO monitor_config (interval, avg_count, cpu_threshold, mem_threshold, disk_threshold) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("准备插入语句失败: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(config.Interval, config.AvgCount, config.CPUThreshold, config.MemThreshold, config.DiskThreshold)
	if err != nil {
		return fmt.Errorf("插入监控配置失败: %v", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// GetMonitorConfig 获取监控配置
func GetMonitorConfig() (*util.MonitorConfig, error) {
	row := DB.QueryRow("SELECT interval, avg_count, cpu_threshold, mem_threshold, disk_threshold FROM monitor_config ORDER BY id DESC LIMIT 1")

	var config util.MonitorConfig
	err := row.Scan(&config.Interval, &config.AvgCount, &config.CPUThreshold, &config.MemThreshold, &config.DiskThreshold)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有配置
		}
		return nil, fmt.Errorf("查询监控配置失败: %v", err)
	}

	return &config, nil
}

// SaveSystemStatus 保存系统状态
func SaveSystemStatus(status util.SystemStatus) error {
	// 带重试机制的数据库操作
	var lastErr error
	for i := 0; i < 3; i++ {
		stmt, err := DB.Prepare("INSERT INTO system_status (cpu_usage, mem_usage, disk_usage, timestamp) VALUES (?, ?, ?, ?)")
		if err != nil {
			lastErr = fmt.Errorf("准备插入语句失败: %v", err)
			time.Sleep(time.Millisecond * 100)
			continue
		}
		defer stmt.Close()

		_, err = stmt.Exec(status.CPUUsage, status.MemUsage, status.DiskUsage, status.Timestamp)
		if err != nil {
			stmt.Close()
			lastErr = fmt.Errorf("插入系统状态失败: %v", err)
			time.Sleep(time.Millisecond * 100)
			continue
		}

		return nil
	}

	return lastErr
}

// SaveScriptHistory 保存脚本执行历史
func SaveScriptHistory(result int, output string) error {
	// 带重试机制的数据库操作
	var lastErr error
	for i := 0; i < 3; i++ {
		stmt, err := DB.Prepare("INSERT INTO script_history (result, output) VALUES (?, ?)")
		if err != nil {
			lastErr = fmt.Errorf("准备插入语句失败: %v", err)
			time.Sleep(time.Millisecond * 100)
			continue
		}
		defer stmt.Close()

		_, err = stmt.Exec(result, output)
		if err != nil {
			stmt.Close()
			lastErr = fmt.Errorf("插入脚本执行历史失败: %v", err)
			time.Sleep(time.Millisecond * 100)
			continue
		}

		return nil
	}

	return lastErr
}

// GetLatestSystemStatus 获取最新的系统状态
func GetLatestSystemStatus() (*util.SystemStatus, error) {
	row := DB.QueryRow("SELECT cpu_usage, mem_usage, disk_usage, timestamp FROM system_status ORDER BY timestamp DESC LIMIT 1")

	var status util.SystemStatus
	err := row.Scan(&status.CPUUsage, &status.MemUsage, &status.DiskUsage, &status.Timestamp)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有状态记录
		}
		return nil, fmt.Errorf("查询系统状态失败: %v", err)
	}

	return &status, nil
}

// GetSystemStatusHistory 获取系统状态历史记录
func GetSystemStatusHistory(limit int) ([]util.SystemStatus, error) {
	rows, err := DB.Query("SELECT cpu_usage, mem_usage, disk_usage, timestamp FROM system_status ORDER BY timestamp DESC LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("查询系统状态历史失败: %v", err)
	}
	defer rows.Close()

	var statuses []util.SystemStatus
	for rows.Next() {
		var status util.SystemStatus
		err := rows.Scan(&status.CPUUsage, &status.MemUsage, &status.DiskUsage, &status.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("扫描系统状态失败: %v", err)
		}
		statuses = append(statuses, status)
	}

	// 检查迭代过程中是否有错误
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历结果时出错: %v", err)
	}

	return statuses, nil
}

// ScriptHistory 脚本执行历史结构
type ScriptHistory struct {
	ID        int    `json:"id"`
	Result    int    `json:"result"`
	Output    string `json:"output"`
	CreatedAt string `json:"created_at"`
}

// GetScriptHistory 获取脚本执行历史记录
func GetScriptHistory(limit int) ([]ScriptHistory, error) {
	rows, err := DB.Query("SELECT id, result, output, created_at FROM script_history ORDER BY id DESC LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("查询脚本执行历史失败: %v", err)
	}
	defer rows.Close()

	var histories []ScriptHistory
	for rows.Next() {
		var history ScriptHistory
		err := rows.Scan(&history.ID, &history.Result, &history.Output, &history.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描脚本执行历史失败: %v", err)
		}
		histories = append(histories, history)
	}

	// 检查迭代过程中是否有错误
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历结果时出错: %v", err)
	}

	return histories, nil
}
