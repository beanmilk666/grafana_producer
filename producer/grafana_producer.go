package main

import (
	"encoding/json"
	"fmt"
	"github.com/c4pt0r/ini"
	"grafana_producer/utils"
	"net/http"
	"strings"
	"time"
)

var passInfo pass
var mysqlClient *utils.MysqlClient

//连接信息
type pass struct {
	username string
	password string
	network  string
	server   string
	port     int
	database string
}

//监控信息
type monitorData struct {
	Time        string
	LongValue   int
	DoubleValue float32
	MetaData    string
	Host        string
}

//http response
type resp struct {
	Code int
	Msg  string
}

func main() {

	err := loadConf()
	if nil != err {
		fmt.Printf("读取配置文件失败，退出进程，err:%v\n", err)
		return
	}

	mysqlClient = utils.NewMysqlClient(passInfo.username, passInfo.password, passInfo.network,
		passInfo.server, passInfo.port, passInfo.database)
	mysqlClient.SetConnMaxLifeTime(time.Second * 100)
	mysqlClient.SetMaxConns(100)
	mysqlClient.SetMaxIdleConns(20)

	if err = mysqlClient.Init(); nil != err {
		fmt.Printf("Open mysql database failed,err:%v\n", err)
		return
	}

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
	var mData monitorData
	if err := json.NewDecoder(request.Body).Decode(&mData); nil != err {
		_ = request.Body.Close()
		fmt.Printf("Json decode failed,err:%v\n", err)
		return
	}

	insert(&mData)

	var response resp
	response.Code = 1
	response.Msg = "Success"

	if err := json.NewEncoder(writer).Encode(response); nil != err {
		fmt.Printf("Json encode failed,err:%v\n", err)
	}
}

func insert(mData *monitorData) {
	metaData := mData.MetaData
	rsSlice := strings.Split(metaData, ".")
	if len(rsSlice) != 2 {
		fmt.Printf("Metric is illegal, the metric is:%v\n", metaData)
		return
	}
	tableName := rsSlice[0]
	metric := rsSlice[1]

	sql := fmt.Sprintf("insert into %s(intvar,doublevar,metric,hostname,time) values(?,?,?,?,?)", tableName)
	_, err := mysqlClient.Db.Exec(sql, mData.LongValue, mData.DoubleValue, metric, mData.Host, mData.Time)
	if nil != err {
		fmt.Printf("Insert failed,err:%v", err)
		return
	}
}
