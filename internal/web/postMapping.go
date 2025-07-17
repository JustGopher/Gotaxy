package web

import (
	"encoding/json"
	"fmt"
	"github/JustGopher/Gotaxy/internal/global"
	"github/JustGopher/Gotaxy/internal/storage/models"
	"net/http"
	"strconv"
)

// Mapping 端口映射结构
type Mapping struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PublicPort int    `json:"publicPort"`
	TargetAddr string `json:"targetAddr"`
	Enable     bool   `json:"enable"`
	Status     string `json:"status"` // 连接状态，从连接池获取
	Traffic    int64  `json:"traffic"`
}

// mappingsHandler 获取所有映射列表
func mappingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从连接池获取映射列表
	poolMappings := global.ConnPool.All()

	// 转换为前端需要的格式
	result := make([]Mapping, 0, len(poolMappings))
	for i, m := range poolMappings {
		port, _ := strconv.Atoi(m.PublicPort)

		// 处理状态显示
		enable := true
		if m.Enable {
			enable = false
		}

		result = append(result, Mapping{
			ID:         i + 1, // 使用索引作为ID
			Name:       m.Name,
			PublicPort: port,
			TargetAddr: m.TargetAddr,
			Enable:     enable,
			Status:     m.Status,
		})
	}

	// 返回JSON格式的映射列表
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   result,
	})
	if err != nil {
		return
	}
}

// addMappingHandler 添加新的映射
func addMappingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求体
	var mapping Mapping
	err := json.NewDecoder(r.Body).Decode(&mapping)
	if err != nil {
		http.Error(w, fmt.Sprintf("解析请求失败: %v", err), http.StatusBadRequest)
		global.ErrorLog.Println("解析添加映射请求失败: ", err)
		return
	}

	// 验证参数
	if mapping.PublicPort <= 0 || mapping.PublicPort > 65535 {
		http.Error(w, "公共端口必须在1-65535之间", http.StatusBadRequest)
		return
	}

	if mapping.TargetAddr == "" {
		http.Error(w, "目标地址不能为空", http.StatusBadRequest)
		return
	}

	// 使用用户提供的名称，如果为空则生成一个
	name := mapping.Name
	if name == "" {
		name = fmt.Sprintf("map_%d", mapping.PublicPort)
	}

	// 处理Enable字段
	dbEnable := false
	if mapping.Enable == true {
		dbEnable = true
	}

	// 保存到数据库
	err = models.InsertMpg(global.DB, models.Mapping{
		Name:       name,
		PublicPort: strconv.Itoa(mapping.PublicPort),
		TargetAddr: mapping.TargetAddr,
		Enable:     dbEnable, // 使用处理后的Enable值
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("添加映射失败: %v", err), http.StatusInternalServerError)
		global.ErrorLog.Println("添加映射失败: ", err)
		return
	}

	// 添加到连接池
	global.ConnPool.Set(name, strconv.Itoa(mapping.PublicPort), mapping.TargetAddr, false, mapping.Traffic)

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "添加映射成功",
		"id":      name,
	})
	if err != nil {
		return
	}
}

// delMappingHandler 删除映射
func delMappingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取要删除的映射ID
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "缺少ID参数", http.StatusBadRequest)
		return
	}

	// 从连接池获取映射列表
	poolMappings := global.ConnPool.All()

	// 查找对应ID的映射
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID参数必须为整数", http.StatusBadRequest)
		return
	}

	if id <= 0 || id > len(poolMappings) {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	// 获取映射名称和端口
	name := poolMappings[id-1].Name
	//port := poolMappings[id-1].PublicPort

	// 从数据库删除映射
	err = models.DeleteMapByName(global.DB, name)
	if err != nil {
		http.Error(w, fmt.Sprintf("删除映射失败: %v", err), http.StatusInternalServerError)
		global.ErrorLog.Println("删除映射失败: ", err)
		return
	}

	// 从连接池删除映射
	global.ConnPool.Delete(name)

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "删除映射成功",
	})
	if err != nil {
		return
	}
}

// UpdateMapEna 启用/禁用映射
func UpdateMapEna(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求体
	var request struct {
		ID     int  `json:"id"`
		Enable bool `json:"enable"` // "running" 或 "stopped"
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, fmt.Sprintf("解析请求失败: %v", err), http.StatusBadRequest)
		global.ErrorLog.Println("解析启用/禁用映射请求失败: ", err)
		return
	}

	// 验证参数
	if request.ID <= 0 {
		http.Error(w, "ID必须大于0", http.StatusBadRequest)
		return
	}

	if request.Enable != true && request.Enable != false {
		http.Error(w, "Enable必须为'running'或'stopped'", http.StatusBadRequest)
		return
	}

	// 从连接池获取映射列表
	poolMappings := global.ConnPool.All()

	// 检查ID是否有效
	if request.ID <= 0 || request.ID > len(poolMappings) {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	// 获取映射名称
	name := poolMappings[request.ID-1].Name

	// 转换Enable值
	dbEnable := false
	poolEnable := false
	if request.Enable == true {
		dbEnable = true
		poolEnable = true
	}

	// 更新数据库中的Enable字段
	_, err = global.DB.Exec("UPDATE mapping SET enable = ? WHERE name = ?", dbEnable, name)
	if err != nil {
		http.Error(w, fmt.Sprintf("更新映射状态失败: %v", err), http.StatusInternalServerError)
		global.ErrorLog.Println("更新映射状态失败: ", err)
		return
	}

	// 更新连接池中的映射状态
	updated := global.ConnPool.UpdateEnable(name, poolEnable)
	if !updated {
		global.ErrorLog.Println("在连接池中未找到映射: ", name)
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "更新映射状态成功",
	})
	if err != nil {
		return
	}
}
