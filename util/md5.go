package util

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

func Md5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data)) // 需要加密的字符串为 123456
	cipherStr := h.Sum(nil)

	return hex.EncodeToString(cipherStr)
}

func MD5Encode(data string) string {
	return strings.ToUpper(Md5Encode(data))
}

// 校验用户密码
func ValidatePasswd(plainpwd, salt, passwd string) bool {
	return Md5Encode(plainpwd+salt) == passwd
}

// 生成用户密码
func MakePasswd(plainpwd, salt string) string {
	return Md5Encode(plainpwd + salt)
}
func HmacSHA256Base64Sign(secret, params string) (string, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	_, err := mac.Write([]byte(params))
	if err != nil {
		return "", err
	}
	signByte := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(signByte), nil
}

// 辅助函数
func ParseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
