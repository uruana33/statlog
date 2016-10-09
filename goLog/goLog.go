package goLog

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"statUpload/isFile"
	"statUpload/parseConfig"
	"syscall"
	"time"
)

type LogCard struct {
	LogItemTime     string
	LogItemLevel    string
	LogItemMoudle   string
	LogItemFileName string
	LogItemLineNo   int
	LogItemMSG      string
}

func GetLogCard() *LogCard {
	return new(LogCard)
}

func SendLog(msg string, logLevel string, bizkey string) {

	funcName, file, line, ok := runtime.Caller(1)
	if ok {
		aa := GetLogCard()
		aa.LogItemTime = time.Now().Format("2006-01-02 15:04")
		aa.LogItemLevel = logLevel
		aa.LogItemMoudle = runtime.FuncForPC(funcName).Name()
		aa.LogItemFileName = path.Base(file)
		aa.LogItemLineNo = line
		aa.LogItemMSG = msg

		mychan := make(chan *LogCard)
		done := make(chan struct{})
		defer close(done)
		defer close(mychan)

		go SendLogFromClient(mychan, done, bizkey)

		mychan <- aa
		done <- struct{}{}
	}
}

func SendLogFromClient(logMSG <-chan *LogCard, done <-chan struct{}, bizKey string) {
	for {
		select {
		case aa := <-logMSG:
			needLogging(aa, bizKey)
		case <-done:
			return
		default:
		}
	}
}

func needLogging(logData *LogCard, bizKey string) {

	ss := fmt.Sprintf("[%s][%s][%s][%s][%d][%s]\n",
		(*logData).LogItemTime,
		(*logData).LogItemLevel,
		(*logData).LogItemMoudle,
		(*logData).LogItemFileName,
		(*logData).LogItemLineNo,
		(*logData).LogItemMSG)

	runLogDir := parseConfig.StatConfig["logBaseDir"]
	//目录不存在
	if !isFile.IsDirExist(runLogDir) {
		err := os.Mkdir(runLogDir, 0755)
		if err != nil {
			panic(err)
		}
	}
	runLogBizDir := fmt.Sprintf("%s%s", runLogDir, bizKey)
	if !isFile.IsDirExist(runLogBizDir) {
		err := os.Mkdir(runLogBizDir, 0755)
		if err != nil {
			panic(err)
		}
	}

	runMsgLog := fmt.Sprintf("%s%s/%s", runLogDir, bizKey, parseConfig.StatConfig["runMsgLog"])
	logFile, err := os.OpenFile(runMsgLog, syscall.O_CREAT+syscall.O_WRONLY+syscall.O_APPEND, 0666)
	defer logFile.Close()
	if err != nil {
		panic(err)
	}
	logFile.Write([]byte(ss))
	logFile.Close()
}
