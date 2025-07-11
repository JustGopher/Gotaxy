package shell

import (
	"context"
	"fmt"
	"github/JustGopher/Gotaxy/internal/global"
	"github/JustGopher/Gotaxy/internal/storage/models"
	"github/JustGopher/Gotaxy/internal/tunnel/serverCore"
	"github/JustGopher/Gotaxy/pkg/utils"
	"log"
	"strconv"

	"github/JustGopher/Gotaxy/pkg/tlsgen"
)

// shell 全局变量
var shell *Shell

// RegisterCMD 注册命令
func RegisterCMD(sh *Shell) {
	shell = sh
	sh.Register("gen-ca", generateCA)
	sh.Register("gen-certs", generateCerts)
	sh.Register("start", start)
	sh.Register("stop", stop)
	sh.Register("show-config", showConfig)
	sh.Register("show-mapping", showMapping)
	sh.Register("set-ip", setIP)
	sh.Register("set-port", setPort)
	sh.Register("set-email", setEmail)
	sh.Register("add-mapping", addMapping)
	sh.Register("del-mapping", delMapping)
	sh.Register("upd-mapping", updMapping)
}

// start 启动服务端
// 格式：start
func start(args []string) {
	// 检查证书是否存在
	if !tlsgen.CheckServerCertExist("certs") {
		fmt.Println("证书缺失，请先执行生成证书")
		return
	}
	global.Ctx, global.Cancel = context.WithCancel(context.Background())
	go serverCore.StartServer(global.Ctx)
}

// stop 停止服务端
// 格式：stop
func stop(args []string) {
	global.Cancel()
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
				log.Println("shellcmd.Rl.Readline() 读取输入失败:", err)
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
	err := tlsgen.GenerateServerAndClientCerts(global.Config.ServerIP, "certs", day, "certs/ca.crt", "certs/ca.key")
	if err != nil {
		log.Println("generateCerts() 生成证书失败:", err)
		return
	}
}

// showConfig 显示服务端配置
// 格式：show-config
func showConfig(args []string) {
	fmt.Println(" IP         ：", global.Config.ServerIP)
	fmt.Println(" ListenPort ：", global.Config.ListenPort)
	fmt.Println(" Email      ：", global.Config.Email)
}

func showMapping(args []string) {
	mpg, err := models.GetAllMpg(global.DB)
	if err != nil {
		return
	}

	fmt.Println("Name\tPublicPort\tTargetAddr\t\tStatus")

	for _, v := range mpg {
		fmt.Println(v.Name, "\t", v.PublicPort, "\t\t", v.TargetAddr, "\t", v.Status)
	}
}

// setIP 设置服务端 IP
// 格式：set-ip [ip]
func setIP(args []string) {
	// 校验数量
	length := len(args)
	if length == 0 {
		fmt.Println("参数不能为空，正确格式为：set-ip <ip>")
		return
	}
	if length != 1 {
		fmt.Printf("无效的参数 '%s'，正确格式为：set-ip <ip>\n", args)
		return
	}
	// 解析参数
	ip := args[0]
	if ip == "" {
		fmt.Println("IP地址不能为空")
		return
	}
	// 使用正则表达式验证 IP 格式
	if !utils.IsValidateIP(ip) {
		fmt.Println("IP地址格式不正确")
		return
	}
	// 设置 IP
	global.Config.ServerIP = ip
	err := models.UpdateCfg(global.DB, "server_ip", ip)
	if err != nil {
		return
	}
}

// setPort 设置服务端 Post
// 格式：set-port [port]
func setPort(args []string) {
	length := len(args)
	if length != 1 {
		fmt.Printf("无效的参数 '%s'，正确格式为：set-port <port>\n", args)
		return
	}

	port, err := strconv.Atoi(args[0])
	if err != nil || port <= 0 || port > 65535 {
		fmt.Printf("无效的参数 '%s'，参数必须是1-65535范围内的数字！\n", args)
		return
	}

	global.Config.ListenPort = args[0]
	err = models.UpdateCfg(global.DB, "listen_port", args[0])
	if err != nil {
		return
	}
}

// setEmail 设置服务端 Email
// 格式：set-email [email]
func setEmail(args []string) {
	length := len(args)
	if length != 1 {
		fmt.Printf("无效的参数 '%s'，正确格式为：set-email <email>\n", args)
		return
	}

	if args[0] == "" {
		fmt.Println("Email地址不能为空")
		return
	}

	if !utils.IsValidateEmail(args[0]) {
		fmt.Println("Email地址格式不正确")
		return
	}

	global.Config.Email = args[0]
	err := models.UpdateCfg(global.DB, "email", args[0])
	if err != nil {
		return
	}
}

// setMapping 设置映射
// 格式：set-mapping [name] [public_port] [target_addr]
func addMapping(args []string) {
	length := len(args)
	if length != 3 {
		fmt.Printf("无效的参数 '%s'，正确格式为：set-mapping <name> <public_port> <target_addr>\n", args)
		return
	}

	if args[0] == "" || args[1] == "" || args[2] == "" {
		fmt.Println("参数缺失!，正确格式为：set-mapping <name> <public_port> <target_addr>")
		return
	}

	port, err := strconv.Atoi(args[1])
	if err != nil || port <= 0 || port > 65535 {
		fmt.Printf("无效的参数 '%s'，参数必须是1-65535范围内的数字！\n", args)
		return
	}

	if !utils.IsValidateAddr(args[2]) {
		fmt.Println("目标地址格式不正确")
		return
	}

	err = models.InsertMpg(global.DB, models.Mapping{
		Name:       args[0],
		PublicPort: args[1],
		TargetAddr: args[2],
		Status:     "close",
	})
	if err != nil {
		return
	}
	global.ConnPool.Set(args[0], args[1], args[2])
}

func delMapping(args []string) {
	if len(args) != 1 {
		fmt.Printf("无效的参数 '%s'，正确格式为：del-mapping <name>\n", args)
	}

	if args[0] == "" {
		fmt.Println("参数缺失!，正确格式为：del-mapping <name>")
		return
	}

	err := models.DeleteMapByName(global.DB, args[0])
	if err != nil {
		return
	}
}

func updMapping(args []string) {
	if len(args) != 3 {
		fmt.Printf("无效的参数 '%s'，正确格式为：upd-mapping <name> <port> <addr>\n", args)
		return
	}

	if args[0] == "" {
		fmt.Println(" name 不能为空！")
		return
	}

	port, err := strconv.Atoi(args[1])
	if err != nil || port <= 0 || port > 65535 {
		fmt.Printf("无效的参数 '%s'，参数必须是1-65535范围内的数字！\n", args)
		return
	}

	if !utils.IsValidateAddr(args[2]) {
		fmt.Println("目标地址格式不正确")
		return
	}

	_, err = models.UpdateMap(global.DB, args[0], args[1], args[2])
	if err != nil {
		fmt.Println(err)
		return
	}
}
