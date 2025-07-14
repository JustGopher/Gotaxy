package web

import (
	"embed"
	"fmt"
	"github/JustGopher/Gotaxy/internal/global"
	"html/template"
	"net/http"
)

//go:embed templates/*
var tmplFS embed.FS
var tmpl *template.Template

// Start 启动 Web 服务
func Start() {
	tmpl = template.Must(template.ParseFS(tmplFS, "templates/*.html"))
	http.HandleFunc("/", index)

	// 初始化路由
	InitRouter()

	fmt.Println("控制面板启动：localhost:9001")
	err := http.ListenAndServe(":9001", nil)
	if err != nil {
		fmt.Println("Web() 启动失败: ", err)
		global.Log.Error("Web() 启动失败: ", err)
		return
	}
}

// 处理主页请求
func index(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		global.Log.Error("index() 模板渲染失败: ", err)
		return
	}
}
