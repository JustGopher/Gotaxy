# Gotaxy

<img align="right" width="280px"  src="docs/images/logo2.png"  alt="logo"> 

[English](README.md) | 简体中文

✈️ Gotaxy 是一款基于 Go 语言开发的轻量级内网穿透工具，帮助开发者将内网服务安全、便捷地暴露到公网。


**_"Go beyond NAT, with style."_**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache-blue.svg)](LICENSE)
[![SQLite](https://img.shields.io/badge/SQLite-1.38-blue?logo=sqlite)](https://pkg.go.dev/modernc.org/sqlite#section-readme)
[![smux](https://img.shields.io/badge/xtaci%2Fsmux-1.5.34-brightgreen)](https://github.com/xtaci/smux)
[![readline](https://img.shields.io/badge/chzyer%2Freadline-1.5.1-orange)](https://github.com/chzyer/readline)
[![Stars](https://img.shields.io/github/stars/JustGopher/Gotaxy?style=social)](https://github.com/JustGopher/Gotaxy/stargazers)


### 核心技术
- **语言**: Go 1.24+
- **网络**: TCP/TLS 协议
- **数据库**: SQLite (modernc.org/sqlite)
- **多路复用**: xtaci/smux
- **交互界面**: chzyer/readline

---

##  🚀 快速开始

### 获取程序

在 Release 中下载最新的版本，提供可执行程序、压缩包、源码，支持AMD64下的Linux和Windows环境运行

### 服务端启动

```bash
./gotaxy-server # 执行程序，若为windows，程序名为 gotaxy-server.exe，下方客户端同理
# 如果为通过源码运行:
# go run cmd/server/server.go
```

##### 生成证书

Gotaxy通过原生库实现自签名 CA 证书：它通过颁发和签名证书，确保内网穿透过程中 “通信双方身份可信” 且 “数据传输加密”，是保障工具安全使用的核心机制。

服务端和客户端证书二者配合 CA 根证书，共同构建了 Gotaxy 从 “身份验证” 到 “数据加密” 的完整安全链路，确保内网穿透过程既安全又可靠。

服务端通过交互命令生成证书:
```bash
gen-ca    [year]  # 生成 CA 根证书
gen-certs [day]   # 服务端和客户端证书
Options:
  year int
        证书有效期，单位为年 (default 10)
  day int
        证书有效期，单位为天 (default 365)
```

设置服务端IP、监听端口，以及需要穿透的内网服务的地址
```bash
set--ip <ip>
set--port <port>
add-mapping <name> <public_port> <target_addr> # 添加映射端口
open-mapping <name> # 新增的映射默认关闭，需手动打开
```

启动服务：
```bash
start # 启动服务端核心服务，开始监听客户端
```

### 客户端连接

启动客户端并建立端口转发隧道，客户端启动需要服务端主机IP和监听端口，同时需要携带服务端生成的TLS证书
```bash
./gotaxy-client start  -h [host] -p <port> [-ca <ca-cert-path>] [-crt <client-cert-path>] [-key <private-key-path>]
# 如果通过源码运行:
# go run cmd/client/client.go -h [host] -p <port> [-ca <ca-cert-path>] [-crt <client-cert-path>] [-key <private-key-path>]
Options:
  -h [host]     
        The hostname or IP address of the server (default "127.0.0.1")
  -p <port>
        The port number to connect to (default 9000)
  -ca <ca-cert-path>
        Path to the CA certificate file (default "certs/ca.crt")
  -crt <client-cert-path>
        Path to the client certificate file (default "certs/client.crt")
  -key <private-key-path>
        Path to the client private key file (default "certs/client.key")`)
```


### ⚙️ 服务端交互命令使用说明

以下列出了服务端的所有可用命令及其效果：



- `gen-ca [time(year)] [-overwrite]`

  有效期: 可选参数，指定CA证书的有效期，默认为10年

  -overwrite: 可选参数，强制覆盖已存在的CA证书

  示例: gen-ca 5 -overwrite  (生成有效期为5年的CA证书并覆盖已有证书)


- `gen-certs [time(day)]`

  有效期: 可选参数，指定证书的有效期(天)，默认为365天

  示例: gen-certs 30  (生成有效期为30天的证书)


- `start`

  功能: 启动服务器，会检查证书是否存在


- `stop`

  功能: 停止运行中的服务器


- `show-config`

  功能: 显示当前服务器IP、监听端口和邮箱配置


- `show-mapping`

  功能: 显示所有配置的端口映射及其状态


- `set-ip <ip>`

  功能: 设置服务端IP地址

  示例: set-ip 192.168.1.100


- `set-port <port>`

  功能: 设置服务端监听端口，范围为1-65535

  示例: set-port 9000


- `set-email <email>`

  功能: 设置服务端邮箱地址，用于接收通知

  示例: set-email admin@example.com


- `add-mapping <name> <public_port> <target_addr>`

  功能: 添加一个新的端口映射配置

  示例: add-mapping web 8080 127.0.0.1:3000


- `del-mapping <name>`

  功能: 删除指定名称的端口映射

  示例: del-mapping web


- `upd-mapping <name> <public_port> <target_addr> <rate>`

  功能: 更新指定名称的端口映射配置

  示例: upd-mapping web 8080 127.0.0.1:3000 2,097,152(2MB)


- `open-mapping <name>`

  功能: 打开指定名称的端口映射

  示例: open-mapping web


- `close-mapping <name>`

  功能: 关闭指定名称的端口映射

  示例: close-mapping web


- `heart`

  功能: 查看当前链接状态


- `mode [vi|emacs]`

  功能: 设置命令行编辑模式

  示例: mode vi  (切换到vi模式)


- `help`

  功能: 显示此帮助信息


- `exit`

  功能: 停止服务并退出命令行界面`

---

### 需求文档

详细需求分析请参阅 [REQUIREMENTS.md](docs/REQUIREMENTS.md) 文件。

---

### 提交贡献

欢迎提交 Issue 和 Pull Request。

如果要贡献代码，请查阅 [CONTRIBUTING.md.md](docs/CONTRIBUTING.md) 文件

提交代码请阅读 [COMMIT_CONVENTION.md](docs/COMMIT_CONVENTION.md)，我们遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范

---

<h3 align="left">贡献墙</h3>


<a href="https://github.com/JustGopher/Gotaxy/graphs/contributors">

<img src="https://contri.buzz/api/wall?repo=JustGopher/Gotaxy&onlyAvatars=true" alt="Contributors' Wall for JustGopher/Gotaxy" />

</a>

<br />
<br />