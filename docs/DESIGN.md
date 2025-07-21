# 架构

目录结构:

```
gotaxy/
├── .chglog
├── .github/workflows
├── certs/                   // 证书
├── cmd/                     // 主程序入口（client/server）
│   ├── server/              // 服务端启动入口
│   └── client/              // 客户端启动入口
├── data/                    // 数据文件
├── docs/                    // 文档
│   ├── images/              // 文档图片
│   ├── CHANGELOG.md 	     // 版本日志
│   ├── CICD.md     	     // CICD方案
│   ├── COMMIT_CONVENTION.md // 提交规范
│   ├── CONTRIBUTING.md      // 贡献指南
│   ├── DESIGN.md            // 架构设计
│   ├── DEVLOG.md            // 开发日志
│   ├── REQUIREMENTS.md      // 需求文档
│   └── TODO.md              // 任务列表
├── internal/                // 核心逻辑代码
│   ├── config/              // 配置相关逻辑
│   ├── global/              // 全局变量
│   ├── heart/               // 心跳检测模块
│   ├── inits/               // 初始化模块
│   ├── pool/                // 连接池
│   ├── shell/               // 交互模块
│   ├── storage/             // SQLite 存储模块
│   ├── tunnel/              // 核心转发模块
│   └── web/                 // Web 控制面板
├── logs/                    // 日志文件
├── pkg/                     // 公共库
│   ├── email/               // 邮件发送模块
│   ├── logger/              // 日志模块
│   ├── tlsgen/              // 证书生成模块
│   └── utils/               // 工具函数模块
├── deploy/                  // 部署脚本（待完成）
│   └── docker-compose.yml
├── .gitignore
├── golangci.yml
├── go.mod
├── LICENSE
├── Makefile                // 构建脚本（待完成）
└── README.md
```
