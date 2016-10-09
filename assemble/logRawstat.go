package assemble

import (
	"bufio"
	"bytes"
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

func writeRawLog(records *statInfo, kind string, biz string) {

	if nil == records {
		goLog.SendLog("Not any data recive from bulking function", "ERROR", biz)
		return
	}

	//目录不存在
	runLogDir := parseConfig.StatConfig["logBaseDir"]
	if !isFile.IsDirExist(runLogDir) {
		err := os.Mkdir(runLogDir, 0755)
		if err != nil {
			msg := fmt.Sprintf("create dir err:<%s>", runLogDir)
			goLog.SendLog(msg, "ERROR", biz)
			return
		}
	}
	runLogBizDir := fmt.Sprintf("%s%s", runLogDir, biz)
	if !isFile.IsDirExist(runLogBizDir) {
		err := os.Mkdir(runLogBizDir, 0755)
		if err != nil {
			msg := fmt.Sprintf("create dir err:<%s>", runLogBizDir)
			goLog.SendLog(msg, "ERROR", biz)
			return
		}
	}

	var rawStatFile string
	switch kind {
	case "TypeSpecial":
		rawStatFile = fmt.Sprintf("%s%s/%s_value_%s", runLogDir, biz, biz, parseConfig.StatConfig["rawStatLog"])
	case "TypeNormal":
		rawStatFile = fmt.Sprintf("%s%s/%s_%s", runLogDir, biz, biz, parseConfig.StatConfig["rawStatLog"])
	}

	//文件不存在
	if !isFile.IsFileExist(rawStatFile) {
		ff, err := os.Create(rawStatFile)
		defer ff.Close()
		if err != nil {
			msg := fmt.Sprintf("Create file faild:<%s>", rawStatFile)
			goLog.SendLog(msg, "ERROR", biz)
			return
		}
	}

	//文件回滚清空
	resetSize := parseConfig.StatConfig["truncateSize"]
	resetSizeInt64, err := strconv.ParseInt(strings.TrimSpace(resetSize), 10, 64)
	if err != nil {
		goLog.SendLog("Read truncate size error", "ERROR", biz)
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
		goLog.SendLog("Could not write rawstat data into file.", "ERROR", biz)
		return
	}

	//获取上一分钟
	timeLayOut := "2006-01-02 15:04"
	tPreMin := time.Now().Unix() - 60
	tPreMinStr := time.Unix(tPreMin, 0).Format(timeLayOut)

	logBuffer := bytes.NewBuffer(make([]byte, 0))
	bwlogBuffer := bufio.NewWriter(logBuffer)

	for logCmd, logMap := range *records {
		for logRetcode, logRetCont := range logMap {
			//格式:时间 | 命令字 | 返回码 | 出现次数 | 访问量 | 平均延时 | 最大延时 | 大于10ms次数 | 大于100ms次数 | 大于500ms次数
			fmt.Fprintf(bwlogBuffer, "%s | %s | %d | %d | %d | %.3f | %.3f | %d | %d | %d\n",
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
		}
	}
	fmt.Fprint(bwlogBuffer, "\n")
	bwlogBuffer.Flush()
	_, err = logBuffer.WriteTo(onWriteLog)
	if err != nil {
		goLog.SendLog("Could not write rawstat data into file.", "ERROR", biz)
		return
	}
}
