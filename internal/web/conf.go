package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"github/JustGopher/Gotaxy/internal/global"
	"github/JustGopher/Gotaxy/internal/storage/models"
	"github/JustGopher/Gotaxy/pkg/utils"
	"net/http"
	"strconv"
)

// GetConf 获取配置
func GetConf(w http.ResponseWriter, r *http.Request) {
	cfg, err := models.GetAllCfg(global.DB)
	if err != nil {
		err := fmt.Errorf("配置加载失败，err：%w", err)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	data, _ := json.Marshal(cfg)
	_, _ = w.Write(data)
}

// UpdateConf 更新配置
func UpdateConf(w http.ResponseWriter, r *http.Request) {
	conf := global.Config
	db := global.DB
	err := json.NewDecoder(r.Body).Decode(&conf)
	if err != nil {
		err := fmt.Errorf("反序列化失败，err：%w", err)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	configFields := map[string]string{
		"server_ip":   conf.ServerIP,
		"listen_port": conf.ListenPort,
		"email":       conf.Email,
		// 新增字段可直接在此处添加
	}

	for key, value := range configFields {
		if value == "" {
			continue
		}
		// 验证字段
		if err := validateConfigField(key, value); err != nil {
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		// 更新字段
		if err := models.UpdateCfg(db, key, value); err != nil {
			err := fmt.Errorf("更新配置失败，err：%w", err)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}
	_, _ = w.Write([]byte("配置更新成功"))
}

// validateConfigField 验证字段
func validateConfigField(name, value string) error {
	switch name {
	case "server_ip":
		if value != "" && !utils.IsValidateIP(value) {
			return errors.New("IP地址格式不正确")
		}
	case "email":
		if value != "" && !utils.IsValidateEmail(value) {
			return errors.New("email格式不正确")
		}
	case "listen_port":
		if port, err := strconv.Atoi(value); err != nil || port < 1 || port > 65535 {
			return errors.New("端口号必须为1-65535之间的数字")
		}
		// 新增字段验证规则可在此处扩展
	}
	return nil
}
