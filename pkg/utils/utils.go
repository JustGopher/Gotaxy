package utils

import "regexp"

// IsValidateIP 验证 IP 是否合法
func IsValidateIP(ip string) bool {
	// 通过正则表达式验证 IP 地址
	s := `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	reg := regexp.MustCompile(s)
	if reg.MatchString(ip) {
		return true
	} else {
		return false
	}
}
