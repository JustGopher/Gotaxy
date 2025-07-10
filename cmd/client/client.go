package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/xtaci/smux"
)

// HelloServe 测试服务
func HelloServe() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		write, err := w.Write([]byte("Hello, World!"))
		if err != nil {
			return
		}
		fmt.Println(write)
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func start(serverAddr string) {
	go HelloServe()

	tlsCfg, err := LoadClientTLSConfig("certs/client.crt", "certs/client.key", "certs/ca.crt")
	if err != nil {
		log.Fatalf("加载 TLS 配置失败: %v", err)
	}

	conn, err := tls.Dial("tcp", serverAddr, tlsCfg)
	if err != nil {
		log.Fatalf("连接服务端失败（TLS）: %v", err)
	}
	log.Println("已通过 TLS 连接服务端")

	session, err := smux.Client(conn, nil)
	if err != nil {
		log.Fatalf("创建 smux 客户端会话失败: %v", err)
	}
	log.Println("smux 会话创建成功")

	for {
		stream, err := session.AcceptStream()
		if err != nil {
			log.Println("接受 stream 失败:", err)
			return
		}
		go handleStream(stream)
	}
}

func handleStream(stream *smux.Stream) {
	reader := bufio.NewReader(stream)
	target, err := reader.ReadString('\n')
	if err != nil {
		log.Println("读取目标地址失败:", err)
		_ = stream.Close()
		return
	}
	target = target[:len(target)-1] // 去除换行

	localConn, err := net.Dial("tcp", target)
	if err != nil {
		log.Printf("连接本地服务 %s 失败: %v", target, err)
		_ = stream.Close()
		return
	}

	log.Printf("转发连接 %s <=> %s", stream.RemoteAddr(), target)
	go proxy(stream, localConn)
	go proxy(localConn, stream)
}

func proxy(dst, src net.Conn) {
	defer func(dst net.Conn) {
		err := dst.Close()
		if err != nil {
			log.Printf("proxy() 关闭连接失败: %v", err)
		}
	}(dst)
	defer func(src net.Conn) {
		err := src.Close()
		if err != nil {
			log.Printf("proxy() 关闭连接失败: %v", err)
		}
	}(src)
	_, _ = io.Copy(dst, src)
}

// LoadClientTLSConfig 客户端 TLS 配置（支持 mTLS）
func LoadClientTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	// 客户端证书（client.crt + client.key）
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("加载客户端证书失败: %w", err)
	}

	// 加载 CA 根证书
	caCertPEM, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("读取 CA 根证书失败: %w", err)
	}
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCertPEM) {
		return nil, fmt.Errorf("解析 CA 根证书失败")
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caPool,
		InsecureSkipVerify: false,
	}, nil
}

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

			start(host + ":" + fmt.Sprintf("%d", port))
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

func showHelp() {
	fmt.Println(`Usage:
  go run cmd/client/client.go -h [host] -p <port>

Options:
  -h [host]     
        The hostname or IP address of the server (default "127.0.0.1")
  -p <port>
        The port number to connect to (default 9000)`)
}

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
