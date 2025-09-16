package util

import (
	"fmt"
	"regexp"
)

// 密码强度等级，D为最低
const (
	levelD = iota
	LevelC
	LevelB
	LevelA
	LevelS
)

/*
  - minLength: 指定密码的最小长度
  - maxLength：指定密码的最大长度

 *  minLevel：指定密码最低要求的强度等级
 *  pwd：明文密码
*/
func VerifyPwdFormat(minLength, maxLength, minLevel int, pwd string) error {
	// 首先校验密码长度是否在范围内
	if len(pwd) < minLength {
		return fmt.Errorf("密码长度必须大于%d！", minLength)
	}
	if len(pwd) > maxLength {
		return fmt.Errorf("密码长度必须小于%d！", maxLength)
	}

	// 初始化密码强度等级为D，利用正则校验密码强度，若匹配成功则强度自增1
	var level int = levelD
	patternList := []string{`[0-9]+`, `[a-z]+`, `[A-Z]+`}
	for _, pattern := range patternList {
		match, _ := regexp.MatchString(pattern, pwd)
		if match {
			level++
		}
	}

	// 如果最终密码强度低于要求的最低强度，返回并报错
	if level < minLevel {
		return fmt.Errorf("密码必须至少包含数字和大、小字母两种以上")
	}
	return nil
}
func VerifyEmailFormat(email string) bool {
	//pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	pattern := `^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`

	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// mobile verify
func VerifyMobileFormat(mobileNum string) bool {
	regular := "^((13[0-9])|(14[5,7])|(15[0-3,5-9])|(17[0,3,5-8])|(18[0-9])|166|198|199|(147))\\d{8}$"

	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}
