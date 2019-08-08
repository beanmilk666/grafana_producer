package main

import (
	"bytes"
	"fmt"
	"github.com/c4pt0r/ini"
	"grafana_producer/hostmonitor/runtime_monitor"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime/debug"
	"time"
)

var monitorInfo MonitorInfo

//监视配置信息
type MonitorInfo struct {
	Url           string
	HostCpuMetric string
}

func main() {
	err := loadConf()
	if nil != err {
		fmt.Printf("读取配置文件失败，退出进程，err:%v\n", err)
		return
	}

	go monitor()

	select {}
}

func monitor() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("monitor err:", err)
			debug.PrintStack()
		}
	}()

	timeTick30 := time.Tick(time.Second * 30)

	for {
		select {
		case <-timeTick30:
			cpuRate := runtime_monitor.GetHostCpuRate()
			go sendMonitorDataFloat(cpuRate, monitorInfo.HostCpuMetric)
		}
	}
}

//发送浮点型监控数据
func sendMonitorDataFloat(value float64, metric string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("sendMonitorDataFloat err:", err)
			debug.PrintStack()
		}
	}()

	format :=
		`{
			"Time": "%s",
			"LongValue": 0,
			"MetaData": "%s",
			"Host": "%s",
			"DoubleValue": %.2f
		}`

	hostName, _ := os.Hostname()
	now := time.Now().UnixNano()
	timeStr := time.Unix(now/1e9, 0).Format("2006-01-02 15:04:05") + fmt.Sprintf(".%03d", now%1e9/1e6)

	msg := fmt.Sprintf(format, timeStr, metric, hostName, value)

	json := []byte(msg)
	resp, httpErr := http.Post(monitorInfo.Url, "application/json", bytes.NewBuffer(json))

	if httpErr != nil {
		fmt.Println("sendMonitorDataFloat err:", httpErr)
	} else if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
		_, _ = io.Copy(ioutil.Discard, resp.Body)
	}
}

func loadConf() error {
	conf := ini.NewConf("config.ini")

	url := conf.String("Basic", "url", "")
	hostCpuMetric := conf.String("Metric", "HostCpuMetric", "")

	if err := conf.Parse(); nil != err {
		return err
	}

	monitorInfo.Url = *url
	monitorInfo.HostCpuMetric = *hostCpuMetric

	return nil
}
