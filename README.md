# Gotaxy

<img align="right" width="280px"  src="docs/images/logo.png"  alt="logo"> 

🚀 Gotaxy 是一款基于 Go 语言开发的轻量级内网穿透工具，帮助开发者将内网服务安全、便捷地暴露到公网。

> _"Go beyond NAT, with style."_

## ✨ TODO

- 🧩 支持端口转发，将内网服务暴露到公网
- 🔒 加密传输，保护数据安全
- 🧰 简洁 CLI 工具，易于部署
- 📦 支持多平台，Windows、Linux、MacOS 等
- 🔗 支持自定义域名和自定义端口
- 📊 支持流量统计，方便了解使用情况
- 🌐 服务端通过 Web 面板控制，管理方便
- 🔄 后续将支持多客户端、 TCP/UDP代理等功能

---

## 🚀 快速开始

### 服务端启动

```bash
go run cmd/server/main.go
```

### 客户端连接

```bash
go run cmd/client/main.go
```
---

## 提交贡献

欢迎提交 Issue 和 Pull Request。

如果要贡献代码，请查阅 [CONTRIBUTING.md.md](docs/CONTRIBUTING.md) 文件

提交代码请阅读 [COMMIT_CONVENTION.md](docs/COMMIT_CONVENTION.md)，我们遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范
