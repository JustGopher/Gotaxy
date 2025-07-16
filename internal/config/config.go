package config

import (
	"database/sql"
	"fmt"

	"github/JustGopher/Gotaxy/internal/pool"
	"github/JustGopher/Gotaxy/internal/storage/models"
)

// Config 配置
type Config struct {
	Name       string `json:"name"`
	ServerIP   string `json:"server_ip"`
	ListenPort string `json:"listen_port"`
	Email      string `json:"email"`
}

// ConfigLoad 配置加载
func (cfg *Config) ConfigLoad(db *sql.DB, pool *pool.Pool) {
	cfgMap, err := models.GetAllCfg(db)
	if err != nil {
		fmt.Printf("ConfigLoad() 查询配置数据失败 -> %v", err)
		return
	}
	if len(cfgMap) != 3 {
		err = models.InsertCfg(db, "server_ip", "127.0.0.1")
		if err != nil {
			fmt.Printf("ConfigLoad() 创建配置数据失败 -> %v", err)
		}
		err = models.InsertCfg(db, "listen_port", "9000")
		if err != nil {
			fmt.Printf("ConfigLoad() 创建配置数据失败 -> %v", err)
		}
		err = models.InsertCfg(db, "email", "")
		if err != nil {
			fmt.Printf("ConfigLoad() 创建配置数据失败 -> %v", err)
		}
		cfgMap, err = models.GetAllCfg(db)
		if err != nil {
			fmt.Printf("ConfigLoad() 创建配置数据失败 -> %v", err)
			return
		}
	}
	cfg.ServerIP = cfgMap["server_ip"]
	cfg.ListenPort = cfgMap["listen_port"]
	cfg.Email = cfgMap["email"]
	fmt.Println("已加载配置...")
	mpg, err := models.GetAllMpg(db)
	if err != nil {
		fmt.Printf("ConfigLoad() 查询映射数据失败 -> %v", err)
		return
	}
	for _, v := range mpg {
		if v.Enable == "open" {
			pool.Set(v.Name, v.PublicPort, v.TargetAddr, true)
		} else {
			pool.Set(v.Name, v.PublicPort, v.TargetAddr, false)
		}
	}
}
