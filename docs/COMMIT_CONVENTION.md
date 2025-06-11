# Git 提交规范

一个规范的 Git 提交备注通常包含以下几个部分：类型、范围、主题和可选的正文与脚注。

**1. 类型（必填）**

类型用于描述提交的性质，常见的类型包括：
 - **feat**：表示新增功能（feature）
 - **fix**：用于修复漏洞（bug fix）
 - **docs**：仅涉及文档（documentation）的更改
 - **style**：仅涉及代码格式、缺少分号、空格等，不影响代码运行的更改
 - **refactor**：代码重构，既不添加功能也不修复漏洞
 - **test**：添加测试用例或更新现有测试
 - **chore**：构建过程或辅助工具的更改，例如脚本、配置文件等
 - **revert**：回滚到之前的提交
 - **perf**：性能优化
 - **ci**：持续集成（Continuous Integration）相关更改


**2. 范围（可选）**

范围是的，用于指定提交影响的范围，比如模块名称、文件名或功能区域。

**3. 主题（可选）**

主题是对提交内容的简短描述，应该简洁明了，尽量控制在一行内，并且使用祈使句形式。

**4. 正文（可选）**

如果提交的内容较为复杂，仅靠主题无法完全表达清楚，可以在正文部分进行详细说明。正文可以包含代码更改的详细描述、问题的背景、解决方案、相关链接等信息。正文的格式建议使用 Markdown 格式，方便阅读和排版。

**5. 脚注（可选）**

脚注部分可以包含一些额外的信息，如关闭的 Issue 编号、相关提交的哈希值等。

## 示例
```text
feat(client): 添加 TLS 连接支持 
```
```text
fix(server): 修复内存泄漏问题 #123
```
```text
docs: 更新快速入门指南
```
```text
feat: add user authentication feature

Implement user authentication using JWT tokens.
This change includes:
- User registration endpoint
- User login endpoint
- Middleware for verifying tokens
```


## 注意事项
 - 标题不超过 50 字符
 - 尽量使用英文句号结尾
 - 关联 issue 使用 # 符号