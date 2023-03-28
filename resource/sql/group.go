package sql

import (
	"fmt"
	mysql "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	log "github.com/lfxnxf/emo-frame/logging"
	"github.com/lfxnxf/emo-frame/resource/sql/trace"
	"strconv"
	"sync/atomic"
	"time"
)

// Client继承了*gorm.DB的所有方法, 详细的使用方法请参考:
// http://gorm.io/docs/connecting_to_the_database.html
type Client struct {
	*gorm.DB
}

type Group struct {
	name    string
	master  *Master
	replica []*Slave
	next    uint64
	total   uint64
}

type Master struct {
	client      *Client
	noLogClient *Client
}

type Slave struct {
	client      *Client
	noLogClient *Client
}

func parseConnAddress(address string) (string, int, int, int, error) {
	u, err := mysql.ParseDSN(address)
	if err != nil {
		return address, -1, -1, 0, err
	}
	q := u.Params
	idleQ, activeQ, lifetimeQ := q["max_idle"], q["max_active"], q["max_lifetime_sec"]
	maxIdle, _ := strconv.Atoi(idleQ)
	if maxIdle == 0 {
		maxIdle = 15
	}
	maxActive, _ := strconv.Atoi(activeQ)
	lifetime, _ := strconv.Atoi(lifetimeQ)
	if lifetime == 0 {
		lifetime = 1800
	}
	delete(q, "max_idle")
	delete(q, "max_active")
	delete(q, "max_lifetime_sec")
	return u.FormatDSN(), maxIdle, maxActive, lifetime, nil
}

func openDB(name, address string, isMaster int, statLevel, format, logLevel string) (*Client, error) {
	addr, maxIdle, maxActive, lifetime, err := parseConnAddress(address)
	if err != nil {
		return nil, err
	}
	db, err := trace.Open(addr, name)
	if err != nil {
		return nil, fmt.Errorf("open mysql [%s] master %s error %s", name, address, err)
	}
	db = db.Debug()
	db = db.BlockGlobalUpdate(true)
	db.SetLogger(newGlobalLogger(statLevel, isMaster, parseDbName(address), format, logLevel))
	db.DB().SetMaxIdleConns(maxIdle)
	db.DB().SetMaxOpenConns(maxActive)
	db.DB().SetConnMaxLifetime(time.Duration(lifetime) * time.Second)

	return &Client{DB: db}, err
}

// NewGroup初始化一个Group， 一个Group包含一个master实例和零个或多个slave实例
func NewGroup(c GroupConfig) (*Group, error) {
	log.Infof("init sql group name [%s], master [%s], slave [%v]", c.Name, c.Master, c.Slaves)
	g := Group{
		name:    c.Name,
		master:  new(Master),
		replica: make([]*Slave, 0),
	}
	var err error

	// 有日志的
	g.master.client, err = openDB(c.Name, c.Master, 1, c.StatLevel, c.LogFormat, c.LogLevel)
	if err != nil {
		return nil, err
	}

	// 无日志的，给定时扫表的操作使用
	g.master.noLogClient, err = openDB(c.Name, c.Master, 1, c.StatLevel, c.LogFormat, "error")
	if err != nil {
		return nil, err
	}

	g.replica = make([]*Slave, 0, len(c.Slaves))
	g.total = 0
	for _, slave := range c.Slaves {
		// 有日志slave
		hasLogC, err := openDB(c.Name, slave, 0, c.StatLevel, c.LogFormat, c.LogLevel)
		if err != nil {
			return nil, err
		}

		// 无日志slave
		noLogC, err := openDB(c.Name, slave, 0, c.StatLevel, c.LogFormat, "error")
		if err != nil {
			return nil, err
		}
		slave := &Slave{
			client:      hasLogC,
			noLogClient: noLogC,
		}
		g.replica = append(g.replica, slave)
		g.total++

	}
	return &g, nil
}

// Master返回master实例
func (g *Group) Master() *Client {
	return g.master.client
}

func (g *Group) MasterNoLog() *Client {
	return g.master.noLogClient
}

// Slave返回一个slave实例，使用轮转算法
func (g *Group) Slave() *Client {
	if g.total == 0 {
		return g.master.client
	}
	next := atomic.AddUint64(&g.next, 1)
	return g.replica[next%g.total].client
}

// Slave返回一个slave实例，使用轮转算法
func (g *Group) SlaveNoLog() *Client {
	if g.total == 0 {
		return g.master.noLogClient
	}
	next := atomic.AddUint64(&g.next, 1)
	return g.replica[next%g.total].noLogClient
}

// Instance函数如果isMaster是true， 返回master实例，否则返回slave实例
func (g *Group) Instance(isMaster bool) *Client {
	if isMaster {
		return g.Master()
	}
	return g.Slave()
}

func parseDbName(s string) string {
	u, err := mysql.ParseDSN(s)
	if err != nil {
		return s
	}
	return u.DBName
}
