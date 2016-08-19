package assemble

import (
	"fmt"
	"os"
	"statUpload/goLog"
	"statUpload/isFile"
	"statUpload/parseConfig"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func writeRawLog(records *statInfo, kind string) {

	if nil == records {
		goLog.SendLog("not any data recive from bulking function", "ERROR", BizKey)
		return
	}
	//获取上一分钟
	timeLayOut := "2006-01-02 15:04"
	tPreMin := time.Now().Unix() - 60
	tPreMinStr := time.Unix(tPreMin, 0).Format(timeLayOut)
	var logStr string
	for logCmd, logMap := range *records {
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
	//目录不存在
	runLogDir := parseConfig.StatConfig["logBaseDir"]
	if !isFile.IsDirExist(runLogDir) {
		err := os.Mkdir(runLogDir, 0755)
		if err != nil {
			panic(err)
		}
	}
	runLogBizDir := fmt.Sprintf("%s%s", runLogDir, BizKey)
	if !isFile.IsDirExist(runLogBizDir) {
		err := os.Mkdir(runLogBizDir, 0755)
		if err != nil {
			panic(err)
		}
	}

	var rawStatFile string
	switch kind {
	case "TypeSpecial":
		rawStatFile = fmt.Sprintf("%s%s/%s_value_%s", runLogDir, BizKey, BizKey, parseConfig.StatConfig["rawStatLog"])
	case "TypeNormal":
		rawStatFile = fmt.Sprintf("%s%s/%s_%s", runLogDir, BizKey, BizKey, parseConfig.StatConfig["rawStatLog"])
	}

	if !isFile.IsFileExist(rawStatFile) {
		_, err := os.Create(rawStatFile)
		if err != nil {
			msg := fmt.Sprintf("Create file faild:<%s>", rawStatFile)
			goLog.SendLog(msg, "ERROR", BizKey)
			panic(msg)
		}
	}

	//文件回滚清空
	resetSize := parseConfig.StatConfig["truncateSize"]
	resetSizeInt64, err := strconv.ParseInt(strings.TrimSpace(resetSize), 10, 64)
	if err != nil {
		goLog.SendLog("Read truncate size error", "ERROR", BizKey)
		resetSizeInt64 = 0
	}
	fstat, _ := os.Stat(rawStatFile)
	fileSize := fstat.Size()
	onWriteLog := new(os.File)
	if fileSize <= resetSizeInt64 {
		onWriteLog, err = os.OpenFile(rawStatFile, syscall.O_CREAT|syscall.O_WRONLY|syscall.O_APPEND, 0666)
	} else {
		if resetSizeInt64 != 0 {
			onWriteLog, err = os.OpenFile(rawStatFile, syscall.O_CREAT|syscall.O_WRONLY|syscall.O_APPEND|syscall.O_TRUNC, 0666)
		} else {
			onWriteLog, err = os.OpenFile(rawStatFile, syscall.O_CREAT|syscall.O_WRONLY|syscall.O_APPEND, 0666)
		}
	}
	defer onWriteLog.Close()
	if err != nil {
		goLog.SendLog("Could not write rawstat data into file.", "ERROR", BizKey)
		return
	}

	onWriteLog.Write([]byte(logStr))
	onWriteLog.Write([]byte("\n"))
}
