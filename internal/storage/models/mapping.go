package models

import (
	"database/sql"
	"fmt"
	"github/JustGopher/Gotaxy/internal/tunnel/serverCore/global"
	"log"
	"strings"
)

type Mapping struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PublicPort string `json:"public_port"`
	TargetAddr string `json:"target_addr"`
	Status     string `json:"status"`
}

func CreateMapping() {
	sqlMap := `Create table if not exists mapping (
    id integer primary key autoincrement,
    name varchar(255) not null unique,
    public_port varchar(255) not null,
    target_addr varchar(255) not null,
    status varchar(255) not null
    );`

	_, err := global.DB.Exec(sqlMap)
	if err != nil {
		log.Printf("创建映射表失败 -> %v", err)
	}
}

func InsertMap(m Mapping) {
	_, err := global.DB.Exec("insert into mapping (name, public_port, target_addr, status) values (?,?,?,?)", m.Name, m.PublicPort, m.TargetAddr, m.Status)
	if err != nil {
		log.Printf("插入映射数据失败 -> %v", err)
	}
}

func GetMapByName(name string) (*Mapping, error) {
	if name == "" {
		return nil, fmt.Errorf("查询映射数据失败！名字不能为空！")
	}

	query := "select * from mapping where name =? limit 1"
	row := global.DB.QueryRow(query, name)

	var mapping Mapping

	err := row.Scan(&mapping.ID, &mapping.Name, &mapping.PublicPort, &mapping.TargetAddr, &mapping.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &mapping, nil
}

func DeleteMapByName(name string) error {
	if name == "" {
		return fmt.Errorf("删除映射数据失败！名字不能为空！")
	}

	_, err := global.DB.Exec("delete from mapping where name =?", name)
	return err
}

func UpdateMap(name string, updates map[string]string) (*Mapping, error) {
	var (
		keys   []string      //用于存储更新的keys
		values []interface{} // 用于存储values
		m      Mapping
	)

	if name == "" {
		return nil, fmt.Errorf("更新映射字段失败！名字不能为空！")
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("没有要更新的字段！")
	}

	for key, value := range updates {
		keys = append(keys, key+" = ?")
		values = append(values, value)
	}

	values = append(values, name)
	query := "update mapping set " + strings.Join(keys, ",") + " where name =?"

	_, err := global.DB.Exec(query, values...)
	if err != nil {
		return nil, fmt.Errorf("更新映射失败: %v", err)
	}

	if updates["name"] == name {
		updates["name"] = name
	}

	err = global.DB.QueryRow("select * from mapping where name =?", updates["name"]).Scan(
		&m.ID, &m.Name, &m.PublicPort, &m.TargetAddr, &m.Status)
	if err != nil {
		return nil, fmt.Errorf("查询映射失败: %v", err)
	}

	return &m, nil
}
