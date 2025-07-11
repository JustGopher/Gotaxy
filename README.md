# Gotaxy

<img align="right" width="280px"  src="docs/images/logo.png"  alt="logo"> 

🚀 Gotaxy 是一款基于 Go 语言开发的轻量级内网穿透工具，帮助开发者将内网服务安全、便捷地暴露到公网。

<div style="
  line-height: 2;               /* 2 倍行高 */
  background: rgba(86,86,86,0.2);          /* 背景色与 GitHub 黄色块一致 */
  border-left: .25em solid #814b4b; /* 左边竖条 */
  padding: .8em 6em .8em .5em;            /* 上下留白，左右留空 */
  display: inline;              /* 关键：让背景只包裹文字 */
  box-decoration-break: clone;  /* 多行时，每行都带圆角和竖条 */
  -webkit-box-decoration-break: clone;
  border-radius: 0 4px 4px 0;   /* 右侧圆角 */
">
  &nbsp;"Go beyond NAT, with style."&nbsp;
</div>

---

###  🚀 快速开始

#### 服务端启动

```bash
go run cmd/server/server.go
```

#### 客户端连接

```bash
go run cmd/client/client.go -h ip -p port
```
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
  <span style="display:inline-block;overflow:hidden;line-height:0;">
    <img
      src="https://contri.buzz/api/wall?repo=JustGopher/Gotaxy&onlyAvatars=true"
      alt="Contributors' Wall for JustGopher/Gotaxy"
      style="display:block; margin-bottom:-50px;"
    />
  </span>
</a>

<br />
<br />
<br />
