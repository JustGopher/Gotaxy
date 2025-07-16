package shell

import (
	"context"
	"fmt"
	"github/JustGopher/Gotaxy/internal/tunnel/proxy"
	"log"
	"strconv"

	"github/JustGopher/Gotaxy/internal/global"
	"github/JustGopher/Gotaxy/internal/storage/models"
	"github/JustGopher/Gotaxy/internal/tunnel/serverCore"
	"github/JustGopher/Gotaxy/pkg/tlsgen"
	"github/JustGopher/Gotaxy/pkg/utils"
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
	sh.Register("add-mapping", AddMapping)
	sh.Register("del-mapping", DelMapping)
	sh.Register("upd-mapping", UpdMapping)
	sh.Register("heart", Heart)
	sh.Register("open-mapping", OpenMapping)
	sh.Register("close-mapping", CloseMapping)
}

// OpenMapping 打开映射
// 格式：open-mapping [映射名称]
func OpenMapping(args []string) {
	/**
	1. 如果服务已启动，检查是否关闭，若未关闭，启动这个映射
	2. 如果服务未启动，仅仅变动是否打开设置
	*/
	if len(args) != 1 {
		fmt.Printf("无效的参数 '%s'，正确格式为：open-mapping [映射名称]\n", args)
		return
	}
	name := args[0]

	if global.IsRun == false {
		ok := global.ConnPool.UpdateEnable(name, true)
		if !ok {
			fmt.Printf("规则 '%s' 不存在\n", name)
			return
		}
		mapping := global.ConnPool.GetMapping(name)
		var enable string
		if mapping.Enable == true {
			enable = "open"
		} else {
			enable = "close"
		}
		updateMap, err := models.UpdateMap(global.DB, mapping.Name, mapping.PublicPort, mapping.TargetAddr, enable)
		if err != nil {
			global.Log.Error("OpenMapping() 修改规则失败", err)
			return
		}
		global.Log.Info("OpenMapping() 修改 '%s' 成功", updateMap.Name)
	} else {
		// 启动映射
		ok := global.ConnPool.UpdateEnable(name, true)
		if !ok {
			fmt.Printf("规则 '%s' 不存在\n", name)
			return
		}
		mapping := global.ConnPool.GetMapping(name)
		// 启动映射
		mapping.Ctx, mapping.CtxCancel = context.WithCancel(context.Background())
		go proxy.StartPublicListener(global.Ctx, mapping)
	}
}

// CloseMapping 关闭映射
// 格式：close-mapping [映射名称]
func CloseMapping(args []string) {
	if len(args) != 1 {
		fmt.Printf("无效的参数 '%s'，正确格式为：close-mapping [映射名称]\n", args)
		return
	}
	name := args[0]
	err := global.ConnPool.Close(name)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("关闭 '%s' 成功", name)
}

// start 启动服务端
// 格式：start
func start(args []string) {
	if global.IsRun == true {
		fmt.Println("服务已启动")
		return
	}
	// 检查证书是否存在
	if !tlsgen.CheckServerCertExist("certs") {
		fmt.Println("证书缺失，请先生成证书")
		return
	}
	global.Ctx, global.Cancel = context.WithCancel(context.Background())
	go serverCore.StartServer(global.Ctx)
	global.IsRun = true
}

// stop 停止服务端
// 格式：stop
func stop(args []string) {
	if global.IsRun == false {
		fmt.Println("服务未启动")
		return
	}
	global.Cancel()
	global.IsRun = false
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
		} else if d, err := strconv.Atoi(input); err == nil {
			if d <= 0 {
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
		if d, err := strconv.Atoi(args[0]); err == nil {
			if d <= 0 {
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
				global.Log.Errorf("generateCA() shellcmd.Rl.Readline() 读取输入失败: %v", err)
				fmt.Println("读取输入失败:", err)
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
		global.Log.Errorf("generateCA() 生成 CA 证书失败: %v", err)
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
		global.Log.Errorf("generateCerts() 生成证书失败: %v", err)
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

// showMapping 显示映射
// 格式：show-mapping
func showMapping(args []string) {
	mpg := global.ConnPool.All()

	fmt.Println("Name\tPublicPort\tTargetAddr\t\tStatus\t\tEnable")

	for _, v := range mpg {
		fmt.Println(v.Name, "\t", v.PublicPort, "\t\t", v.TargetAddr, "\t", v.Status, "\t", v.Enable)
	}
}

// setIP 设置服务端 IP
// 格式：set-ip <ip>
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
		global.Log.Errorf("setIP() 更新配置数据失败: %v", err)
		fmt.Println("更新配置数据失败:", err)
		return
	}
}

// setPort 设置服务端 Post
// 格式：set-port <port>
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
// 格式：set-email <email>
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
		global.Log.Errorf("setEmail() 更新配置数据失败: %v", err)
		fmt.Println("更新配置数据失败:", err)
		return
	}
}

// AddMapping 设置映射
// 格式：add-mapping <name> <public_port> <target_addr>
func AddMapping(args []string) {
	length := len(args)
	if length != 3 {
		fmt.Printf("无效的参数 '%s'，正确格式为：add-mapping <name> <public_port> <target_addr>\n", args)
		return
	}

	if args[0] == "" || args[1] == "" || args[2] == "" {
		fmt.Println("参数缺失!，正确格式为：add-mapping <name> <public_port> <target_addr>")
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
		Enable:     "close",
	})
	if err != nil {
		global.Log.Errorf("addMapping() 插入映射数据失败: %v", err)
		fmt.Println("插入映射数据失败:", err)
		return
	}
	global.ConnPool.Set(args[0], args[1], args[2], false)
}

// DelMapping 删除映射
// 格式：del-mapping <name>
func DelMapping(args []string) {
	if len(args) != 1 {
		fmt.Printf("无效的参数 '%s'，正确格式为：del-mapping <name>\n", args)
	}

	if args[0] == "" {
		fmt.Println("参数缺失!，正确格式为：del-mapping <name>")
		return
	}
	mpg := global.ConnPool.GetMapping(args[0])
	if mpg == nil {
		fmt.Println("映射不存在，请检查name是否正确")
		return
	}

	if mpg.Status == "active" {
		fmt.Println("当前映射正在运行中，无法删除，请关闭后重试")
		return
	}

	global.ConnPool.Delete(mpg.Name)

	err := models.DeleteMapByName(global.DB, args[0])
	if err != nil {
		global.Log.Errorf("delMapping() 删除映射数据失败: %v", err)
		fmt.Println("删除映射数据失败:", err)
		return
	}
}

// UpdMapping 更新映射
// 格式：upd-mapping <name> <port> <addr>
func UpdMapping(args []string) {
	if len(args) != 4 {
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

	_, err = models.UpdateMap(global.DB, args[0], args[1], args[2], args[3])
	if err != nil {
		global.Log.Errorf("updMapping() 更新映射数据失败: %v", err)
		fmt.Println("更新映射数据失败:", err)
		return
	}
}

func Heart(args []string) {
	fmt.Println(global.Ring.Status(global.IsRun))
}
