package serverCore

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github/JustGopher/Gotaxy/internal/heart"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github/JustGopher/Gotaxy/internal/global"

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

	global.Ring = heart.NewHeartbeatRing(20)

	<-ctx.Done()
	fmt.Println("收到退出信号，停止中...")

	// 主动关闭当前会话
	// 从 atomic.Value 中取出当前活跃的 session, 调用 session.Close() 主动关闭连接，防止资源泄露, 这会通知所有基于此会话创建的 stream 关闭。
	if val := currentSession.Load(); val != nil {
		session := val.(*smux.Session)
		_ = session.Close()
	}
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

		go startHeartbeat(session, 5*time.Second, global.Ring)

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

		// OpenStream() 的作用是：在当前控制连接上创建一个逻辑流（可以理解为一条虚拟 TCP 通道）
		// smux 会把多个这样的 stream 数据复用在一条实际连接上
		stream, err := session.OpenStream()
		if err != nil {
			global.Log.Error("session.OpenStream() smux stream创建失败: ", err)
			fmt.Printf("smux stream 创建失败: %v", err)
			_ = publicConn.Close()
			continue
		}

		// 通知客户端目标地址
		// 在流建立后，服务端会先向客户端写一行数据：告诉它，目标地址是哪个（如 127.0.0.1:8080）。
		// 客户端拿到这个地址后，就会在自己的本地去连接这个目标地址，然后形成一对 stream <-> 内网本地服务连接 的代理。
		_, err = stream.Write([]byte("DIRECT\n" + target + "\n"))
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
		_ = dst.Close()
	}(dst)
	defer func(src net.Conn) {
		_ = src.Close()
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

// startHeartbeat 启动心跳检测
// 每 interval 时间间隔向客户端发送 HEARTBEAT 报文，客户端收到后回复 PONG 报文
func startHeartbeat(session *smux.Session, interval time.Duration, ring *heart.HeartbeatRing) {
	go func() {
		for {
			time.Sleep(2 * time.Second)

			stream, err := session.OpenStream()
			if err != nil {
				global.Log.Error("heartbeat: OpenStream失败:", err)
				ring.Add(false, 0)
				continue
			}

			start := time.Now()
			_, err = stream.Write([]byte("HEARTBEAT\nPING\n"))
			if err != nil {
				global.Log.Warn("heartbeat: 写入失败:", err)
				ring.Add(false, 0)
				_ = stream.Close()
				continue
			}

			buffer := make([]byte, 4)
			stream.SetReadDeadline(time.Now().Add(5 * time.Second))
			_, err = io.ReadFull(stream, buffer)
			if err != nil || string(buffer) != "PONG" {
				global.Log.Error("heartbeat: 读失败:", err)
				ring.Add(false, 0)
				_ = stream.Close()
				continue
			}

			delay := time.Since(start)
			ring.Add(true, delay)
			global.Log.Info("收到pong,delay:", delay)
			_ = stream.Close()

			time.Sleep(interval)
		}
	}()
}
