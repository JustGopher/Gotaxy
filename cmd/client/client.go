package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github/JustGopher/Gotaxy/internal/tunnel/clientCore"

	"github.com/urfave/cli/v2"
)

func main() {
	// 检查是否有--help参数
	for _, arg := range os.Args {
		if arg == "--help" || arg == "-help" {
			showHelp()
			os.Exit(0)
		}
	}

	// 参数检查函数
	checkArgs()

	app := &cli.App{
		Name:      "client",
		Usage:     "内网穿透客户端",
		HelpName:  "client",
		ArgsUsage: "-h [host] -p [port]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "h",
				Value: "127.0.0.1",
				Usage: "The hostname or IP address of the server",
			},
			&cli.IntFlag{
				Name:  "p",
				Value: 9000,
				Usage: "The port number to connect to",
			},
		},
		Action: func(c *cli.Context) error {
			host := c.String("h")
			port := c.Int("p")

			fmt.Printf("解析后的参数: host=%s, port=%d\n", host, port)

			clientCore.Start(host + ":" + fmt.Sprintf("%d", port))
			return nil
		},
		// 自定义命令行参数错误处理
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			fmt.Printf("错误: %v\n", err)
			showHelp()
			return err
		},
	}

	// 禁用默认的帮助标志和命令
	app.HideHelp = true
	app.HideHelpCommand = true

	// 完全覆盖默认的帮助模板
	cli.AppHelpTemplate = ""
	cli.CommandHelpTemplate = ""
	cli.SubcommandHelpTemplate = ""

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

// 显示帮助信息
func showHelp() {
	fmt.Println(`Usage:
  go run cmd/client/client.go -h [host] -p <port>

Options:
  -h [host]     
        The hostname or IP address of the server (default "127.0.0.1")
  -p <port>
        The port number to connect to (default 9000)`)
}

// 检查参数是否正确
func checkArgs() {
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "-h" || arg == "-p" {
			if i+1 >= len(os.Args) {
				fmt.Println("错误：参数后面缺少值")
				showHelp()
				os.Exit(1)
			} else if strings.HasPrefix(os.Args[i+1], "-") {
				fmt.Println("错误：参数后面缺少值")
				showHelp()
				os.Exit(1)
			}
		}
	}
}
