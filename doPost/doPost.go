package doPost

import (
	"bytes"
	"fmt"
	"net/http"
	"statUpload/assemble"
	"statUpload/filter"
	"statUpload/zkconfig"
)

func Send() {

	for bizKey, bizMap := range zkconfig.ServiceConfMap {

		fileList := make([]string, 0)

		statLogDir := bizMap["logDir"]
		servicePort := bizMap["srvPort"]

		var statLog string
		var statLogBak string
		if 0 == len(servicePort) {
			statLog = fmt.Sprintf("%s%s/%s", statLogDir, bizKey, bizMap["statLogFile"])
			statLogBak = fmt.Sprintf("%s%s/%s", statLogDir, bizKey, bizMap["statLogFileBak"])
		} else {
			statLog = fmt.Sprintf("%s%s%s/%s", statLogDir, bizKey, servicePort, bizMap["statLogFile"])
			statLogBak = fmt.Sprintf("%s%s%s/%s", statLogDir, bizKey, servicePort, bizMap["statLogFileBak"])
		}

		fileList = append(fileList, statLog)
		fileList = append(fileList, statLogBak)

		var recordBuffer string

		for _, statFile := range fileList {
			recordBuffer += filter.ReadRecord(statFile, bizKey)
			if 0 == len(recordBuffer) {
				//runLog.MyLog(bizKey, runLog.ERROR, "Not Any Record(s) in stat file:"+statFile, 1)
				continue
			}
		}

		data := assemble.PackageData(recordBuffer, bizKey)

		url := "http://127.0.0.1:1988/v1/push"
		//jsonStr := []byte(data)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
		req.Header.Set("Content-Type", "application/json")

		myPost := &http.Client{}
		resp, err := myPost.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	}

}
