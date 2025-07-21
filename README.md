# Gotaxy

<img align="right" width="280px"  src="docs/images/logo2.png"  alt="logo"> 

🚀 Gotaxy 是一款基于 Go 语言开发的轻量级内网穿透工具，帮助开发者将内网服务安全、便捷地暴露到公网。


#### _"Go beyond NAT, with style."_

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Contributors](https://img.shields.io/github/contributors/JustGopher/Gotaxy)](https://github.com/JustGopher/Gotaxy/graphs/contributors)
[![Stars](https://img.shields.io/github/stars/JustGopher/Gotaxy?style=social)](https://github.com/JustGopher/Gotaxy/stargazers)

### 核心技术
- **语言**: Go 1.24+
- **网络**: TCP/TLS 协议
- **数据库**: SQLite (modernc.org/sqlite)
- **多路复用**: xtaci/smux
- **交互界面**: chzyer/readline

---

#  🚀 快速开始

## 运行项目

### 拷贝代码：

#### 服务端启动

```bash
go run cmd/server/server.go
```

##### 下载证书

CA 根证书在 Gotaxy 中的作用类似于 “身份证颁发机构”：它通过颁发和签名证书，确保内网穿透过程中 “通信双方身份可信” 且 “数据传输加密”，是保障工具安全使用的核心机制。


服务端和客户端证书二者配合 CA 根证书，共同构建了 Gotaxy 从 “身份验证” 到 “数据加密” 的完整安全链路，确保内网穿透过程既安全又可靠。
- 在命令行中:
```bash
# 生成 CA 根证书
gen-ca [year]
# 服务端和客户端证书
gen-certs [day]
Options:
  -year int
        证书有效期，单位为年 (default 10)
  -day int
        证书有效期，单位为天 (default 365)
```

- 在命令行中：
```bash
# 启动服务端
start
```

#### 客户端连接

```bash
go run cmd/client/client.go -h [host] -p <port> [-ca <ca-cert-path>] [-crt <client-cert-path>] [-key <private-key-path>]
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

### 下载打包文件：

- 点击项目页面的 "Actions" 选项卡
- 找到并下载最新的 "Release" 版本

#### 服务端启动：

```bash
# 生成 CA 根证书
./gotaxy-server gen-ca

# 使用 CA 签发服务端证书
./gotaxy-server gen-certs

# 启动服务端
./gotaxy-server start
```

#### 客户端连接：

```bash
# 启动客户端并建立端口转发隧道
./gotaxy-client start  -h [host] -p <port> [-ca <ca-cert-path>] [-crt <client-cert-path>] [-key <private-key-path>]
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

## ⚙️ 命令使用说明

以下列出了服务端的所有可用命令及其效果：

### 服务端命令（运行项目之后）

- gen-ca - 生成CA证书

  格式: gen-ca [有效期(年)] [-overwrite]

  有效期: 可选参数，指定CA证书的有效期，默认为10年
  
  -overwrite: 可选参数，强制覆盖已存在的CA证书

  示例: gen-ca 5 -overwrite  (生成有效期为5年的CA证书并覆盖已有证书)


- gen-certs - 生成服务端和客户端证书

  格式: gen-certs [有效期(日)]
  
  有效期: 可选参数，指定证书的有效期(天)，默认为10天
  
  示例: gen-certs 30  (生成有效期为30天的证书)
  

- start - 启动内网穿透服务器

  功能: 启动服务器，会检查证书是否存在


- stop - 停止内网穿透服务器

  功能: 停止运行中的服务器


- show-config - 显示服务端配置

  功能: 显示当前服务器IP、监听端口和邮箱配置


- show-mapping - 显示所有端口映射

  功能: 显示所有配置的端口映射及其状态


- set-ip - 设置服务端IP地址

  格式: set-ip <ip>

  功能: 设置服务端IP地址

  示例: set-ip 192.168.1.100


- set-port - 设置服务端监听端口

  格式: set-port <port>

  功能: 设置服务端监听端口，范围为1-65535

  示例: set-port 9000


- set-email - 设置服务端邮箱

  格式: set-email <email>

  功能: 设置服务端邮箱地址，用于接收通知

  示例: set-email admin@example.com


- add-mapping - 添加端口映射

  格式: add-mapping <名称> <公网端口> <目标地址> <状态>

  功能: 添加一个新的端口映射配置

  示例: add-mapping web 8080 127.0.0.1:3000


- del-mapping - 删除端口映射

  格式: del-mapping <名称>

  功能: 删除指定名称的端口映射

  示例: del-mapping web


- upd-mapping - 更新端口映射

  格式: upd-mapping <名称> <公网端口> <目标地址> <状态>

  功能: 更新指定名称的端口映射配置

  示例: upd-mapping web 8080 127.0.0.1:3000 open


- mode - 切换编辑模式

  格式: mode [vi|emacs]

  功能: 设置命令行编辑模式

  示例: mode vi  (切换到vi模式)


- help - 显示此帮助信息


- exit - 退出程序

  功能: 停止服务并退出命令行界面`

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