package config

import (
	"database/sql"
	"fmt"
	"github/JustGopher/Gotaxy/internal/pool"
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
func (cfg *Config) ConfigLoad(db *sql.DB, pool *pool.Pool) {
	// todo 从数据库加载配置到 config 和连接池
	cfgMap, err := models.GetAllCfg(db)
	if err != nil {
		log.Printf("ConfigLoad() 查询配置数据失败 -> %v", err)
		return
	}
	if len(cfgMap) != 3 {
		err = models.InsertCfg(db, "server_ip", "127.0.0.1")
		if err != nil {
			log.Printf("ConfigLoad() 创建配置数据失败 -> %v", err)
		}

		err = models.InsertCfg(db, "listen_port", "9000")
		if err != nil {
			log.Printf("ConfigLoad() 创建配置数据失败 -> %v", err)
		}

		err = models.InsertCfg(db, "email", "")
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

	mpg, err := models.GetAllMpg(db)
	if err != nil {
		return
	}

	for _, v := range mpg {
		pool.Set(v.Name, v.PublicPort, v.TargetAddr)
	}
}
