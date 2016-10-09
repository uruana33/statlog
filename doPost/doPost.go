package doPost

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"statUpload/assemble"
	"statUpload/filter"
	"statUpload/goLog"
	"statUpload/isFile"
	"statUpload/zkconfig"
)

type DataType struct {
	biz  string
	kind string
	data []byte
}

func dealchan(statData DataType) {

	if 0 == len(statData.data) {
		msg := fmt.Sprintf("nothing recive from chan...<biz:%s><kind:%s>", statData.biz, statData.kind)
		goLog.SendLog(msg, "INFO", statData.biz)
		return
	}

	falconData := make(chan assemble.FalconStruct)
	formatDone := make(chan struct{})
	go func() {
		defer close(falconData)
		defer close(formatDone)
		go assemble.PackageData(statData.data, statData.kind, statData.biz, falconData, formatDone)
		<-formatDone
	}()

	PostFalcon(falconData, statData.biz)

}

func Send() {

	/*
		go func() {
			http.ListenAndServe("10.112.15.48:80", nil)
		}()
	*/

	done := make(chan struct{})
	dataTypeChan := make(chan DataType)
	go func() {
		defer close(done)
		defer close(dataTypeChan)
		for {
			select {
			case statData := <-dataTypeChan:
				go dealchan(statData)
			case <-done:
				return
			default:
			}
		}
	}()

	dataKind(done, dataTypeChan)
}

//对不同的数据分类
func dataKind(done chan<- struct{}, dataTypeChan chan<- DataType) {

	for bizKey, bizMap := range zkconfig.ServiceConfMap {

		var statLog, statLogBak, statValueLog, statValueLogBak string
		statLogDir := bizMap["logDir"]
		if 0 != len(bizMap["statLogFile"]) {
			statLog = fmt.Sprintf("%s%s/%s", statLogDir, bizKey, bizMap["statLogFile"])
		} else {
			statLog = fmt.Sprintf("%s%s/%s", statLogDir, bizKey, "no-stat-log")
		}
		if 0 != len(bizMap["statLogFileBak"]) {
			statLogBak = fmt.Sprintf("%s%s/%s", statLogDir, bizKey, bizMap["statLogFileBak"])
		} else {
			statLogBak = fmt.Sprintf("%s%s/%s", statLogDir, bizKey, "no-stat1-log")
		}

		if 0 != len(bizMap["statValueLogFile"]) {
			statValueLog = fmt.Sprintf("%s%s/%s", statLogDir, bizKey, bizMap["statValueLogFile"])
		} else {
			statValueLog = fmt.Sprintf("%s%s/%s", statLogDir, bizKey, "no-statvalue-log")
		}
		if 0 != len(bizMap["statValueLogFileBak"]) {
			statValueLogBak = fmt.Sprintf("%s%s/%s", statLogDir, bizKey, bizMap["statValueLogFileBak"])
		} else {
			statValueLogBak = fmt.Sprintf("%s%s/%s", statLogDir, bizKey, "no-statvalue1-log")
		}

		statFile := make([]string, 0)
		statValueFile := make([]string, 0)
		statFile = append(statFile, statLog)
		statFile = append(statFile, statLogBak)
		statValueFile = append(statValueFile, statValueLog)
		statValueFile = append(statValueFile, statValueLogBak)

		recordBufferNormal := make([]byte, 0)
		for _, file := range statFile {
			if !isFile.IsFileExist(file) {
				msg := fmt.Sprintf("file not exist:<%s>", file)
				goLog.SendLog(msg, "ERROR", bizKey)
				continue
			}
			tempbuff := filter.ReadRecord(file, bizKey)
			recordBufferNormal = append(recordBufferNormal, tempbuff...)
		}
		if 0 == len(recordBufferNormal) {
			msg := fmt.Sprintf("Normal stat.log not useful data current...<biz:%s>", bizKey)
			goLog.SendLog(msg, "INFO", bizKey)
			continue
		}
		aa := new(DataType)
		aa.biz = bizKey
		aa.kind = "TypeNormal"
		aa.data = recordBufferNormal
		dataTypeChan <- *aa

		if 0 == len(statValueLog) || 0 == len(statValueFile) {
			continue
		}
		var exist bool
		recordBufferSpecial := make([]byte, 0)
		for _, valuefile := range statValueFile {
			if !isFile.IsFileExist(valuefile) {
				msg := fmt.Sprintf("file not exist:<%s>", valuefile)
				goLog.SendLog(msg, "ERROR", bizKey)
				exist = false
				continue
			}
			exist = true
			tempvaluebuff := filter.ReadRecord(valuefile, bizKey)
			recordBufferSpecial = append(recordBufferSpecial, tempvaluebuff...)
		}
		if !exist && (0 == len(recordBufferSpecial)) {
			continue
		}
		if 0 == len(recordBufferSpecial) {
			msg := fmt.Sprintf("Specail stat_value.log not useful data current...<biz:%s>", bizKey)
			goLog.SendLog(msg, "INFO", bizKey)
			continue
		}
		bb := new(DataType)
		bb.biz = bizKey
		bb.kind = "TypeSpecial"
		bb.data = recordBufferSpecial
		dataTypeChan <- *bb
	}
	done <- struct{}{}
}

func PostFalcon(falconData <-chan assemble.FalconStruct, bizKey string) {

	url := "http://127.0.0.1:1988/v1/push"
	postList := make([]assemble.FalconStruct, 0)
	for postData := range falconData {
		postList = append(postList, postData)
		if len(postList) == 20 {
			data, _ := json.Marshal(postList)
			res, err := http.Post(url, "application/json;charset=utf-8", bytes.NewBuffer(data))
			if err != nil {
				goLog.SendLog(err.Error(), "ERROR", bizKey)
			}
			postList = postList[:0]
			res.Body.Close()
		}

	}
	//post剩下的
	data, _ := json.Marshal(postList[:len(postList)])
	res, err := http.Post(url, "application/json;charset=utf-8", bytes.NewBuffer(data))
	if err != nil {
		goLog.SendLog(err.Error(), "ERROR", bizKey)
	}
	res.Body.Close()
	postList = postList[:0]

	return
}
