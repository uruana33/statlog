package filter

import (
	"os"
	"regexp"
	"statUpload/mylog"
	"syscall"
	"time"
)

/*
function:从日志文件中筛选需要的记录内容(上一分钟的内容)
inpurt:stat日志文件名
output:上一分钟的日志内容,string类型的buffer
*/
func ReadRecord(fileName string, bizKey string) (recordBuff string) {
	ff, err := os.OpenFile(fileName, syscall.O_RDONLY, 0666)
	defer ff.Close()
	if err != nil {
		runLog.MyLog(bizKey, runLog.ERROR, "Open statlog Failed !!!", 1)
		return
	}

	fstat, _ := os.Stat(fileName)
	fileSize := fstat.Size()

	//预读取大小2MB
	var scanRange int64 = 2 * 1024 * 1024
	buff := make([]byte, scanRange)
	ff.Seek(-scanRange, os.SEEK_END)
	curBlockSize, _ := ff.Read(buff)
	allScanRange := int64(curBlockSize)
	flag := false

	//存放所有前一分钟时间字符串所在行的起始位置,按出现的倒序存放
	var preTimeStrList []int64
	//存放所有当前分钟时间字符串所在行的起始位置,按出现的倒序存放
	var curTimeStrList []int64
Over:
	for {
		ff.Seek(-allScanRange, os.SEEK_END)
		//获取符号=所在的位置,返回所有匹配字符串的起始索引和结束索引
		re, _ := regexp.Compile(`=+`)
		allIndex := re.FindAllStringIndex(string(buff), -1)
		for ii := len(allIndex) - 1; ii >= 0; ii-- {
			//计算符号=所处的绝对位置,相对于文件尾
			//绝对位置(相对文件末尾)=之前块大小+本次beg位置到本次块末尾之间的大小
			begTemp := (allScanRange - int64(curBlockSize)) + int64(curBlockSize-allIndex[ii][0])
			strBuff := make([]byte, 100)
			ff.Seek(-begTemp, os.SEEK_END)
			length, _ := ff.Read(strBuff)
			reTemp, _ := regexp.Compile(`[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}`)
			timeStr := reTemp.FindString(string(strBuff[:length]))
			if 0 == len(timeStr) {
				continue
			}
			//该时间字符串是老时间
			if checkTime(timeStr) < 0 {
				break Over
			}
			//该时间字符串是当前时间
			if checkTime(timeStr) > 0 {
				curTimeStrList = append(curTimeStrList, begTemp)
				continue
			}
			//该时间字符串是前一分钟时间
			if checkTime(timeStr) == 0 {
				preTimeStrList = append(preTimeStrList, begTemp)
				continue
			}
		}

		if !flag {
			break Over
		}

		allScanRange += int64(curBlockSize)
		if allScanRange > fileSize {
			flag = true
			lastSize := fileSize - (allScanRange - int64(curBlockSize))
			allScanRange = fileSize
			buff = make([]byte, lastSize)
		}
		ff.Seek(-allScanRange, os.SEEK_END)
		curBlockSize, _ = ff.Read(buff)
	}

	var begPosInFile int64 = 0
	var endPosInFile int64 = 0

	for _, locate := range preTimeStrList {
		if 0 == begPosInFile || locate > begPosInFile {
			begPosInFile = locate
		}
	}

	for _, locate := range curTimeStrList {
		if 0 == endPosInFile || locate > endPosInFile {
			endPosInFile = locate
		}
	}

	dataSize := begPosInFile - endPosInFile
	tmpbuf := make([]byte, dataSize)
	ff.Seek(-begPosInFile, os.SEEK_END)
	nn, _ := ff.Read(tmpbuf)
	recordBuff = string(tmpbuf[:nn])
	return
}

func timeStr2Int(timeStr string) int64 {
	timeLayOut := "2006-01-02 15:04"
	tt, _ := time.Parse(timeLayOut, timeStr)
	return tt.Unix() - 8*60*60
}

func checkTime(timeStr string) int {
	tPreMin := time.Now().Unix() - 60
	logTimeInt := timeStr2Int(timeStr)
	switch {
	case logTimeInt < tPreMin:
		return -1
	case logTimeInt > tPreMin:
		return 1
	case logTimeInt == tPreMin:
		return 0
	default:
		return -1
	}
}
