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
)

// StartPublicListener 持续监听公网端口流量，建立 stream 连接
// status 只在此函数中更新
func StartPublicListener(ctx context.Context, mapping *pool.Mapping) {
	pubPort := mapping.PublicPort
	target := mapping.TargetAddr
	listener, err := net.Listen("tcp", ":"+pubPort)
	if err != nil {
		if strings.Contains(err.Error(), "address already in use") {
			log.Printf("端口 %s 已被占用", pubPort)
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

	go func() {
		// 监听端口关闭
		select {
		case <-ctx.Done():
			global.Log.Info("关闭端口监听 :", pubPort)
			fmt.Printf("关闭端口监听 :%s", pubPort)
			_ = listener.Close()
			return
		case <-mapping.Ctx.Done():
			global.Log.Info("关闭端口监听 :", pubPort)
			fmt.Printf("关闭端口监听 :%s", pubPort)
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
				global.Log.Error("listener.Accept() 连接失败:", err)
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
