package assemble

import (
	"fmt"
	"os"
	"statUpload/goLog"
	"statUpload/parseConfig"
	"strconv"
	"strings"
	"syscall"
)

//用于记录该命令字的某一返回码是否被更新
type stFlag struct {
	stacc int64
	flag  bool
}

type statusCodeInfo map[string]map[int64]*stFlag

//状态码写入文件
func statusLoadToFile(mm *os.File, srcMap interface{}) bool {

	switch records := srcMap.(type) {
	case *statInfo:
		var mmLogStr string
		for mmCMD, mmRetMap := range *records {
			if "ALL" == mmCMD {
				continue
			}
			var mmStr string
			for mmCode, mmCont := range mmRetMap {
				mmStr = fmt.Sprintf("%s@%d@%d\n", mmCMD, mmCode, int64(mmCont.total))
				mmLogStr += mmStr
			}
		}
		_, err1 := mm.WriteString(mmLogStr + "\n")
		if err1 != nil {
			msg := "Write to status file failed !!!"
			goLog.SendLog(msg, "ERROR", BizKey)
			return false
		}
		return true

	case *statusCodeInfo:
		var mmLogStr string
		for mmCMD, mmRetMap := range *records {
			if "ALL" == mmCMD {
				continue
			}
			var mmStr string
			for mmCode, mmCont := range mmRetMap {
				mmStr = fmt.Sprintf("%s@%d@%d\n", mmCMD, mmCode, int64(mmCont.stacc))
				mmLogStr += mmStr
			}
		}
		_, err1 := mm.WriteString(mmLogStr + "\n")
		if err1 != nil {
			msg := "Write to status file failed !!!"
			goLog.SendLog(msg, "ERROR", BizKey)
			return false
		}
		return true
	}
	return false
}

func statusLoadToMap(statusFile string) (*statusCodeInfo, bool) {
	//文件存在,读取文件中的配置,更新记录,或者新增记录
	rdStatFile, err := os.OpenFile(statusFile, syscall.O_RDWR, 0666)
	defer rdStatFile.Close()
	if err != nil {
		msg := fmt.Sprintf("Open statusFile:<%s> failed!\n", statusFile)
		goLog.SendLog(msg, "ERROR", BizKey)
		return nil, false
	}
	ll, _ := os.Stat(statusFile)
	rdBuffer := make([]byte, ll.Size())
	realSize, _ := rdStatFile.Read(rdBuffer)
	//预读取:将文件中的记录存放到map中,格式:命令字@返回码@访问量
	codeMap := make(statusCodeInfo)
	for _, line := range strings.Split(string(rdBuffer[:realSize]), "\n") {
		//错误格式,忽略
		if 2 != strings.Count(line, "@") {
			continue
		}
		arr := strings.Split(line, "@")
		cmdword := strings.TrimSpace(arr[0])
		if 0 == len(cmdword) {
			continue
		}
		stcode, _ := strconv.ParseInt(strings.TrimSpace(arr[1]), 10, 64)
		stacc, _ := strconv.ParseInt(strings.TrimSpace(arr[2]), 10, 64)
		//新增命令字
		if _, ok := codeMap[cmdword]; !ok {
			codeMap[cmdword] = make(map[int64]*stFlag)
		}
		aFlag := new(stFlag)
		aFlag.stacc = stacc
		aFlag.flag = false
		codeMap[cmdword][stcode] = aFlag
	}
	return &codeMap, true
}

//记录历史状态码的访问量
func logHisRetCodeAccess(eachRecord *statInfo) {

	runLogDir := parseConfig.StatConfig["logBaseDir"]
	statusFile := fmt.Sprintf("%s%s/%s", runLogDir, BizKey, parseConfig.StatConfig["statCode"])
	_, err := os.Stat(statusFile)
	if err != nil && os.IsNotExist(err) {
		mm, err := os.Create(statusFile)
		defer mm.Close()
		if err != nil {
			msg := "Create status file failed !!!"
			goLog.SendLog(msg, "ERROR", BizKey)
			return
		}
		if !statusLoadToFile(mm, eachRecord) {
			return
		}
	}
	myCodeMap := new(statusCodeInfo)
	var ok bool
	if myCodeMap, ok = statusLoadToMap(statusFile); !ok {
		return
	}
	//eachRecord中存放最新状态,myCodeMap从文件中加载历史所有命令字状态
	//给eachRecord填充历史中遗漏的命令字,这些命令字统计量是清零的
	for cmd, stMap := range *myCodeMap {
		for retcode := range stMap {
			//新增历史命令字
			if _, ok := (*eachRecord)[cmd]; !ok {
				(*eachRecord)[cmd] = make(map[int64]*retCont)
			}
			//新增历史命令字的返回码
			if _, ok := (*eachRecord)[cmd][retcode]; !ok {
				(*eachRecord)[cmd][retcode] = new(retCont)
			}
		}
	}
	//将处理后的结果重新写回文件
	nn, err := os.OpenFile(statusFile, syscall.O_TRUNC|syscall.O_WRONLY, 0666)
	defer nn.Close()
	if err != nil {
		msg := "Open status file failed when re-writing..."
		goLog.SendLog(msg, "ERROR", BizKey)
		return
	}
	if !statusLoadToFile(nn, eachRecord) {
		return
	}
	return
}
