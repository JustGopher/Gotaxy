package shellCore

import (
	"fmt"
	"github/JustGopher/Gotaxy/pkg/shellcli"
	"github/JustGopher/Gotaxy/pkg/tlsgen"

	"log"
	"strconv"
)

// shell 全局变量
var shell *shellcli.Shell

// Register 注册命令
func Register(sh *shellcli.Shell) {
	shell = sh
	sh.Register("gen-ca", generateCA)
	sh.Register("gen-certs", generateCerts)
}

// generateCA 生成 CA 证书
// 格式：gen-ca [有效期(年)] [-overwrite]
func generateCA(args []string) {
	year := 10
	overwrite := false
	length := len(args)
	if length > 2 {
		fmt.Printf("无效的参数 '%s'，正确格式为：gen-ca [有效期] [-overwrite]\n", args)
		return
	}
	// 如果为一个参数
	if length == 1 {
		// 判断是整数还是 -overwrite
		input := args[0]
		if input == "-overwrite" {
			overwrite = true
		} else if d, err := strconv.Atoi(input); err == nil && d > 0 {
			if err != nil || d <= 0 {
				fmt.Printf("无效的有效期参数 '%s'，请传入正整数，例如：gen-ca 10\n", input)
				return
			}
			year = d
		} else {
			fmt.Printf("无效的参数 '%s'，正确格式为：gen-ca [正整数] [-overwrite]\n", input)
			return
		}
	}
	// 如果为两个参数
	if length == 2 {
		// 第一个参数为整数
		if d, err := strconv.Atoi(args[0]); err == nil && d > 0 {
			if err != nil || d <= 0 {
				fmt.Printf("无效的参数 '%s'，正确格式为：gen-ca [正整数] [-overwrite]\n", args[0])
				return
			}
			year = d
		} else {
			fmt.Printf("无效的参数 '%s'，正确格式为：gen-ca [正整数] [-overwrite]\n", args[0])
			return
		}
		// 第二个参数为 -overwrite
		if args[1] != "-overwrite" {
			fmt.Printf("无效的参数 '%s'，正确格式为：gen-ca [正整数] [-overwrite]\n", args[1])
			return
		} else {
			overwrite = true
		}
	}
	// 询问是否确定重新生成 CA 证书
	if overwrite {
		for {
			fmt.Printf("确定要重新生成 CA 证书吗？(y/n) \n")
			readline, err := shell.Rl.Readline()
			if err != nil {
				log.Println("shell.Rl.Readline() 读取输入失败:", err)
				return
			}
			if readline == "n" {
				fmt.Println("已取消重新生成 CA 证书")
				return
			} else if readline == "y" {
				break
			} else {
				fmt.Println("无效的输入，请输入 'y' 或 'n'")
				continue
			}
		}
	}
	// 生成 CA 证书
	err := tlsgen.GenerateCA("certs", year, overwrite)
	if err != nil {
		log.Println("generateCA() 生成 CA 证书失败:", err)
		return
	}
}

// generateCerts 生成 server 和 client 证书
// 格式：gen-certs [有效期(日)]
func generateCerts(args []string) {
	// 默认天数
	day := 10

	// 校验数量
	length := len(args)
	if length > 1 {
		fmt.Printf("无效的参数 '%s'，正确格式为：gen-certs [正整数]\n", args)
		return
	}

	// 解析参数
	if len(args) == 1 {
		d, err := strconv.Atoi(args[0])
		if err != nil || d <= 0 {
			fmt.Printf("无效的参数 '%s'，正确格式为：gen-certs [正整数]\n", args[0])
			return
		}
		day = d
	}

	// 生成证书
	err := tlsgen.GenerateServerAndClientCerts("certs", day, "certs/ca.crt", "certs/ca.key")
	if err != nil {
		log.Println("generateCerts() 生成证书失败:", err)
		return
	}
}
