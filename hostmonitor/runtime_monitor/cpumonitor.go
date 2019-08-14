package runtime_monitor

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

const (
	LINUX        string = "linux"
	UTIME_INDEX         = 13 //任务在用户态运行时间的下标
	STIME_INDEX         = 14 //任务在核心态运行时间的下标
	CUTIME_INDEX        = 15 //任务的所有已死线程在用户态运行时间下标
	CSTIME_INDEX        = 16 //任务的所有已死线程在核心态运行时间下标
)

//根据pid获取单个进程程内存占用率
func GetProcessCpuRateByPid(pid string) float64 {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Get process cpu rate failure.")
			debug.PrintStack()
		}
	}()
	//当前操作系统类型
	osType := runtime.GOOS
	if LINUX != osType {
		fmt.Println("Only Linux is supported by this package temporarily.")
		return 0
	}
	pTime1 := getProcessCpuTime(pid)
	tTime1 := getTotalCpuTime()
	time.Sleep(500 * time.Millisecond)
	pTime2 := getProcessCpuTime(pid)
	tTime2 := getTotalCpuTime()
	ratio := (float64(pTime2) - float64(pTime1)) / (float64(tTime2) - float64(tTime1))
	return ratio * 100
}

//获取主机CPU占用率
func GetHostCpuRate() float64 {
	idle0, total0 := getHostCpuSample()
	time.Sleep(500 * time.Millisecond)
	idle1, total1 := getHostCpuSample()
	idleTicks := float64(idle1 - idle0)
	totalTicks := float64(total1 - total0)
	return 100 * (totalTicks - idleTicks) / totalTicks
}

//根据进程名称获取单个进程占用率
func GetProcessCpuRateByPName(pName string) float64 {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Get process cpu rate failure.")
			debug.PrintStack()
		}
	}()
	shell := "ps -ef | grep \"" + pName + "\" | grep -v grep | awk '{print $2}'"
	cmd := exec.Command("/bin/bash", "-c", shell)
	bytes, err := cmd.Output()
	if nil != err {
		fmt.Println(err)
		return 0
	}
	rs := string(bytes)
	for -1 != strings.Index(rs, "\n") {
		rs = strings.Replace(string(rs), "\n", ",", 1)
	}
	rsSlice := strings.Split(rs, ",") //最后位置有一个多余的换行,所以切片的实际长度要减一
	if len(rsSlice)-1 > 1 {
		fmt.Println("This process is not the only one, please confirm the process name again!")
		return 0
	} else if 0 == len(rsSlice)-1 {
		fmt.Println("This process is not existing, please confirm the process name again!")
		return 0
	}
	return GetProcessCpuRateByPid(rsSlice[0])
}

//根据进程名称获取单个进程占用率(排除其他干扰进程名称)
func GetProcessCpuRateByPNameInvert(pName string, invertKeyWord string) float64 {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Get process cpu rate failure.")
			debug.PrintStack()
		}
	}()
	shell := "ps -ef | grep \"" + pName + "\" | grep -v grep | grep -v " + invertKeyWord + "| awk '{print $2}'"
	cmd := exec.Command("/bin/bash", "-c", shell)
	bytes, err := cmd.Output()
	if nil != err {
		fmt.Println(err)
		return 0
	}
	rs := string(bytes)
	for -1 != strings.Index(rs, "\n") {
		rs = strings.Replace(string(rs), "\n", ",", 1)
	}
	rsSlice := strings.Split(rs, ",") //最后位置有一个多余的换行,所以切片的实际长度要减一
	if len(rsSlice)-1 > 1 {
		fmt.Println("This process is not the only one, please confirm the process name again!")
		return 0
	} else if 0 == len(rsSlice)-1 {
		fmt.Println("This process is not existing, please confirm the process name again!")
		return 0
	}
	return GetProcessCpuRateByPid(rsSlice[0])
}

//根据pid获取进程使用时间，单位为jiffies
//进程的总CPU时间:time=utime+stime+cutime+sutime（详细解释见常亮定义）
func getProcessCpuTime(pid string) int {
	fileName := fmt.Sprintf("/proc/%s/stat", pid) //进程的状态信息文件（包含进程CPU使用时间）
	file, err := os.Open(fileName)
	defer file.Close()
	if nil != err {
		fmt.Println("Open file err:", err)
		return 0
	}
	con, err := ioutil.ReadAll(file)
	if nil != err {
		fmt.Println("Read file err:", err)
		return 0
	}
	//[]byte直接转化成string会在最后加上一个换行
	cons := strings.Replace(string(con), "\n", "", 1)
	rsSlice := strings.Split(cons, " ")
	uTime, _ := strconv.Atoi(rsSlice[UTIME_INDEX])
	sTime, _ := strconv.Atoi(rsSlice[STIME_INDEX])
	cuTime, _ := strconv.Atoi(rsSlice[CUTIME_INDEX])
	csTime, _ := strconv.Atoi(rsSlice[CSTIME_INDEX])
	return uTime + sTime + cuTime + csTime
}

//获取CPU总使用时间，单位为jiffies
//[cpu  4852929 335 3980828 1018525134 359190 0 186774 0 0 0]总长度是12个（cpu后有一个空格）
func getTotalCpuTime() int {
	file, err := os.Open("/proc/stat")
	defer file.Close()
	if nil != err {
		fmt.Println("Open file err:", err)
		return 0
	}
	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n') //读取第一行，总CPU耗时信息在第一行
	line = strings.Replace(line, "\n", "", 1)
	if err != nil {
		if err == io.EOF {
			return 0
		}
		return 0
	}
	rsSlice := strings.Split(line, " ")
	totalTime := 0
	for _, sTime := range rsSlice[2:] {
		t, _ := strconv.Atoi(sTime)
		totalTime += t
	}
	return totalTime
}

//获取主机CPU使用样本
func getHostCpuSample() (idle, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val // tally up all the numbers to get total ticks
				if i == 4 {  // idle is the 5th field in the cpu line
					idle = val
				}
			}
			return
		}
	}
	return
}
