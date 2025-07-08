package serverCore

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

const (
	token        = "your_fixed_token" // 替换为你的固定 token
	externalPort = "9080"             // 服务端暴露的外部端口
)

// StartServer 启动监听
func StartServer() {
	// 服务端监听客户端连接的端口
	clientListener, err := net.Listen("tcp", ":9000")
	if err != nil {
		fmt.Println("服务端监听客户端连接失败:", err)
		os.Exit(1)
	}
	defer func(clientListener net.Listener) {
		err := clientListener.Close()
		if err != nil {
			log.Println("关闭客户端连接失败:", err)
			return
		}
	}(clientListener)
	fmt.Println("服务端监听客户端连接端口 9000...")

	// 服务端暴露给外部的端口
	externalListener, err := net.Listen("tcp", ":"+externalPort)
	if err != nil {
		fmt.Println("服务端监听外部端口失败:", err)
		os.Exit(1)
	}
	defer func(externalListener net.Listener) {
		err := externalListener.Close()
		if err != nil {
			log.Println("关闭外部连接失败:", err)
			return
		}
	}(externalListener)
	fmt.Println("服务端暴露外部端口", externalPort, "...")

	for {
		clientConn, err := clientListener.Accept()
		if err != nil {
			fmt.Println("接受客户端连接失败:", err)
			continue
		}
		fmt.Println("接受客户端连接:", clientConn.RemoteAddr())

		go handleClient(clientConn, externalListener)
	}
}
func handleClient(clientConn net.Conn, externalListener net.Listener) {
	defer func(clientConn net.Conn) {
		err := clientConn.Close()
		if err != nil {
			log.Println("关闭客户端连接失败:", err)
			return
		}
	}(clientConn)

	// 读取客户端发送的 token
	reader := bufio.NewReader(clientConn)
	tokenStr, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("读取客户端 token 失败:", err)
		return
	}

	// 去掉换行符
	tokenStr = strings.TrimSpace(tokenStr)

	// 验证 token
	if tokenStr != token {
		fmt.Println("客户端 token 验证失败:", tokenStr)
		return
	}

	fmt.Println("客户端 token 验证成功:", tokenStr)

	// 接受外部连接
	for {
		externalConn, err := externalListener.Accept()
		if err != nil {
			fmt.Println("接受外部连接失败:", err)
			continue
		}
		fmt.Println("接受外部连接:", externalConn.RemoteAddr())

		go forwardData(externalConn, clientConn)
		go forwardData(clientConn, externalConn)
	}
}

// 转发数据
func forwardData(src net.Conn, dst net.Conn) {
	defer func(src net.Conn) {
		err := src.Close()
		if err != nil {
			log.Println("关闭连接失败:", err)
			return
		}
	}(src)
	defer func(dst net.Conn) {
		err := dst.Close()
		if err != nil {
			log.Println("关闭连接失败:", err)
			return
		}
	}(dst)

	i, err := io.Copy(dst, src)
	if err != nil {
		return
	}
	fmt.Println("数据转发完成:", src.RemoteAddr(), "->", dst.RemoteAddr(), " 传输字节数: ", i, " bytes")
}
