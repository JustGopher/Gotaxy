package serverCore

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github/JustGopher/Gotaxy/internal/global"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync/atomic"

	"github.com/xtaci/smux"
)

var portMap = map[string]string{}

// 当前活跃 session （用 atomic.Value 可原子替换）
var currentSession atomic.Value

// StartServer 启动服务
func StartServer(ctx context.Context) {
	pool := global.ConnPool
	if pool == nil {
		global.Log.Info("StartServer() 连接池未初始化")
		panic("StartServer() 连接池未初始化")
	}
	allPortMap := pool.GetAllPort()
	for port, add := range allPortMap {
		portMap[port] = add
	}

	go waitControlConn(ctx)

	for pubPort := range portMap {
		go startPublicListener(ctx, pubPort)
	}
	<-ctx.Done()
	fmt.Println("收到退出信号，停止中...")
}

// 不断接受控制连接
func waitControlConn(ctx context.Context) {
	tlsCfg, err := LoadServerTLSConfig("certs/server.crt", "certs/server.key", "certs/ca.crt")
	if err != nil {
		global.Log.Error("waitControlConn() 加载 TLS 配置失败: ", err)
		panic("加载 TLS 配置失败: " + err.Error())
	}

	listener, err := tls.Listen("tcp", ":"+global.Config.ListenPort, tlsCfg)
	if err != nil {
		global.Log.Error("waitControlConn() 监听失败: ", err)
		panic("监听失败: " + err.Error())
	}
	fmt.Printf("控制端口监听 %s 端口中...\n", global.Config.ListenPort)

	go func() {
		<-ctx.Done()
		fmt.Println("关闭控制连接监听")
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return // 正常退出
			default:
				fmt.Println("控制连接接入失败:", err)
				global.Log.Error("控制连接接入失败:", err)
				continue
			}
		}

		session, err := smux.Server(conn, nil)
		if err != nil {
			fmt.Println("创建会话失败:", err)
			global.Log.Error("创建会话失败:", err)
			_ = conn.Close()
			continue
		}

		global.Log.Info("会话建立成功")
		currentSession.Store(session)
	}
}

// 持续监听公网端口流量，建立 stream
func startPublicListener(ctx context.Context, pubPort string) {
	listener, err := net.Listen("tcp", ":"+pubPort)
	if err != nil {
		if strings.Contains(err.Error(), "address already in use") {
			log.Printf("端口 %s 已被占用", pubPort)
			return
		}
		fmt.Printf("监听端口 %s 失败: %v\n", pubPort, err)
		return
	}
	target := portMap[pubPort]
	log.Printf("监听端口 %s 映射到客户端 %s\n", pubPort, target)

	go func() {
		<-ctx.Done()
		global.Log.Info("关闭公网端口监听 :", pubPort)
		fmt.Printf("关闭公网端口监听 :%s", pubPort)
		_ = listener.Close()
	}()

	for {
		publicConn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				global.Log.Error("listener.Accept() 连接失败:", err)
				fmt.Printf("连接失败: %v", err)
				continue
			}
		}
		// 正常 smux 流转发
		sessionVal := currentSession.Load()
		if sessionVal == nil {
			fmt.Println("无有效客户端连接，关闭连接")
			_ = publicConn.Close()
			continue
		}
		session, _ := sessionVal.(*smux.Session)

		stream, err := session.OpenStream()
		if err != nil {
			global.Log.Error("session.OpenStream() smux stream创建失败: ", err)
			fmt.Printf("smux stream 创建失败: %v", err)
			_ = publicConn.Close()
			continue
		}

		// 通知客户端目标地址
		_, err = stream.Write([]byte(target + "\n"))
		if err != nil {
			global.Log.Error("写入目标地址失败:", err)
			_ = publicConn.Close()
			_ = stream.Close()
			continue
		}

		fmt.Printf("建立转发: 端口 %s <=> 客户端本地 %s", pubPort, target)
		go proxy(publicConn, stream)
		go proxy(stream, publicConn)
	}
}

// proxy 数据转发
func proxy(dst, src net.Conn) {
	defer func(dst net.Conn) {
		err := dst.Close()
		if err != nil {
			global.Log.Info("proxy() 关闭连接失败: ", err)
		}
	}(dst)
	defer func(src net.Conn) {
		err := src.Close()
		if err != nil {
			global.Log.Info("proxy() 关闭连接失败: ", err)
		}
	}(src)
	_, _ = io.Copy(dst, src)
}

// LoadServerTLSConfig 加载服务端 TLS 配置（含双向认证）
func LoadServerTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	// 加载服务端证书
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("加载服务端证书失败: %w", err)
	}

	// 加载 CA 证书用于校验客户端证书
	caCertPEM, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("读取 CA 文件失败: %w", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCertPEM) {
		return nil, fmt.Errorf("解析 CA 证书失败")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert, // 开启 mTLS
		ClientCAs:    caCertPool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}
