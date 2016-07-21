package assemble

import (
	"fmt"
	"os"
	"statUpload/mylog"
	"statUpload/zkconfig"
	"strconv"
	"strings"
	"syscall"
)

//用于记录该命令字的某一返回码是否被更新
type stFlag struct {
	stacc int64
	flag  bool
}

//记录历史状态码的访问量
func logHisRetCodeAccess() (myCodeMap map[string]map[int64]*stFlag) {

	runLogDir := zkconfig.InitConfMap["logBaseDir"]
	serviceName := zkconfig.ServiceConfMap[BizKey]["srvName"]
	statusFile := fmt.Sprintf("%s%s/%s", runLogDir, serviceName, zkconfig.InitConfMap["statCode"])

	//文件不存在,新建,并记录状态
	_, err := os.Stat(statusFile)
	if err != nil && os.IsNotExist(err) {
		mm, err := os.Create(statusFile)
		defer mm.Close()

		if err != nil {
			runLog.MyLog(BizKey, runLog.ERROR, "Create status file failed !!!", 1)
			return
		}

		var mmLogStr string
		for mmCMD, mmRetMap := range eachRecord {
			if "ALL" == mmCMD {
				continue
			}
			var mmStr string
			for mmCode, mmCont := range mmRetMap {
				mmStr = fmt.Sprintf("%s@%d@%d\n", mmCMD, mmCode, int64(mmCont.total))
				mmLogStr += mmStr
			}
		}

		_, err1 := mm.WriteString(mmLogStr)
		if err1 != nil {
			runLog.MyLog(BizKey, runLog.ERROR, "Write to status file failed !!!", 1)
			return
		}
		return
	}

	//文件存在,读取文件中的配置,更新记录,或者新增记录
	rdStatFile, err := os.OpenFile(statusFile, syscall.O_RDWR, 0666)
	defer rdStatFile.Close()

	if err != nil {
		runLog.MyLog(BizKey, runLog.ERROR, "Open status file failed !!!", 1)
		return
	}

	//将文件内容全部读入buffer
	ll, _ := os.Stat(statusFile)
	rdBuffer := make([]byte, ll.Size())
	realSize, _ := rdStatFile.Read(rdBuffer)

	//预读取:将文件中的记录存放到map(myCodeMap)中
	//格式:命令字@返回码@访问量
	myCodeMap = make(map[string]map[int64]*stFlag)
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
		if _, ok := myCodeMap[cmdword]; !ok {
			myCodeMap[cmdword] = make(map[int64]*stFlag)
		}
		aFlag := new(stFlag)
		aFlag.stacc = stacc
		aFlag.flag = false
		myCodeMap[cmdword][stcode] = aFlag
	}

	//对比eachRecord,更新(或新增)预处理得到的map:myCodeMap
	for cmd, retmap := range eachRecord {
		//新增命令字
		if _, ok := myCodeMap[cmd]; !ok {
			myCodeMap[cmd] = make(map[int64]*stFlag)
		}
		//更新eachRecord存在的命令字
		for code, cont := range retmap {
			//新增返回码
			if _, ok := myCodeMap[cmd][code]; !ok {
				myCodeMap[cmd][code] = new(stFlag)
			}
			//更新操作
			myCodeMap[cmd][code].stacc = int64(cont.total)
			myCodeMap[cmd][code].flag = true
		}
	}

	//eachRecord中不存在的命令字,即历史记录的命令字,需要清零
	for _, codeMap := range myCodeMap {
		for _, mystflag := range codeMap {
			if !mystflag.flag {
				mystflag.stacc = 0
			}
		}
	}

	//将处理后的结果重新写回文件
	nn, err := os.OpenFile(statusFile, syscall.O_TRUNC|syscall.O_WRONLY, 0666)
	defer nn.Close()
	if err != nil {
		runLog.MyLog(BizKey, runLog.ERROR, "Open status file after dealing failed !!!", 1)
		return
	}
	var nnLogStr string
	for nnCMD, nnCodeMap := range myCodeMap {
		if "ALL" == nnCMD {
			continue
		}
		var nnStr string
		for nnCode, nnCont := range nnCodeMap {
			nnStr = fmt.Sprintf("%s@%d@%d\n", nnCMD, nnCode, int64(nnCont.stacc))
			nnLogStr += nnStr
		}
	}
	_, err2 := nn.WriteString(nnLogStr + "\n")
	if err2 != nil {
		runLog.MyLog(BizKey, runLog.ERROR, "Write NEW-STAT to status file failed !!!", 1)
	}
	return
}
