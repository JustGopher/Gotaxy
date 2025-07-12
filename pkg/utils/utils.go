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

// IsValidateEmail 验证 Email 是否合法
func IsValidateEmail(email string) bool {
	// 通过正则表达式验证 Email 地址
	s := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	reg := regexp.MustCompile(s)
	if reg.MatchString(email) {
		return true
	} else {
		return false
	}
}

func IsValidateAddr(addr string) bool {
	// 通过正则表达式验证 IP 地址
	s := `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?):[0-9]{1,5}$`
	reg := regexp.MustCompile(s)
	if reg.MatchString(addr) {
		return true
	} else {
		return false
	}
}
