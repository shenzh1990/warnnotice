package util

import (
	"time"
)

func UnTimeTOISOTime(untime int64) string {
	println(untime)
	t := time.UnixMilli(untime)
	return t.Format(time.RFC3339) // RFC3339是ISO 8601的一个子集 输出：2021-01-01T00:00:00Z
}
func ISOTimeTOUnTime(untime string) (int64, error) {
	t, err := time.Parse(time.RFC3339, untime)
	if err != nil {
		return 0, err
	}
	return t.UnixMilli(), nil // 输出：1609459200000
}
func GetNow() time.Time {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err == nil {
		return time.Now().In(location)
	}
	return time.Now()
}
