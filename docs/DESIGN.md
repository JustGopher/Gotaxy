# 架构

目录结构:

```
gotaxy/
├── .chglog
├── .github/workflows
├── docs/ 
│   ├── images/              // 文档图片
│   ├── CHANGELOG.md 	     // 版本日志
│   ├── CICD.md     	     // 版本日志
│   ├── COMMIT_CONVENTION.md // 提交规范
│   ├── CONTRIBUTING.md      // 贡献指南
│   ├── DESIGN.md            // 版本日志
│   ├── DEVLOG.md            // 版本日志
│   └── DEVLOG.md            // 开发日志
├── cmd/                     // 主程序入口（client/server）
│   ├── server/              // 服务端启动入口
│   │   └── server.go
│   └── client/              // 客户端启动入口
│       └── client.go
├── internal/                // 核心逻辑代码
│   ├── config/              // 配置相关逻辑（读取 SQLite / JSON）
│   │   └── config.go
│   ├── tunnel/              // 穿透核心模块（TCP转发、连接管理）
│   │   ├── serverCore/      // 服务端核心逻辑
│   │   ├── clientCore/      // 客户端核心逻辑
│   │   └── pool/            // 连接池，管理所有客户端连接
│   ├── heartbeat/           // 心跳检测模块
│   │   └── heartbeat.go
│   ├── metrics/             // 监控模块
│   ├── notify/              // 通知模块
│   ├── web/                 // Web 控制台（Gin框架）
│   │   ├── api/v0.1/        // 业务接口
│   │   ├── service/         // 业务逻辑
│   │   ├── router.go
│   │   └── static/          // 前端资源（HTML/JS/CSS）
│   ├── storage/             // SQLite 操作模块
│   │   ├── db/              // 数据库初始化和表操作
│   │   └── models/
│   ├── test/                // 测试
│   └── logx/                // 日志模块（zap封装）
│       └── logger.go
├── pkg/                     // 公共库
│   ├── crypto/              // 加密模块
│   ├── utils/               // 纯函数工具类
│   ├── token/               // token相关
│   └── cli/                 // 命令行相关
├── deploy/                  // 部署脚本
│   └── docker-compose.yml
├── .gitignore
├── golangci.yml
├── go.mod
├── LICENSE
├── Makefile
└── README.md
```
