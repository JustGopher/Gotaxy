package pool

import (
	"fmt"
	"sync"
)

/**
  设计模型：配置持久化 + 状态缓存
*/

// Mapping 单个映射关系
type Mapping struct {
	Name       string // 规则名称，例如 "rule1"
	PublicPort string // 公网监听端口，例如 "9080"
	TargetAddr string // 映射目标地址，例如 "127.0.0.1:8080"
	Status     string // 连接状态，例如 "active", "inactive"
	Enable     string
}

// Pool 映射关系池
type Pool struct {
	mutex            sync.RWMutex
	table            map[string]*Mapping
	totalConnections int64
}

// NewPool 初始化连接池
func NewPool() *Pool {
	return &Pool{
		table: make(map[string]*Mapping),
	}
}

// Set 添加新映射关系
func (p *Pool) Set(name, port, target string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.totalConnections++
	key := fmt.Sprintf("%s%v", "rule", p.totalConnections+1)
	p.table[key] = &Mapping{
		Name:       name,
		PublicPort: port,
		TargetAddr: target,
		Status:     "inactive",
	}
}

// GetAllPort 获取所有映射关系的公网监听端口
func (p *Pool) GetAllPort() map[string]string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	m := make(map[string]string)
	for _, mapping := range p.table {
		m[mapping.PublicPort] = mapping.TargetAddr
	}
	return m
}

// Delete 删除指定规则映射关系
func (p *Pool) Delete(port string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	delete(p.table, port)
	p.totalConnections--
}

// All 获取所有映射关系
func (p *Pool) All() []*Mapping {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	list := make([]*Mapping, 0, len(p.table))
	for _, m := range p.table {
		list = append(list, m)
	}
	return list
}
