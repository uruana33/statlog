package assemble

import (
	"fmt"
	"os"
	"reflect"
	"statlog/parseConfig"
	"statlog/zkconfig"
	"strconv"
	"strings"
	"time"
)

//falcon平台json格式
type FalconStruct struct {
	Metric      string
	TAGS        string
	Endpoint    string
	Timestamp   int64
	Step        int64
	Value       string
	CounterType string
}

var ThisHostname string
var ThisCluster string

func getTypeFormat(data interface{}) (forMatStr string) {
	switch tt := data.(type) {
	case int64:
		forMatStr = strconv.FormatInt(tt, 10)
	case float64:
		forMatStr = strconv.FormatFloat(tt, 'f', 3, 64)
	case string:
		forMatStr = tt
	default:
		forMatStr = fmt.Sprintf("%v", tt)
	}
	return
}

//命令字ALL
func formatKeyALL(tPreMin int64, serviceName string, servicePort string, keyALLReq float64) *FalconStruct {
	eachjson := new(FalconStruct)
	eachjson.Endpoint = ThisHostname
	eachjson.Timestamp = tPreMin
	eachjson.Step = 60
	eachjson.CounterType = "GAUGE"
	eachjson.Metric = "QPS"
	eachjson.Value = getTypeFormat(keyALLReq)
	if 0 == len(servicePort) {
		strs := []string{"service=", serviceName, ",cluster=", ThisCluster}
		eachjson.TAGS = strings.Join(strs, "")
	} else {
		strs := []string{"service=", serviceName, ",port=", servicePort, ",cluster=", ThisCluster}
		eachjson.TAGS = strings.Join(strs, "")
	}
	return eachjson
}

//状态码
func formatStatusCode(tPreMin int64, appid string, cmdword string, serviceName string, servicePort string, statCode int64, statValue interface{}) *FalconStruct {
	eachjson := new(FalconStruct)
	eachjson.Metric = "StatusCode"
	eachjson.Endpoint = ThisHostname
	eachjson.Timestamp = tPreMin
	eachjson.Step = 60
	eachjson.CounterType = "GAUGE"
	eachjson.Value = getTypeFormat(statValue)
	if "null" == appid {
		if 0 == len(servicePort) {
			strs := []string{"cmdword=", cmdword, ",retcode=", strconv.FormatInt(statCode, 10), ",service=", serviceName, ",cluster=", ThisCluster}
			eachjson.TAGS = strings.Join(strs, "")
		} else {
			strs := []string{"cmdword=", cmdword, ",retcode=", strconv.FormatInt(statCode, 10), ",service=", serviceName, ",port=", servicePort, ",cluster=", ThisCluster}
			eachjson.TAGS = strings.Join(strs, "")
		}
	} else {
		if 0 == len(servicePort) {
			strs := []string{"appid=", appid, ",cmdword=", cmdword, ",retcode=", strconv.FormatInt(statCode, 10), ",service=", serviceName, ",cluster=", ThisCluster}
			eachjson.TAGS = strings.Join(strs, "")
		} else {
			strs := []string{"appid=", appid, ",cmdword=", cmdword, ",retcode=", strconv.FormatInt(statCode, 10), ",service=", serviceName, ",port=", servicePort, ",cluster=", ThisCluster}
			eachjson.TAGS = strings.Join(strs, "")
		}
	}
	return eachjson
}

//正常命令字
func formatNormalCMD(tPreMin int64, appid string, cmdword string, serviceName string, servicePort string, serviceMetric string, itemName string, itemValue interface{}) *FalconStruct {

	eachjson := new(FalconStruct)
	eachjson.Metric = serviceMetric
	eachjson.Endpoint = ThisHostname
	eachjson.Timestamp = tPreMin
	eachjson.Step = 60
	eachjson.CounterType = "GAUGE"
	eachjson.Value = getTypeFormat(itemValue)
	if "null" == appid {
		if 0 == len(servicePort) {
			strs := []string{"cmdword=", cmdword, ",item=", itemName, ",service=", serviceName, ",cluster=", ThisCluster}
			eachjson.TAGS = strings.Join(strs, "")
		} else {
			strs := []string{"cmdword=", cmdword, ",item=", itemName, ",service=", serviceName, ",port=", servicePort, ",cluster=", ThisCluster}
			eachjson.TAGS = strings.Join(strs, "")
		}
	} else {
		if 0 == len(servicePort) {
			strs := []string{"appid=", appid, ",cmdword=", cmdword, ",item=", itemName, ",service=", serviceName, ",cluster=", ThisCluster}
			eachjson.TAGS = strings.Join(strs, "")
		} else {
			strs := []string{"appid=", appid, ",cmdword=", cmdword, ",item=", itemName, ",service=", serviceName, ",port=", servicePort, ",cluster=", ThisCluster}
			eachjson.TAGS = strings.Join(strs, "")
		}
	}
	return eachjson
}

//将数据封装成json格式
func falconFormat(uploadChan interface{}, biz string, jsonData chan<- FalconStruct, formatDone chan<- struct{}) {

	serviceName := zkconfig.ServiceConfMap[biz]["srvName"]
	serviceMetric := zkconfig.ServiceConfMap[biz]["srvMetric"]
	servicePort := zkconfig.ServiceConfMap[biz]["srvPort"]
	ThisHostname, _ = os.Hostname()
	ThisCluster = parseConfig.StatConfig["cluster"]
	tPreMin := time.Now().Unix() - 60

	switch myChan := uploadChan.(type) {
	case chan normalStatPiece:
		for myUpload := range myChan {
			if "ALL" == myUpload.appidcmd {
				eachjson := formatKeyALL(tPreMin, serviceName, servicePort, myUpload.statCont.REQTotal)
				jsonData <- *eachjson
				continue
			}

			//普通命令字
			appid, cmdword := getCmdName(myUpload.appidcmd)
			refDataType := reflect.TypeOf(myUpload.statCont)
			refDataValue := reflect.ValueOf(myUpload.statCont)
			//遍历结构体
			for i := 0; i < refDataType.NumField(); i++ {
				//对返回码明细进行特殊处理,Metric不一样
				if "RetCodeDetail" == refDataType.Field(i).Name {
					/*
					for statCode, statValue := range myUpload.statCont.RetCodeDetail {
						eachjson := formatStatusCode(tPreMin, appid, cmdword, serviceName, servicePort, statCode, statValue)
						jsonData <- *eachjson
					}
					*/
					continue
				}
				itemName := refDataType.Field(i).Name
				itemValue := refDataValue.Field(i)
				eachjson := formatNormalCMD(tPreMin, appid, cmdword, serviceName, servicePort, serviceMetric, itemName, itemValue)
				jsonData <- *eachjson
			}
		}
	case chan specialStatPiece:
		for myUploadKV := range myChan {
			if "ALL" == myUploadKV.appidcmd {
				continue
			}
			//普通命令字
			appid, cmdword := getCmdName(myUploadKV.appidcmd)
			refDataType := reflect.TypeOf(myUploadKV.statCont)
			refDataValue := reflect.ValueOf(myUploadKV.statCont)
			//遍历结构体
			for i := 0; i < refDataType.NumField(); i++ {
				itemName := refDataType.Field(i).Name
				itemValue := refDataValue.Field(i)
				eachjson := formatNormalCMD(tPreMin, appid, cmdword, serviceName, servicePort, serviceMetric, itemName, itemValue)
				jsonData <- *eachjson
			}
		}
	default:
	}

	formatDone <- struct{}{}
	return
}
