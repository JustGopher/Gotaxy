package pool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/xtaci/smux"
)

/**
  设计模型：配置持久化 + 状态缓存
*/

// Mapping 单个映射关系
type Mapping struct {
	Name       string             // 规则名称，例如 "rule1"
	PublicPort string             // 公网监听端口，例如 "9080"
	TargetAddr string             // 映射目标地址，例如 "127.0.0.1:8080"
	Ctx        context.Context    // 上下文，用于关闭连接
	CtxCancel  context.CancelFunc // 上下文取消函数，用于关闭连接
	Traffic    int64              // 流量统计
	Status     string             // 连接状态，例如 "active", "inactive"
	Enable     bool               // 是否启用，例如 true, false
}

// Pool 映射关系池
type Pool struct {
	mutex            sync.RWMutex
	table            map[string]*Mapping
	currentSession   atomic.Value
	totalConnections int64
}

// NewPool 初始化连接池
func NewPool() *Pool {
	return &Pool{
		table: make(map[string]*Mapping),
	}
}

// Set 添加新映射关系
func (p *Pool) Set(name, port, target string, enable bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.totalConnections++
	//key := fmt.Sprintf("%s%v", "rule", p.totalConnections+1)
	p.table[name] = &Mapping{
		Name:       name,
		PublicPort: port,
		TargetAddr: target,
		Status:     "inactive",
		Enable:     enable,
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
func (p *Pool) Delete(name string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	delete(p.table, name)
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

// UpdateEnable 更新映射的启用状态
func (p *Pool) UpdateEnable(name string, enable bool) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 查找对应名称的映射
	for _, mapping := range p.table {
		if mapping.Name == name {
			mapping.Enable = enable
			return true
		}
	}
	return false
}

// UpdateStatus 更新映射的连接状态
func (p *Pool) UpdateStatus(name string, status string) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 查找对应名称的映射
	for _, mapping := range p.table {
		if mapping.Name == name {
			mapping.Status = status
			return true
		}
	}
	return false
}

// Close 关闭连接
func (p *Pool) Close(name string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	// 煎炒name是否存在
	if _, ok := p.table[name]; !ok {
		return errors.New("规则不存在，请检查name是否正确")
	}
	// 关闭上下文
	p.table[name].CtxCancel()
	return nil
}

// GetSession 获取当前活跃的会话
func (p *Pool) GetSession() *smux.Session {
	session, ok := p.currentSession.Load().(*smux.Session)
	if !ok {
		return nil
	}
	return session
}

// SetSession 设置当前活跃的会话
func (p *Pool) SetSession(session *smux.Session) {
	p.currentSession.Store(session)
}

// GetMapping 获取指定名称的映射关系
func (p *Pool) GetMapping(name string) *Mapping {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.table[name]
}
