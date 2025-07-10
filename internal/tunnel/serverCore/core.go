package serverCore

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github/JustGopher/Gotaxy/internal/tunnel/serverCore/global"
	"io"
	"log"
	"net"
	"os"
	"sync/atomic"

	"github.com/xtaci/smux"
)

var listenPort = "9000"

var portMap = map[string]string{}

// 当前活跃 session （用 atomic.Value 可原子替换）
var currentSession atomic.Value

// StartServer 启动服务
func StartServer(ctx context.Context) {
	pool := global.ConnPool
	if pool == nil {
		log.Fatal("连接池未初始化")
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
	log.Println("服务端收到退出信号，停止中...")
}

// 不断接受控制连接
func waitControlConn(ctx context.Context) {
	tlsCfg, err := LoadServerTLSConfig("certs/server.crt", "certs/server.key", "certs/ca.crt")
	if err != nil {
		log.Fatalf("加载 TLS 配置失败: %v", err)
	}

	listener, err := tls.Listen("tcp", ":"+listenPort, tlsCfg)
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}
	log.Printf("控制端口监听 :%s 中...\n", listenPort)

	go func() {
		<-ctx.Done()
		log.Println("关闭控制连接监听")
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return // 正常退出
			default:
				log.Println("控制连接接入失败:", err)
				continue
			}
		}

		session, err := smux.Server(conn, nil)
		if err != nil {
			log.Println("创建 smux 会话失败:", err)
			_ = conn.Close()
			continue
		}

		log.Println("smux 会话建立成功")
		currentSession.Store(session)
	}
}

// 持续监听公网端口流量，建立 stream
func startPublicListener(ctx context.Context, pubPort string) {
	listener, err := net.Listen("tcp", ":"+pubPort)
	if err != nil {
		log.Fatalf("监听公网端口 %s 失败: %v", pubPort, err)
	}
	target := portMap[pubPort]
	log.Printf("公网监听 :%s 映射到客户端内网 %s\n", pubPort, target)

	go func() {
		<-ctx.Done()
		log.Printf("关闭公网端口监听 :%s", pubPort)
		_ = listener.Close()
	}()

	for {
		publicConn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Printf("公网连接失败: %v", err)
				continue
			}
		}
		// 正常 smux 流转发
		sessionVal := currentSession.Load()
		if sessionVal == nil {
			log.Println("无有效客户端连接，关闭连接")
			_ = publicConn.Close()
			continue
		}
		session, _ := sessionVal.(*smux.Session)

		stream, err := session.OpenStream()
		if err != nil {
			log.Printf("smux stream 创建失败: %v", err)
			_ = publicConn.Close()
			continue
		}

		// 通知客户端目标地址
		_, err = stream.Write([]byte(target + "\n"))
		if err != nil {
			log.Println("写入目标地址失败:", err)
			_ = publicConn.Close()
			_ = stream.Close()
			continue
		}

		log.Printf("建立转发: 公网 :%s <=> 客户端本地 %s", pubPort, target)
		go proxy(publicConn, stream)
		go proxy(stream, publicConn)
	}
}

// proxy 数据转发
func proxy(dst, src net.Conn) {
	defer func(dst net.Conn) {
		err := dst.Close()
		log.Printf("proxy() 关闭连接失败: %v", err)
	}(dst)
	defer func(src net.Conn) {
		err := src.Close()
		if err != nil {
			log.Printf("proxy() 关闭连接失败: %v", err)
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
