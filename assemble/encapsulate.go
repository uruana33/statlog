package assemble

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"statUpload/zkconfig"
	"strings"
	"time"
)

//需要打包成以下这种json格式
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
	switch data.(type) {
	case string:
		forMatStr = fmt.Sprintf("%s", data)
	case int64:
		forMatStr = fmt.Sprintf("%d", data)
	case float64:
		forMatStr = fmt.Sprintf("%.3f", data)
	default:
		forMatStr = fmt.Sprintf("%v", data)
	}
	return
}

//将数据封装成json格式
func format2JSON() (jsonData []byte) {

	serviceName := zkconfig.ServiceConfMap[BizKey]["srvName"]
	serviceMetric := zkconfig.ServiceConfMap[BizKey]["srvMetric"]

	tPreMin := time.Now().Unix() - 60

	var pDL jsonSlice

	for myCmd, myData := range myUpload {
		var appid string
		var cmdword string
		if "ALL" == myCmd {
			var eachjson jsonStruct
			//eachjson.Endpoint, _ = os.Hostname()
			eachjson.Endpoint = "game.mstore.test"
			eachjson.Timestamp = tPreMin
			eachjson.Step = 60
			eachjson.CounterType = "GAUGE"
			//命令字ALL
			eachjson.Metric = "QPS"
			eachjson.TAGS = fmt.Sprintf("service=%s", serviceName)
			eachjson.Value = (*myData).REQTotal
			pDL.postDataList = append(pDL.postDataList, eachjson)
			continue
		} else {
			//普通命令字
			//不是以下划线分隔的命令字格式
			if 1 > strings.Count(myCmd, "_") {
				continue
			}

			aList := strings.SplitN(myCmd, "_", 2)
			if 2 == len(aList) {
				reg := regexp.MustCompile(`[\d]+`)
				if reg.MatchString(aList[0]) {
					appid = aList[0]
					cmdword = aList[1]
				} else {
					appid = "null"
					cmdword = myCmd
				}
			} else {
				appid = "null"
				cmdword = myCmd
			}
		}

		//无法对struct进行for-range操作,用reflect来实现
		refDataType := reflect.TypeOf(*myData)
		refDataValue := reflect.ValueOf(*myData)
		for i := 0; i < refDataType.NumField(); i++ {
			//对返回码明细进行特殊处理,Metric不一样
			if "RetCodeDetail" == refDataType.Field(i).Name {
				for statCode, statValue := range (*myData).RetCodeDetail {
					var eachjson jsonStruct
					eachjson.Metric = "StatusCode"
					//eachjson.Endpoint, _ = os.Hostname()
					eachjson.Endpoint = "game.mstore.test"
					eachjson.Timestamp = tPreMin
					eachjson.Step = 60
					eachjson.CounterType = "GAUGE"
					if "null" == appid {
						eachjson.TAGS = fmt.Sprintf("cmdword=%s,retcode=%d,service=%s", cmdword, statCode, serviceName)
					} else {
						eachjson.TAGS = fmt.Sprintf("appid=%s,cmdword=%s,retcode=%d,service=%s", appid, cmdword, statCode, serviceName)
					}
					//eachjson.Value = fmt.Sprintf("%.3f", statValue)
					eachjson.Value = getTypeFormat(statValue)
					pDL.postDataList = append(pDL.postDataList, eachjson)
				}
				continue
			}

			var eachjson jsonStruct
			eachjson.Metric = serviceMetric
			//eachjson.Endpoint, _ = os.Hostname()
			eachjson.Endpoint = "game.mstore.test"
			eachjson.Timestamp = tPreMin
			eachjson.Step = 60
			eachjson.CounterType = "GAUGE"
			//eachjson.Value = fmt.Sprintf("%v", refDataValue.Field(i))
			eachjson.Value = getTypeFormat(refDataValue.Field(i))

			if "null" == appid {
				eachjson.TAGS = fmt.Sprintf("cmdword=%s,item=%s,service=%s", cmdword, refDataType.Field(i).Name, serviceName)
			} else {
				eachjson.TAGS = fmt.Sprintf("appid=%s,cmdword=%s,item=%s,service=%s", appid, cmdword, refDataType.Field(i).Name, serviceName)
			}

			pDL.postDataList = append(pDL.postDataList, eachjson)
		}
	}

	jsonData, err := json.Marshal(pDL.postDataList)
	if err != nil {
		panic(err)
	}
	return
}
