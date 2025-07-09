package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/xtaci/smux"
)

const serverAddr = "localhost:9000"

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

func main() {
	go HelloServe()
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("连接服务端失败: %v", err)
	}
	log.Println("已连接服务端")

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
