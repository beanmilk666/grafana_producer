package utils

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type MysqlClient struct {
	db              *sql.DB
	connMaxLifeTime time.Duration //最大连接周期，超过连接时间则关闭连接
	maxConns        int           //最大连接数
	maxIdleConns    int           //最大闲置连接数
	closed          bool
	username        string
	password        string
	network         string
	server          string
	port            int
	database        string
}

func NewMysqlClient(username string, password string, network string, server string, port int, dbName string) *MysqlClient {
	return &MysqlClient{
		username: username,
		password: password,
		network:  network,
		server:   server,
		port:     port,
		database: dbName,
	}
}

func (mc *MysqlClient) SetConnMaxLifeTime(time time.Duration) {
	mc.connMaxLifeTime = time
}

func (mc *MysqlClient) SetMaxConns(v int) {
	mc.maxConns = v
}

func (mc *MysqlClient) SetMaxIdleConns(v int) {
	mc.maxIdleConns = v
}

func (mc *MysqlClient) Init() error {
	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", mc.username, mc.password, mc.network, mc.server, mc.port, mc.database)
	db, err := sql.Open("mysql", dsn)
	if nil != err {
		return err
	}

	if 0 == mc.connMaxLifeTime {
		db.SetConnMaxLifetime(time.Second * 100)
	} else {
		db.SetConnMaxLifetime(mc.connMaxLifeTime)
	}

	if 0 == mc.maxConns {
		db.SetMaxOpenConns(100)
	} else {
		db.SetMaxOpenConns(mc.maxConns)
	}

	if 0 == mc.maxIdleConns {
		db.SetMaxIdleConns(20)
	} else {
		db.SetMaxIdleConns(mc.maxIdleConns)
	}

	mc.db = db
	return nil
}
