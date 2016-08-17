package doPost

import (
	"bytes"
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

var dataTypeChan = make(chan DataType)
var done = make(chan int)

func dealchan(statData DataType) {
	rawBiz := statData.biz
	rawType := statData.kind
	rawData := statData.data
	if 0 == len(rawData) {
		msg := fmt.Sprintf("not any data recive from chan...<biz:%s><kind:%s>", rawBiz, rawType)
		goLog.SendLog(msg, "INFO", rawBiz)
		return
	}
	data := assemble.PackageData(rawData, rawType, rawBiz)
	if 0 == len(data) {
		msg := fmt.Sprintf("not any data for post to falcon...<biz:%s><kind:%s>", rawBiz, rawType)
		goLog.SendLog(msg, "INFO", rawBiz)
		return
	}
	go PostFalcon(data, rawBiz)
}

func Send() {
	go func() {
	DONE:
		for {
			select {
			case statData := <-dataTypeChan:
				dealchan(statData)
			case <-done:
				break DONE
			}
		}
	}()

	dataKind()
}

func dataKind() {

	for bizKey, bizMap := range zkconfig.ServiceConfMap {
		statLogDir := bizMap["logDir"]
		statLog := fmt.Sprintf("%s%s/%s", statLogDir, bizKey, bizMap["statLogFile"])
		statLogBak := fmt.Sprintf("%s%s/%s", statLogDir, bizKey, bizMap["statLogFileBak"])
		statValueLog := fmt.Sprintf("%s%s/%s", statLogDir, bizKey, bizMap["statValueLogFile"])
		statValueLogBak := fmt.Sprintf("%s%s/%s", statLogDir, bizKey, bizMap["statValueLogFileBak"])

		statFile := make([]string, 0)
		statValueFile := make([]string, 0)
		statFile = append(statFile, statLog)
		statFile = append(statFile, statLogBak)
		statValueFile = append(statValueFile, statValueLog)
		statValueFile = append(statValueFile, statValueLogBak)

		var recordBufferNormal string
		var recordBufferSpecial string
		//go func() {
		for _, file := range statFile {
			if !isFile.IsFileExist(file) {
				msg := fmt.Sprintf("file not exist:<%s>", file)
				goLog.SendLog(msg, "ERROR", bizKey)
				continue
			}
			recordBufferNormal += string(filter.ReadRecord(file, bizKey))
		}
		if 0 == len(recordBufferNormal) {
			msg := fmt.Sprintf("Normal stat.log not useful data current...<biz:%s>", bizKey)
			goLog.SendLog(msg, "INFO", bizKey)
			continue
		}
		aa := new(DataType)
		aa.biz = bizKey
		aa.kind = "TypeNormal"
		aa.data = []byte(recordBufferNormal)
		dataTypeChan <- *aa
		//}()

		//	go func() {
		for _, file := range statValueFile {
			if !isFile.IsFileExist(file) {
				msg := fmt.Sprintf("file not exist:<%s>", file)
				goLog.SendLog(msg, "ERROR", bizKey)
				continue
			}
			recordBufferSpecial += string(filter.ReadRecord(file, bizKey))
		}
		if 0 == len(recordBufferSpecial) {
			msg := fmt.Sprintf("Specail stat_value.log not useful data current...<biz:%s>", bizKey)
			goLog.SendLog(msg, "INFO", bizKey)
			continue
		}
		bb := new(DataType)
		bb.biz = bizKey
		bb.kind = "TypeSpecial"
		bb.data = []byte(recordBufferSpecial)
		dataTypeChan <- *bb
		//	}()
	}
	done <- 1
}

func PostFalcon(data []byte, bizKey string) {
	if len(data) == 0 {
		return
	}
	url := "http://127.0.0.1:1988/v1/push"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		goLog.SendLog(err.Error(), "ERROR", bizKey)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	myPost := &http.Client{}
	resp, err := myPost.Do(req)
	defer resp.Body.Close()
	if err != nil {
		goLog.SendLog(err.Error(), "ERROR", bizKey)
		return
	}
}
