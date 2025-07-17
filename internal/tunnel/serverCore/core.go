package serverCore

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github/JustGopher/Gotaxy/internal/heart"
	"github/JustGopher/Gotaxy/internal/tunnel/proxy"
	"io"
	"os"
	"time"

	"github/JustGopher/Gotaxy/internal/global"

	"github.com/xtaci/smux"
)

// StartServer 启动服务
func StartServer(ctx context.Context) {
	connPool := global.ConnPool
	if connPool == nil {
		global.Log.Info("StartServer() 连接池未初始化")
		panic("StartServer() 连接池未初始化")
	}
	// 开启控制端口监听
	go waitControlConn(ctx)

	// 开启穿透端口监听
	allPortMap := connPool.All()
	for _, mapping := range allPortMap {
		if mapping.Enable {
			mapping.Ctx, mapping.CtxCancel = context.WithCancel(context.Background())
			go proxy.StartPublicListener(ctx, mapping)
		}
	}

	global.Ring = heart.NewHeartbeatRing(20)

	<-ctx.Done()
	fmt.Println("收到退出信号，停止中...")

	// 主动关闭当前会话
	// 从 atomic.Value 中取出当前活跃的 session, 调用 session.Close() 主动关闭连接，防止资源泄露, 这会通知所有基于此会话创建的 stream 关闭。
	if session := connPool.GetSession(); session != nil {
		_ = session.Close()
	}
}

// 不断接受控制连接
func waitControlConn(ctx context.Context) {
	tlsCfg, err := LoadServerTLSConfig("certs/server.crt", "certs/server.key", "certs/ca.crt")
	if err != nil {
		global.ErrorLog.Println("waitControlConn() 加载 TLS 配置失败: ", err)
		panic("加载 TLS 配置失败: " + err.Error())
	}

	listener, err := tls.Listen("tcp", ":"+global.Config.ListenPort, tlsCfg)
	if err != nil {
		global.ErrorLog.Println("waitControlConn() 监听失败: ", err)
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
				global.ErrorLog.Println("控制连接接入失败:", err)
				continue
			}
		}

		session, err := smux.Server(conn, nil)
		if err != nil {
			fmt.Println("创建会话失败:", err)
			global.ErrorLog.Println("创建会话失败:", err)
			_ = conn.Close()
			continue
		}

		go startHeartbeat(session, 5*time.Second, global.Ring)

		global.InfoLog.Println("会话建立成功")
		global.ConnPool.SetSession(session)
	}
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
			err = stream.SetReadDeadline(time.Now().Add(5 * time.Second))
			if err != nil {
				global.Log.Error("heartbeat(): SetReadDeadline失败:", err)
				return
			}
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
