package filter

import (
	"os"
	"regexp"
	"syscall"
	"time"
)

/*
function:从日志文件中筛选需要的记录内容(上一分钟的内容)
inpurt:stat日志文件名
output:上一分钟的日志内容,string类型的buffer
*/

func ReadRecord(fileName string, biz string) []byte {

	ff, err := os.OpenFile(fileName, syscall.O_RDONLY, 0666)
	defer ff.Close()
	if err != nil {
		return nil
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
	preTimeStrList := make([]int64, 0)
	//存放所有当前分钟时间字符串所在行的起始位置,按出现的倒序存放
	curTimeStrList := make([]int64, 0)
	reTitle, _ := regexp.Compile(`=+`)
	reTimeStr, _ := regexp.Compile(`[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}`)
	strBuff := make([]byte, 100)
	tPreMin := time.Now().Unix() - 60
Over:
	for {
		ff.Seek(-allScanRange, os.SEEK_END)
		//获取符号=所在的位置,返回所有匹配字符串的起始索引和结束索引
		allIndex := reTitle.FindAllStringIndex(string(buff), -1)
		//仅扫描当前buffer块
		for ii := len(allIndex) - 1; ii >= 0; ii-- {
			//计算符号=所处的绝对位置,相对于文件尾
			//绝对位置(相对文件末尾)=之前块大小+本次beg位置到本次块末尾之间的大小
			begTemp := (allScanRange - int64(curBlockSize)) + int64(curBlockSize-allIndex[ii][0])
			ff.Seek(-begTemp, os.SEEK_END)

			//读取特定长度,用于匹配时间
			length, _ := ff.Read(strBuff)
			timeStr := reTimeStr.FindString(string(strBuff[:length]))
			if 0 == len(timeStr) {
				continue
			}
			timeInt := checkTime(tPreMin, timeStr)
			//该时间字符串是老时间
			if timeInt < 0 {
				break Over
			}
			//该时间字符串是当前时间
			if timeInt > 0 {
				curTimeStrList = append(curTimeStrList, begTemp)
				continue
			}
			//该时间字符串是前一分钟时间
			if timeInt == 0 {
				preTimeStrList = append(preTimeStrList, begTemp)
				continue
			}
		}

		//判断是否完成整个文件的扫描
		if flag {
			break Over
		}

		allScanRange += int64(curBlockSize)
		//判断绝对位置是否已超出文件大小
		if allScanRange > fileSize {
			flag = true
			lastSize := fileSize - (allScanRange - int64(curBlockSize))
			allScanRange = fileSize
			//重新调整buffer大小,避免溢出
			buff = buff[:lastSize]
		}
		ff.Seek(-allScanRange, os.SEEK_END)
		curBlockSize, _ = ff.Read(buff)
	}

	buff = buff[:0]

	//无记录
	if len(preTimeStrList) == 0 {
		return nil
	}

	var begPosInFile int64 = 0
	var endPosInFile int64 = 0

	//相对于文件末尾,确定前一分钟记录的起始位置,选出最大值
	for _, locate := range preTimeStrList {
		if 0 == begPosInFile || locate > begPosInFile {
			begPosInFile = locate
		}
	}

	//相对于文件末尾,确定本分钟记录的终止位置,选出最大值
	for _, locate := range curTimeStrList {
		if 0 == endPosInFile || locate > endPosInFile {
			endPosInFile = locate
		}
	}

	blockSize := begPosInFile - endPosInFile
	blockBuff := make([]byte, blockSize)
	ff.Seek(-begPosInFile, os.SEEK_END)
	nn, _ := ff.Read(blockBuff)
	return blockBuff[:nn]
}

func timeStr2Int(timeStr string) int64 {
	timeLayOut := "2006-01-02 15:04"
	tt, _ := time.Parse(timeLayOut, timeStr)
	return tt.Unix() - 8*60*60
}

func checkTime(tPreMin int64, timeStr string) int {
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
