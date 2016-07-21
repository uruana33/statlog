package runLog

import (
	"fmt"
	"log"
	"os"
	"statUpload/zkconfig"
	"syscall"
)

const (
	DEBUG string = "[DEBUG] "
	INFO  string = "[INFO] "
	ERROR string = "[ERROR] "
)

func MyLog(bizKey string, level string, logCont string, flag int) {

	runLogDir := zkconfig.InitConfMap["logBaseDir"]
	runMsgLog := fmt.Sprintf("%s%s/%s", runLogDir, bizKey, zkconfig.InitConfMap["runMsgLog"])

	logFile, err := os.OpenFile(runMsgLog, syscall.O_CREAT+syscall.O_WRONLY+syscall.O_APPEND, 0666)
	defer logFile.Close()

	if err != nil {
		panic("create log file error!!:" + runMsgLog)
	}

	runLog := log.New(logFile, "[xxxx] ", log.LstdFlags|log.Llongfile)

	if 0 == flag {
		runLog.SetFlags(0)
	} else {
		runLog.SetFlags(runLog.Flags() | log.LstdFlags | log.Llongfile)
	}
	runLog.SetPrefix(level)
	runLog.Println(logCont)
}
