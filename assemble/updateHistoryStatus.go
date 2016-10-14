package assemble

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"statUpload/goLog"
	"statUpload/isFile"
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

//type statusCodeInfo map[string]map[int64]*stFlag
type statusCodeInfo map[string][]int64

//状态码写入文件
func statusLoadToFile(mm *os.File, srcMap *statInfo) bool {

	statusBuffer := bytes.NewBuffer(make([]byte, 0))
	bwlogBuffer := bufio.NewWriter(statusBuffer)
	for mmCMD, mmRetMap := range *srcMap {
		if "ALL" == mmCMD {
			continue
		}
		for mmCode, mmCont := range mmRetMap {
			fmt.Fprintf(bwlogBuffer, "%s@%d@%d\n", mmCMD, mmCode, int64(mmCont.total))
		}
	}
	fmt.Fprint(bwlogBuffer, "\n")
	bwlogBuffer.Flush()
	_, err := statusBuffer.WriteTo(mm)
	if err != nil {
		return false
	}
	return true
}

func statusLoadToMap(statusFile string, biz string) *statusCodeInfo {
	//文件存在,读取文件中的配置,更新记录,或者新增记录
	rdStatFile, err := os.OpenFile(statusFile, syscall.O_RDWR, 0666)
	defer rdStatFile.Close()
	if err != nil {
		msg := fmt.Sprintf("Open statusFile failed:<%s>", statusFile)
		goLog.SendLog(msg, "ERROR", biz)
		return nil
	}

	//预读取:将文件中的记录存放到map中,格式:命令字@返回码@访问量
	codeMap := make(statusCodeInfo)
	bufReadLine := bufio.NewReader(rdStatFile)
	for {
		line, err := bufReadLine.ReadString('\n')
		if len(line) > 0 {
			cmdword, code := encapStatus(strings.Trim(line, "\n"))
			if 0 == len(cmdword) {
				continue
			}
			//新增命令字
			if _, ok := codeMap[cmdword]; !ok {
				codeMap[cmdword] = make([]int64, 0)
			}
			codeMap[cmdword] = append(codeMap[cmdword], code)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			msg := fmt.Sprintf("reading status file error:<%s --> %s>", statusFile, err.Error())
			goLog.SendLog(msg, "ERROR", biz)
		}
	}
	return &codeMap
}

func encapStatus(line string) (cmdword string, stcode int64) {
	//错误格式,忽略
	if 2 != strings.Count(line, "@") {
		return
	}
	arr := strings.Split(line, "@")
	cmdword = strings.TrimSpace(arr[0])
	if 0 == len(cmdword) {
		return
	}

	//过滤不正确的命令字
	aList := strings.SplitN(cmdword, "_", 2)
	if 2 == len(aList) {
		if CMDREG.MatchString(aList[1]) {
			return
		}
	}

	stcode, _ = strconv.ParseInt(strings.TrimSpace(arr[1]), 10, 64)
	return
}

//记录历史状态码的访问量
func logHisRetCodeAccess(eachRecord *statInfo, biz string) {

	runLogDir := parseConfig.StatConfig["logBaseDir"]
	statusFile := fmt.Sprintf("%s%s/%s", runLogDir, biz, parseConfig.StatConfig["statCode"])

	//文件不存在
	if !isFile.IsFileExist(statusFile) {
		mm, err := os.Create(statusFile)
		defer mm.Close()
		if err != nil {
			msg := fmt.Sprintf("Create status file fail:<%s>", statusFile)
			goLog.SendLog(msg, "ERROR", biz)
			return
		}
		if !statusLoadToFile(mm, eachRecord) {
			return
		}
	}

	myCodeMap := statusLoadToMap(statusFile, biz)
	if myCodeMap == nil {
		return
	}

	//eachRecord中存放最新状态,myCodeMap从文件中加载历史所有命令字状态
	//给eachRecord填充历史中遗漏的命令字,这些命令字统计量是清零的
	for cmdWord, retcodeList := range *myCodeMap {
		for _, retcode := range retcodeList {
			//新增历史命令字
			if _, ok := (*eachRecord)[cmdWord]; !ok {
				(*eachRecord)[cmdWord] = make(map[int64]*retCont)
			}
			//新增历史命令字,返回码对应的统计需清零
			if _, ok := (*eachRecord)[cmdWord][retcode]; !ok {
				(*eachRecord)[cmdWord][retcode] = new(retCont)
			}
		}
	}

	//将处理后的结果重新写回文件
	nn, err := os.OpenFile(statusFile, syscall.O_TRUNC|syscall.O_WRONLY, 0666)
	defer nn.Close()
	if err != nil {
		msg := "Open status file failed when re-writing..."
		goLog.SendLog(msg, "ERROR", biz)
		return
	}
	if !statusLoadToFile(nn, eachRecord) {
		return
	}
	return
}
