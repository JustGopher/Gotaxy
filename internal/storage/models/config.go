package models

import (
	"database/sql"
	"log"
)

type Config struct {
	ID    int    `json:"id"`
	Key   string `json:"key"` // unique
	Value string `json:"value"`
}

// CreateCfgStructure 创建配置表结构
func CreateCfgStructure(db *sql.DB) error {
	sqlCon := `Create table if not exists config (
    id integer primary key,
    key varchar(255) not null unique,
    value varchar(255) not null
	);`

	_, err := db.Exec(sqlCon)
	if err != nil {
		log.Printf("创建配置表结构失败 -> %v", err)
		return err
	}
	return nil
}

// InsertCfg 创建配置数据
func InsertCfg(db *sql.DB, key string, value string) error {
	_, err := db.Exec("insert into config (key, value) values (?,?)", key, value)
	if err != nil {
		log.Printf("插入配置数据失败 -> %v", err)
		return err
	}
	return nil
}

// GetAllCfg 获取所有配置数据
func GetAllCfg(db *sql.DB) (map[string]string, error) {
	rows, err := db.Query("select key,value from config")
	if err != nil {
		log.Printf("查询配置数据失败 -> %v", err)
		return nil, err
	}

	defer rows.Close()

	configMap := make(map[string]string)

	for rows.Next() {
		var key, value string
		err := rows.Scan(&key, &value)
		if err != nil {
			log.Printf("扫描配置数据失败 -> %v", err)
			return nil, err
		}
		configMap[key] = value
	}
	return configMap, nil
}

// UpdateCfg 更新配置数据
func UpdateCfg(db *sql.DB, key string, value string) error {
	_, err := db.Exec("update config set value = ? where key = ?", value, key)
	if err != nil {
		log.Printf("更新配置数据失败 -> %v", err)
		return err
	}
	return nil
}
