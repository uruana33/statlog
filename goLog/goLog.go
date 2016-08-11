package goLog

import (
	"fmt"
	"os"
	"path"
	"runtime"
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
		done := make(chan int)

		go SendLogFromClient(mychan, done, bizkey)

		mychan <- aa
		done <- 1
	}
}

func SendLogFromClient(logMSG chan *LogCard, done chan int, bizKey string) {
DONE:
	for {
		select {
		case aa := <-logMSG:
			needLogging(aa, bizKey)
		case <-done:
			break DONE
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
	runMsgLog := fmt.Sprintf("%s%s/%s", runLogDir, bizKey, parseConfig.StatConfig["runMsgLog"])
	logFile, err := os.OpenFile(runMsgLog, syscall.O_CREAT+syscall.O_WRONLY+syscall.O_APPEND, 0666)
	defer logFile.Close()
	if err != nil {
		panic(err)
	}
	logFile.Write([]byte(ss))
	logFile.Close()
}
