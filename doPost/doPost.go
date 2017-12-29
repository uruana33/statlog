package doPost

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"statlog/assemble"
	"statlog/filter"
	"statlog/goLog"
	"statlog/isFile"
	"statlog/zkconfig"
	"io/ioutil"
)

type DataType struct {
	biz  string
	kind string
	data []byte
}

func dealchan(statData *DataType) {

	if 0 == len(statData.data) {
		msg := fmt.Sprintf("nothing recive from chan...<biz:%s><kind:%s>", statData.biz, statData.kind)
		goLog.SendLog(msg, "INFO", statData.biz)
		return
	}

	falconData := make(chan assemble.FalconStruct, 100)
	done := make(chan struct{})
	defer close(falconData)
	defer close(done)
	go assemble.PackageData(statData.data, statData.kind, statData.biz, falconData, done)
	PostFalcon(falconData, done, statData.biz)
}


func start(bizKey string, logPath string, logName string) {
	if 0 == len(logName) {
		return
	}
	statLog := fmt.Sprintf("%s%s/%s", logPath, bizKey, logName)
	if !isFile.IsFileExist(statLog) {
		msg := fmt.Sprintf("file not exist:<%s>", statLog)
		goLog.SendLog(msg, "ERROR", bizKey)
		return
	}
	msg := fmt.Sprintf("read useful records, <biz:%s>, <file:%s>", bizKey, statLog)
	goLog.SendLog(msg, "INFO", bizKey)
	//提取数据块耗时
	before := time.Now().Nanosecond()
	result := filter.ReadRecord(statLog)
	after := time.Now().Nanosecond()
	msg = fmt.Sprintf("biz %s filter records size %d, use time %d ms", bizKey, len(result), (after-before)/1000000)
	goLog.SendLog(msg, "ERROR", bizKey)
	if len(result) > 0 {
		aData := new(DataType)
		aData.biz = bizKey
		aData.kind = "TypeNormal"
		aData.data = result

		go dealchan(aData)
	}
}

func PostFalcon(falconData <-chan assemble.FalconStruct, done <- chan struct{}, bizKey string) {
	url := "http://127.0.0.1:1988/v1/push"
	contentType := "application/json;charset=utf-8"
	postList := make([]assemble.FalconStruct, 500)
	DD:
	for {
		select {
		case fData := <-falconData:
			postList = append(postList, fData)
		case <-done:
			break DD
		}
	}
	data, _ := json.Marshal(postList)
	body := bytes.NewReader(data)
	request, err := http.NewRequest("POST", url, body)
	request.Header.Set("Content-Type", contentType)
	if err != nil {
		postList = postList[:0]
		return
	}

	var resp *http.Response
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		postList = postList[:0]
		return
	}
	defer resp.Body.Close()

	rspBody, _ := ioutil.ReadAll(resp.Body)
	goLog.SendLog(string(rspBody), "ERROR", bizKey)
	postList = postList[:0]
	return
}



func Send() {

	/*
		go func() {
			http.ListenAndServe("10.114.24.9:80", nil)
		}()
	*/
	for bizKey, bizMap := range zkconfig.ServiceConfMap {
		start(bizKey, bizMap["logDir"], bizMap["statLogFile"])
		start(bizKey, bizMap["logDir"], bizMap["statLogFileBak"])
		start(bizKey, bizMap["logDir"], bizMap["statValueLogFile"])
		start(bizKey, bizMap["logDir"], bizMap["statValueLogFileBak"])
	}
}