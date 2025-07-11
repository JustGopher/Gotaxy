package models

import (
	"database/sql"
	"fmt"
	"github/JustGopher/Gotaxy/internal/global"
	"log"
	"strings"
)

type Config struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ServerIP   string `json:"server_ip"`
	ListenPost string `json:"Listen_post"`
}

func CreateConfig() {
	sqlCon := `Create table if not exists config (
    id integer primary key autoincrement,
    name varchar(255) not null unique,
    server_ip varchar(255) not null,
    listen_post varchar(255) not null
	);`

	_, err := global.DB.Exec(sqlCon)
	if err != nil {
		log.Printf("创建配置表失败 -> %v", err)
	}
}

func InsertCon(con Config) {
	_, err := global.DB.Exec("insert into config (name, server_ip, listen_post) values (?, ?, ?)", con.Name, con.ServerIP, con.ListenPost)
	if err != nil {
		log.Printf("插入配置数据失败 -> %v", err)
	}
}

func GetConByName(name string) (*Config, error) {

	if name == "" {
		return nil, fmt.Errorf("查询配置数据失败！名字不能为空！")
	}

	query := "select * from config where name = ? limit 1"
	row := global.DB.QueryRow(query, name)

	var config Config

	err := row.Scan(&config.ID, &config.Name, &config.ServerIP, &config.ListenPost)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &config, nil
}

func DeleteConByName(name string) error {
	if name == "" {
		return fmt.Errorf("删除配置数据失败！名字不能为空！")
	}

	_, err := global.DB.Exec("delete from config where name =?", name)
	return err
}

func UpdateCon(name string, updates map[string]string) (*Config, error) {
	var (
		keys   []string      //用于存储更新的keys
		values []interface{} // 用于存储values
		config Config
	)

	if name == "" {
		return nil, fmt.Errorf("更新配置字段失败！名字不能为空！")
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("没有要更新的字段！")
	}

	for field, value := range updates {
		keys = append(keys, field+" = ?")
		values = append(values, value)
	}

	values = append(values, name)
	query := "update config set" + " " + strings.Join(keys, ",") + " where name =?"

	_, err := global.DB.Exec(query, values...)
	if err != nil {
		return nil, fmt.Errorf("更新配置失败: %v", err)
	}

	err = global.DB.QueryRow("select * from config where name =?", updates["name"]).Scan(
		&config.ID, &config.Name, &config.ServerIP, &config.ListenPost)
	if err != nil {
		return nil, fmt.Errorf("查询更新后的配置失败: %v", err)
	}
	return &config, nil
}
