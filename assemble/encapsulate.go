package assemble

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"statUpload/goLog"
	"statUpload/parseConfig"
	"statUpload/zkconfig"
	"time"
)

//falcon平台json格式
type jsonStruct struct {
	Metric      string
	TAGS        string
	Endpoint    string
	Timestamp   int64
	Step        int64
	Value       interface{}
	CounterType string
}

type jsonSlice struct {
	postDataList []jsonStruct
}

func getTypeFormat(data interface{}) (forMatStr string) {
	switch tt := data.(type) {
	case string:
		forMatStr = fmt.Sprintf("%s", tt)
	case int64:
		forMatStr = fmt.Sprintf("%d", tt)
	case float64:
		forMatStr = fmt.Sprintf("%.3f", tt)
	default:
		forMatStr = fmt.Sprintf("%v", tt)
	}
	return
}

//命令字ALL
func formatKeyALL(tPreMin int64, serviceName string, servicePort string, keyALLReq float64) *jsonStruct {
	var eachjson jsonStruct
	eachjson.Endpoint, _ = os.Hostname()
	//eachjson.Endpoint = "game.mlacc.test"
	eachjson.Timestamp = tPreMin
	eachjson.Step = 60
	eachjson.CounterType = "GAUGE"
	eachjson.Metric = "QPS"
	if 0 == len(servicePort) {
		eachjson.TAGS = fmt.Sprintf("service=%s,cluster=%s", serviceName, parseConfig.StatConfig["cluster"])
	} else {
		eachjson.TAGS = fmt.Sprintf("service=%s,port=%s,cluster=%s", serviceName, servicePort, parseConfig.StatConfig["cluster"])
	}
	eachjson.Value = getTypeFormat(keyALLReq)
	return &eachjson
}

//状态码
func formatStatusCode(tPreMin int64, appid string, cmdword string, serviceName string, servicePort string, statCode int64, statValue interface{}) *jsonStruct {
	var eachjson jsonStruct
	eachjson.Metric = "StatusCode"
	eachjson.Endpoint, _ = os.Hostname()
	//eachjson.Endpoint = "game.mlacc.test"
	eachjson.Timestamp = tPreMin
	eachjson.Step = 60
	eachjson.CounterType = "GAUGE"
	if "null" == appid {
		if 0 == len(servicePort) {
			eachjson.TAGS = fmt.Sprintf("cmdword=%s,retcode=%d,service=%s,cluster=%s", cmdword, statCode, serviceName, parseConfig.StatConfig["cluster"])
		} else {
			eachjson.TAGS = fmt.Sprintf("cmdword=%s,retcode=%d,service=%s,port=%s,cluster=%s", cmdword, statCode, serviceName, servicePort, parseConfig.StatConfig["cluster"])
		}
	} else {
		if 0 == len(servicePort) {
			eachjson.TAGS = fmt.Sprintf("appid=%s,cmdword=%s,retcode=%d,service=%s,cluster=%s", appid, cmdword, statCode, serviceName, parseConfig.StatConfig["cluster"])
		} else {
			eachjson.TAGS = fmt.Sprintf("appid=%s,cmdword=%s,retcode=%d,service=%s,port=%s,cluster=%s", appid, cmdword, statCode, serviceName, servicePort, parseConfig.StatConfig["cluster"])
		}
	}
	eachjson.Value = getTypeFormat(statValue)
	return &eachjson
}

//正常命令字
func formatNormalCMD(tPreMin int64, appid string, cmdword string, serviceName string, servicePort string, serviceMetric string, itemName string, itemValue interface{}) *jsonStruct {

	var eachjson jsonStruct
	eachjson.Metric = serviceMetric
	eachjson.Endpoint, _ = os.Hostname()
	//eachjson.Endpoint = "game.mlacc.test"
	eachjson.Timestamp = tPreMin
	eachjson.Step = 60
	eachjson.CounterType = "GAUGE"
	eachjson.Value = getTypeFormat(itemValue)
	if "null" == appid {
		if 0 == len(servicePort) {
			eachjson.TAGS = fmt.Sprintf("cmdword=%s,item=%s,service=%s,cluster=%s", cmdword, itemName, serviceName, parseConfig.StatConfig["cluster"])
		} else {
			eachjson.TAGS = fmt.Sprintf("cmdword=%s,item=%s,service=%s,port=%s,cluster=%s", cmdword, itemName, serviceName, servicePort, parseConfig.StatConfig["cluster"])
		}
	} else {
		if 0 == len(servicePort) {
			eachjson.TAGS = fmt.Sprintf("appid=%s,cmdword=%s,item=%s,service=%s,cluster=%s", appid, cmdword, itemName, serviceName, parseConfig.StatConfig["cluster"])
		} else {
			eachjson.TAGS = fmt.Sprintf("appid=%s,cmdword=%s,item=%s,service=%s,port=%s,cluster=%s", appid, cmdword, itemName, serviceName, servicePort, parseConfig.StatConfig["cluster"])
		}
	}
	return &eachjson
}

//将数据封装成json格式
func falconFormat(data interface{}) (jsonData []byte) {

	serviceName := zkconfig.ServiceConfMap[BizKey]["srvName"]
	serviceMetric := zkconfig.ServiceConfMap[BizKey]["srvMetric"]
	servicePort := zkconfig.ServiceConfMap[BizKey]["srvPort"]
	tPreMin := time.Now().Unix() - 60
	var pDL jsonSlice
	jsonData = make([]byte, 0)
	switch myUpload := data.(type) {
	case *uploadNormal:
		for myCmd, myData := range *myUpload {
			if "ALL" == myCmd {
				eachjson := formatKeyALL(tPreMin, serviceName, servicePort, (*myData).REQTotal)
				pDL.postDataList = append(pDL.postDataList, *eachjson)
				continue
			}
			//普通命令字
			appid, cmdword := getCmdName(myCmd)
			refDataType := reflect.TypeOf(*myData)
			refDataValue := reflect.ValueOf(*myData)
			for i := 0; i < refDataType.NumField(); i++ {
				//对返回码明细进行特殊处理,Metric不一样
				if "RetCodeDetail" == refDataType.Field(i).Name {
					for statCode, statValue := range (*myData).RetCodeDetail {
						eachjson := formatStatusCode(tPreMin, appid, cmdword, serviceName, servicePort, statCode, statValue)
						pDL.postDataList = append(pDL.postDataList, *eachjson)
					}
					continue
				}
				itemName := refDataType.Field(i).Name
				itemValue := refDataValue.Field(i)
				eachjson := formatNormalCMD(tPreMin, appid, cmdword, serviceName, servicePort, serviceMetric, itemName, itemValue)
				pDL.postDataList = append(pDL.postDataList, *eachjson)
			}
		}
		data, err := json.Marshal(pDL.postDataList)
		if err != nil {
			msg := fmt.Sprintf("json change faild:%s\n", err)
			goLog.SendLog(msg, "ERROR", BizKey)
			return nil
		}
		jsonData = data

	case *uploadSpecial:
		for myCmd, myData := range *myUpload {
			if "ALL" == myCmd {
				eachjson := formatKeyALL(tPreMin, serviceName, servicePort, (*myData).KVTotal)
				pDL.postDataList = append(pDL.postDataList, *eachjson)
				continue
			}
			//普通命令字
			appid, cmdword := getCmdName(myCmd)
			refDataType := reflect.TypeOf(*myData)
			refDataValue := reflect.ValueOf(*myData)
			for i := 0; i < refDataType.NumField(); i++ {
				itemName := refDataType.Field(i).Name
				itemValue := refDataValue.Field(i)
				eachjson := formatNormalCMD(tPreMin, appid, cmdword, serviceName, servicePort, serviceMetric, itemName, itemValue)
				pDL.postDataList = append(pDL.postDataList, *eachjson)
			}
		}
		data, err := json.Marshal(pDL.postDataList)
		if err != nil {
			msg := fmt.Sprintf("json change faild:%s\n", err)
			goLog.SendLog(msg, "ERROR", BizKey)
			return nil
		}
		jsonData = data
	}
	return jsonData
}
