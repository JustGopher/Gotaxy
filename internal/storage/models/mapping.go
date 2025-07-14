package models

import (
	"database/sql"
	"fmt"
)

// Mapping 映射表结构
type Mapping struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PublicPort string `json:"public_port"`
	TargetAddr string `json:"target_addr"`
	Status     string `json:"status"`
}

// CreateMpgStructure 创建映射表结构
func CreateMpgStructure(db *sql.DB) error {
	sqlMap := `Create table if not exists mapping (
    id integer primary key autoincrement,
    name varchar(255) not null unique,
    public_port varchar(255) not null,
    target_addr varchar(255) not null,
    status varchar(255) not null
    );`

	_, err := db.Exec(sqlMap)
	if err != nil {
		return fmt.Errorf("创建映射表失败: %v", err)
	}
	return nil
}

// InsertMpg 插入映射数据
func InsertMpg(db *sql.DB, m Mapping) error {
	_, err := db.Exec("insert into mapping (name, public_port, target_addr, status) values (?,?,?,?)",
		m.Name, m.PublicPort, m.TargetAddr, m.Status)
	if err != nil {
		return fmt.Errorf("插入映射数据失败: %v", err)
	}
	return nil
}

// GetAllMpg 查询映射数据
func GetAllMpg(db *sql.DB) ([]Mapping, error) {
	rows, err := db.Query("select * from mapping")
	if err != nil {
		return nil, fmt.Errorf("GetAllMpg() 查询映射数据失败: %v", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var mappingSli []Mapping
	for rows.Next() {
		var m Mapping
		err := rows.Scan(&m.ID, &m.Name, &m.PublicPort, &m.TargetAddr, &m.Status)
		if err != nil {
			return nil, fmt.Errorf("GetAllMpg() 解析映射数据失败: %v", err)
		}
		mappingSli = append(mappingSli, m)
	}
	return mappingSli, nil
}

// DeleteMapByName 删除映射
func DeleteMapByName(db *sql.DB, name string) error {
	if name == "" {
		return fmt.Errorf("DeleteMapByName() 删除映射失败: 映射名称不能为空")
	}

	_, err := db.Exec("delete from mapping where name =?", name)
	return fmt.Errorf("DeleteMapByName() 删除映射失败: %v", err)
}

// UpdateMap 更新映射
func UpdateMap(db *sql.DB, name string, key string, value string) (*Mapping, error) {
	var m Mapping

	_, err := db.Exec("update mapping set public_port = ?, target_addr = ? where name = ?", key, value, name)
	if err != nil {
		return nil, fmt.Errorf("UpdateMap() 更新映射失败: %v", err)
	}

	err = db.QueryRow("select * from mapping where name =?", name).Scan(
		&m.ID, &m.Name, &m.PublicPort, &m.TargetAddr, &m.Status)
	if err != nil {
		return nil, fmt.Errorf("UpdateMap() 查询映射失败: %v", err)
	}

	return &m, nil
}
