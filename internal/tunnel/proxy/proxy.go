package proxy

import (
	"context"
	"fmt"
	"github/JustGopher/Gotaxy/internal/global"
	"github/JustGopher/Gotaxy/internal/pool"
	"io"
	"log"
	"net"
	"strings"

	"golang.org/x/time/rate"
)

// StartPublicListener 持续监听公网端口流量，建立 stream 连接
// status 只在此函数中更新
func StartPublicListener(ctx context.Context, mapping *pool.Mapping) {
	pubPort := mapping.PublicPort
	target := mapping.TargetAddr
	listener, err := net.Listen("tcp", ":"+pubPort)
	if err != nil {
		if strings.Contains(err.Error(), "address already in use") {
			fmt.Printf("端口 %s 已被占用\n", pubPort)
			return
		}
		fmt.Printf("监听端口 %s 失败: %v\n", pubPort, err)
		return
	}
	log.Printf("监听端口 %s 映射到客户端 %s\n", pubPort, target)

	mapping.Status = "active"
	defer func() {
		mapping.Status = "inactive"
	}()

	// 初始化限流器，每秒 rateLimit 字节
	rateLimit := int(mapping.RateLimit)
	limiter := rate.NewLimiter(rate.Limit(rateLimit), rateLimit) // 每秒允许 rateLimit 字节

	go func() {
		// 监听端口关闭
		select {
		case <-ctx.Done():
			global.InfoLog.Println("关闭端口监听 :", pubPort)
			fmt.Printf("关闭端口监听 :%s\n", pubPort)
			_ = listener.Close()
			return
		case <-mapping.Ctx.Done():
			global.InfoLog.Println("关闭端口监听 :", pubPort)
			fmt.Printf("关闭端口监听 :%s\n", pubPort)
			_ = listener.Close()
			return
		}
	}()

	for {
		if !mapping.Enable {
			_ = listener.Close()
			return
		}
		publicConn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case <-mapping.Ctx.Done():
				return
			default:
				global.ErrorLog.Println("listener.Accept() 连接失败:", err)
				fmt.Printf("连接失败: %v", err)
				continue
			}
		}
		// 正常 smux 流转发
		session := global.ConnPool.GetSession()
		if session == nil {
			fmt.Println("无有效客户端连接，关闭连接")
			_ = publicConn.Close()
			continue
		}

		// OpenStream() 的作用是：在当前控制连接上创建一个逻辑流（可以理解为一条虚拟 TCP 通道）
		// smux 会把多个这样的 stream 数据复用在一条实际连接上
		stream, err := session.OpenStream()
		if err != nil {
			global.ErrorLog.Println("session.OpenStream() smux stream创建失败: ", err)
			fmt.Printf("smux stream 创建失败: %v", err)
			_ = publicConn.Close()
			continue
		}

		// 通知客户端目标地址
		// 在流建立后，服务端会先向客户端写一行数据：告诉它，目标地址是哪个（如 127.0.0.1:8080）。
		// 客户端拿到这个地址后，就会在自己的本地去连接这个目标地址，然后形成一对 stream <-> 内网本地服务连接 的代理。
		_, err = stream.Write([]byte("DIRECT\n" + target + "\n"))
		if err != nil {
			global.ErrorLog.Println("写入目标地址失败:", err)
			_ = publicConn.Close()
			_ = stream.Close()
			continue
		}

		fmt.Printf("建立转发: 端口 %s <=> 客户端本地 %s\n", pubPort, target)
		// nolint:contextcheck
		go rateLimitedProxy(mapping.Ctx, publicConn, stream, limiter, nil)
		// nolint:contextcheck
		go rateLimitedProxy(mapping.Ctx, stream, publicConn, limiter, mapping)
	}
}

// rateLimitedProxy 使用 rate.Limiter 限制速率的代理
func rateLimitedProxy(ctx context.Context, dst, src net.Conn, limiter *rate.Limiter, mapping *pool.Mapping) {
	defer func(dst net.Conn) {
		_ = dst.Close()
	}(dst)
	defer func(src net.Conn) {
		_ = src.Close()
	}(src)

	buf := make([]byte, 1024*512) // 缓存大小，512kb

	for {
		// 检查上下文是否已取消,如果未取消
		select {
		case <-ctx.Done():
			return
		case <-global.Ctx.Done():
			return
		default:
		}
		n, err := src.Read(buf)
		if err != nil {
			if err != io.EOF {
				global.ErrorLog.Println("数据读取失败:", err)
			}
			break
		}

		// 等待令牌（控制速率）
		err = limiter.WaitN(ctx, n) // 等待令牌，限制每次读取的字节数
		if err != nil {
			global.ErrorLog.Println("限流失败:", err)
			break
		}

		// 向目标写入数据
		_, err = dst.Write(buf[:n])
		if err != nil {
			global.ErrorLog.Println("数据写入失败:", err)
			break
		}

		// 更新流量统计
		if mapping != nil {
			mapping.Traffic += int64(n)
			global.InfoLog.Printf("流量转发: %d 字节", n)
		}
	}
}

// proxy 数据转发
// nolint:unused
func proxy(dst, src net.Conn, mapping *pool.Mapping) {
	defer func(dst net.Conn) {
		_ = dst.Close()
	}(dst)
	defer func(src net.Conn) {
		_ = src.Close()
	}(src)

	if mapping == nil {
		_, _ = io.Copy(dst, src)
		return
	}

	// 正常流量转发
	byteCount, _ := io.Copy(dst, src)
	mapping.Traffic += byteCount
	global.Config.TotalTraffic += byteCount
	global.InfoLog.Printf("流量转发: %d 字节", byteCount)
}
