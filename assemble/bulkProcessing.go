package assemble

import (
	"fmt"
	"regexp"
	"statUpload/goLog"
	"strconv"
	"strings"
)

type retCont struct {
	accNum  float64 `该返回码出现的次数`
	total   float64 `该返回码被访问的次数`
	avg     float64 `该返回码延时(或其他含义)平均值`
	max     float64 `该返回码延时(或其他含义)最大值`
	min     float64 `该返回码延时(或其他含义)最小值`
	distri1 float64 `该返回码延时分布(或其他含义)大于A范围次数`
	distri2 float64 `该返回码延时分布(或其他含义)大于B范围次数`
	distri3 float64 `该返回码延时分布(或其他含义)大于C范围次数`
}

//第一个key为string:appid和命令字,例如:290001_CMD_GET
//第二个key为int64:表示返回码,如0,3001,3002
//retCont:其他信息(在retCont中定义)
type statInfo map[string]map[int64]*retCont

var CMDREG = regexp.MustCompile(`\d{4,}$`)

//对statlog日志进行粗处理,将结果都存放在变量eachRecord中
func statFmtInit(eachRecord *statInfo, eachLine []string, biz string) {
	var retTemp retCont
	var err error

	//过滤不正确的命令字
	cmdWord := strings.TrimSpace(eachLine[0])
	aList := strings.SplitN(cmdWord, "_", 2)
	if 2 == len(aList) {
		if CMDREG.MatchString(aList[1]) {
			return
		}
	}

	//某一返回码对应的访问量
	retTemp.total, err = strconv.ParseFloat(strings.TrimSpace(eachLine[2]), 64)
	if err != nil {
		retTemp.total = 0.0
		msg := fmt.Sprintf("convert total:<%s><%s>", eachLine[2], err)
		goLog.SendLog(msg, "ERROR", biz)
	}
	//某一返回码对应的平均延时
	retTemp.avg, err = strconv.ParseFloat(strings.TrimSpace(eachLine[4]), 64)
	if err != nil {
		retTemp.avg = 0.0
		msg := fmt.Sprintf("convert avg:<%s><%s>", eachLine[4], err)
		goLog.SendLog(msg, "ERROR", biz)
	}
	//某一返回码对应的最大延时
	retTemp.max, err = strconv.ParseFloat(strings.TrimSpace(eachLine[5]), 64)
	if err != nil {
		retTemp.max = 0.0
		msg := fmt.Sprintf("convert max:<%s><%s>", eachLine[5], err)
		goLog.SendLog(msg, "ERROR", biz)
	}
	//某一返回码对应的最小延时
	retTemp.min, err = strconv.ParseFloat(strings.TrimSpace(eachLine[6]), 64)
	if err != nil {
		retTemp.min = 0.0
		msg := fmt.Sprintf("convert min:<%s><%s>", eachLine[6], err)
		goLog.SendLog(msg, "ERROR", biz)
	}
	//某一返回码对应的延时分布
	retTemp.distri1, err = strconv.ParseFloat(strings.TrimSpace(eachLine[8]), 64)
	if err != nil {
		retTemp.distri1 = 0.0
		msg := fmt.Sprintf("convert distri1:<%s><%s>", eachLine[8], err)
		goLog.SendLog(msg, "ERROR", biz)
	}
	retTemp.distri2, err = strconv.ParseFloat(strings.TrimSpace(eachLine[9]), 64)
	if err != nil {
		retTemp.distri2 = 0.0
		msg := fmt.Sprintf("convert distri2:<%s><%s>", eachLine[9], err)
		goLog.SendLog(msg, "ERROR", biz)
	}
	retTemp.distri3, err = strconv.ParseFloat(strings.TrimSpace(eachLine[10]), 64)
	if err != nil {
		retTemp.distri3 = 0.0
		msg := fmt.Sprintf("convert distri3:<%s><%s>", eachLine[10], err)
		goLog.SendLog(msg, "ERROR", biz)
	}

	//命令字不存在
	if _, ok := (*eachRecord)[cmdWord]; !ok {
		(*eachRecord)[cmdWord] = make(map[int64]*retCont)
	}
	codeInt, _ := strconv.ParseInt(strings.TrimSpace(eachLine[1]), 10, 64)
	//返回码不存在
	if _, ok := (*eachRecord)[cmdWord][codeInt]; !ok {
		(*eachRecord)[cmdWord][codeInt] = new(retCont)
	}
	//该命令字的返回码出现次数
	(*eachRecord)[cmdWord][codeInt].accNum += 1
	//该命令字的返回码的访问量总和
	(*eachRecord)[cmdWord][codeInt].total += retTemp.total
	//该命令字的返回码平均延时总和
	(*eachRecord)[cmdWord][codeInt].avg += retTemp.avg
	//该命令字的返回码的最大延时
	if (*eachRecord)[cmdWord][codeInt].max <= retTemp.max {
		(*eachRecord)[cmdWord][codeInt].max = retTemp.max
	}
	//只有返回码为0的情况下,才对下面的参数进行统计
	if 0 == codeInt {
		(*eachRecord)[cmdWord][codeInt].distri1 += retTemp.distri1
		(*eachRecord)[cmdWord][codeInt].distri2 += retTemp.distri2
		(*eachRecord)[cmdWord][codeInt].distri3 += retTemp.distri3
	}
}

func bulking(biz string, buffCont []byte) *statInfo {
	if 0 == len(buffCont) {
		return nil
	}
	eachRecord := make(statInfo)
	for _, line := range strings.Split(string(buffCont), "\n") {
		if ignLine(line) {
			continue
		}
		recordList := strings.Split((strings.TrimSpace(line)), "|")
		//逐行处理
		statFmtInit(&eachRecord, recordList, biz)
	}
	buffCont = buffCont[:0]
	return &eachRecord
}

//过滤空白行,包含特殊字符串的行
func ignLine(line string) bool {
	switch {
	case 0 == len(line):
		return true
	case 0 == len(strings.TrimSpace(line)):
		return true
	case "\n" == line:
		return true
	case "\r\n" == line:
		return true
	case strings.Contains(line, "===="):
		return true
	case strings.Contains(line, "RESULT"):
		return true
	case strings.Contains(line, "----"):
		return true
	default:
		return false
	}
}
