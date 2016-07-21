package assemble

import (
	"fmt"
	"os"
	"statUpload/zkconfig"
	"syscall"
	"time"
)

func writeRawLog() {

	timeLayOut := "2006-01-02 15:04"
	//获取上一分钟
	tPreMin := time.Now().Unix() - 60
	tPreMinStr := time.Unix(tPreMin, 0).Format(timeLayOut)

	var logStr string
	for logCmd, logMap := range eachRecord {
		var aTempStr string
		for logRetcode, logRetCont := range logMap {
			//格式:时间 | 命令字 | 返回码 | 出现次数 | 访问量 | 平均延时 | 最大延时 | 大于10ms次数 | 大于100ms次数 | 大于500ms次数
			aTempStr = fmt.Sprintf("%s | %s | %d | %d | %d | %.3f | %.3f | %d | %d | %d\n",
				tPreMinStr,
				logCmd,
				logRetcode,
				int64(logRetCont.accNum),
				int64(logRetCont.total),
				logRetCont.avg/logRetCont.accNum,
				logRetCont.max,
				int64(logRetCont.distri1),
				int64(logRetCont.distri2),
				int64(logRetCont.distri3))
			logStr += aTempStr
		}
	}

	runLogDir := zkconfig.InitConfMap["logBaseDir"]
	serviceName := zkconfig.ServiceConfMap[BizKey]["srvName"]
	rawStatFile := fmt.Sprintf("%s%s/%s", runLogDir, serviceName, zkconfig.InitConfMap["rawStatLog"])

	onWriteLog, err := os.OpenFile(rawStatFile, syscall.O_CREAT+syscall.O_WRONLY+syscall.O_APPEND, 0666)
	defer onWriteLog.Close()
	if err != nil {
		panic(err)
		return
	}
	onWriteLog.Write([]byte(logStr))
	onWriteLog.Write([]byte("\n"))
}
