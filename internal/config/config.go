package config

import (
	"database/sql"
	"fmt"
	"github/JustGopher/Gotaxy/internal/storage/models"
	"log"
)

type Config struct {
	Name       string `json:"name"`
	ServerIP   string `json:"server_ip"`
	ListenPort string `json:"listen_port"`
	Email      string `json:"email"`
}

// ConfigLoad 配置加载
func (cfg *Config) ConfigLoad(db *sql.DB) {
	// todo 从数据库加载配置到 config 和连接池
	cfgMap, err := models.GetAllCfg(db)
	if err != nil {
		log.Printf("ConfigLoad() 查询配置数据失败 -> %v", err)
		return
	}
	if len(cfgMap) != 3 {
		err = models.CreateCfg(db, "server_ip", "127.0.0.1")
		if err != nil {
			log.Printf("ConfigLoad() 创建配置数据失败 -> %v", err)
		}

		err = models.CreateCfg(db, "listen_post", "9000")
		if err != nil {
		}

		err = models.CreateCfg(db, "email", "")
		if err != nil {
			log.Printf("ConfigLoad() 创建配置数据失败 -> %v", err)
		}
		cfgMap, err = models.GetAllCfg(db)
		if err != nil {
			log.Printf("ConfigLoad() 查询配置数据失败 -> %v", err)
			return
		}
	}
	fmt.Println(cfgMap)
	cfg.ServerIP = cfgMap["server_ip"]
	cfg.ListenPort = cfgMap["listen_port"]
	cfg.Email = cfgMap["email"]
}
