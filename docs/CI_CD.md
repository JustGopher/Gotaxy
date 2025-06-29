# CI/CD方案

- **代码托管**：
    - 使用 GitHub 管理源码，开发者通过 Pull Request (PR) 提交代码变更

- **持续集成 (CI)**：

    - 使用 GitHub Actions 在 PR 阶段自动触发工作流

        - 执行 `golangci-lint` 进行静态代码检查，确保代码规范

        - Go test 执行测试并检查覆盖率，设置覆盖率阈值 > 20%

    - 全部检查通过后才允许合并（使用 PR required checks 强制）

- **持续部署 (CD)**：
    - docker 启动服务端并挂载程序、配置文件
    - 合并 dev 分支代码后，Jenkins 自动拉取最新代码，打包替换服务器原程序并启动