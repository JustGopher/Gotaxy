package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

const (
	token      = "fixed_token"    // 替换为你的固定 token
	serverAddr = "127.0.0.1:9000" // 服务端的地址和端口
)

func main() {
	// 启动测试服务
	go HelloServe("8080")

	// 客户端连接到服务端
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Println("连接到服务端失败:", err)
		os.Exit(1)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("关闭连接失败:", err)
			return
		}
	}(conn)
	log.Println("成功连接到服务端:", serverAddr)

	// 发送 token 验证身份
	_, err = conn.Write([]byte(token + "\n"))
	if err != nil {
		fmt.Println("发送 token 失败:", err)
		return
	}

	// 客户端连接到本地的 HTTP 服务
	localAddr := "127.0.0.1:8080" // 本地 HTTP 服务的地址和端口
	localConn, err := net.Dial("tcp", localAddr)
	if err != nil {
		log.Println("连接到本地 HTTP 服务失败:", err)
		return
	}
	defer func(localConn net.Conn) {
		err := localConn.Close()
		if err != nil {
			log.Println("关闭本地连接失败:", err)
			return
		}
	}(localConn)
	log.Println("成功连接到本地 HTTP 服务:", localAddr)

	// 转发数据
	go forwardData(localConn, conn)
	go forwardData(conn, localConn)

	// 防止主程序退出
	select {}
}

// forwardData 转发数据
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

// HelloServe 测试服务
func HelloServe(port string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		write, err := w.Write([]byte("Hello, World!"))
		if err != nil {
			return
		}
		fmt.Println(write)
	})
	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}
}
