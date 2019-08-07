package main

import (
	"fmt"
	"github.com/c4pt0r/ini"
	"grafana_producer"
	"net/http"
)

var passInfo pass
var mysqlClient *grafana_producer.MysqlClient

//连接信息
type pass struct {
	username string
	password string
	network  string
	server   string
	port     int
	database string
}

func main() {

	err := loadConf()
	if nil != err {
		fmt.Printf("读取配置文件失败，退出进程，err:%v\n", err)
		return
	}

	//mysqlClient = grafana_producer.NewMysqlClient(passInfo.username, passInfo.password, passInfo.network,
	//	passInfo.server, passInfo.port, passInfo.database)
	//mysqlClient.SetConnMaxLifeTime(time.Second * 100)
	//mysqlClient.SetMaxConns(100)
	//mysqlClient.SetMaxIdleConns(20)

	//if err = mysqlClient.Init(); nil != err {
	//	fmt.Printf("Open mysql database failed,err:%v\n", err)
	//	return
	//}

	http.HandleFunc("/monitor", rsvMonitorData)
	if err := http.ListenAndServe("0.0.0.0:10000", nil); nil != err {
		fmt.Printf("Start HttpListener failed,err:%v\n", err)
	}

}

func loadConf() error {
	conf := ini.NewConf("config.ini")

	username := conf.String("Pass", "username", "")
	password := conf.String("Pass", "password", "")
	network := conf.String("Pass", "network", "")
	server := conf.String("Pass", "server", "")
	port := conf.Int("Pass", "port", 0)
	database := conf.String("Pass", "database", "")

	if err := conf.Parse(); nil != err {
		return err
	}

	passInfo.username = *username
	passInfo.password = *password
	passInfo.network = *network
	passInfo.server = *server
	passInfo.port = *port
	passInfo.database = *database

	return nil
}

func rsvMonitorData(writer http.ResponseWriter, request *http.Request) {
	fmt.Println(writer)
}
