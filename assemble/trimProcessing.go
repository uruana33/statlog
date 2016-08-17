package assemble

import (
	"regexp"
	"statUpload/zkconfig"
	"strings"
)

//最后要上报给平台的数据格式
type uploadRawData struct {
	REQTotal      float64           `总访问量`
	ErrorCount    float64           `返回码不为零的总量`
	SuccRate      float64           `成功率`
	AvgTimeDelay  float64           `平均延时`
	MaxTimeDelay  float64           `最大延时`
	GT100msRate   float64           `延时大于100ms比例`
	GT500msRate   float64           `延时大于500ms比例`
	LT1000msRate  float64           `延时小于1000ms比例`
	RetCodeDetail map[int64]float64 `返回码明细`
}

type uploadRawDataKV struct {
	KVTotal      float64 `该key访问量`
	KVAvgByte    float64 `平均key大小`
	KVMaxByte    float64 `最大key大小`
	KVGT1KBRate  float64 `key大于1KB的比例`
	KVGT5KBRate  float64 `key大于5KB的比例`
	KVLT10KBRate float64 `key小于10KB的比例`
}

type uploadNormal map[string]*uploadRawData
type uploadSpecial map[string]*uploadRawDataKV

func getCmdName(appidcmd string) (appid string, cmdword string) {
	aList := strings.SplitN(appidcmd, "_", 2)
	if 2 == len(aList) {
		reg := regexp.MustCompile(`[\d]+`)
		if reg.MatchString(aList[0]) {
			appid = aList[0]
			cmdword = aList[1]
		} else {
			appid = "null"
			cmdword = appidcmd
		}
	} else {
		appid = "null"
		cmdword = appidcmd
	}
	return
}

func errNum(retCode int64, total float64, appidcmd string) float64 {

	serviceName := zkconfig.ServiceConfMap[BizKey]["srvName"]
	//从zk读取白名单配置
	//该命令字是否配有白名单
	if _, ok := zkconfig.WhiteListMap[serviceName]; ok {
		_, cmdword := getCmdName(appidcmd)
		if ignList, ok := zkconfig.WhiteListMap[serviceName][cmdword]; ok {
			found := false
			for _, value := range ignList {
				//白名单中是否存在该返回码
				if value == retCode {
					found = true
					break
				}
			}
			if !found {
				//有白名单,但是没加该返回码,所以错误码统计累计
				return total
			} else {
				//找到该白名单,忽略改返回码
				return 0.0
			}
		} else {
			//没有该命令字的白名单,所以错误码统计累计
			return total
		}
	} else {
		//没有该服务的白名单,所以错误码统计累计
		return total
	}
}

func encapTypeNormal(records *statInfo) *uploadNormal {

	//对传过来的map更新
	logHisRetCodeAccess(records)

	myUpload := make(uploadNormal)
	for appidcmd, retmap := range *records {
		var allsum float64 = 0.0
		var errsum float64 = 0.0
		var gt100ms float64 = 0.0
		var gt500ms float64 = 0.0
		var gt1000ms float64 = 0.0
		//返回码明细
		tempElem := new(uploadRawData)
		tempElem.RetCodeDetail = make(map[int64]float64)
		//处理该命令字下的所有返回码
		for retCode, contStruc := range retmap {
			//返回码明细
			tempElem.RetCodeDetail[retCode] = contStruc.total / 60.0
			//总访问量
			allsum += contStruc.total
			//延时分布统计
			if 0 == retCode {
				gt100ms += contStruc.distri1 + contStruc.distri2 + contStruc.distri3
				gt500ms += contStruc.distri2 + contStruc.distri3
				gt1000ms += contStruc.distri3
				//平均值
				tempElem.AvgTimeDelay = contStruc.avg / contStruc.accNum
				//最大值
				tempElem.MaxTimeDelay = contStruc.max
			}
			if 0 != retCode {
				errsum += errNum(retCode, contStruc.total, appidcmd)
			}
		}

		//针对该命令字的封装:tempElem结构
		tempElem.REQTotal = allsum / 60.0
		tempElem.ErrorCount = errsum / 60.0
		tempElem.SuccRate = (1 - errsum/allsum) * 100.0
		tempElem.GT100msRate = (gt100ms / allsum) * 100.0
		tempElem.GT500msRate = (gt500ms / allsum) * 100.0
		tempElem.LT1000msRate = (1 - gt1000ms/allsum) * 100.0
		myUpload[appidcmd] = tempElem
	}
	return &myUpload
}

func encapTypeSpecial(records *statInfo) *uploadSpecial {

	myUploadKV := make(uploadSpecial)
	for appidcmd, retmap := range *records {
		if "ALL" == appidcmd {
			continue
		}
		var allsum float64 = 0.0
		var ge1KB float64 = 0.0
		var ge5KB float64 = 0.0
		var ge10KB float64 = 0.0
		//返回码明细
		tempElem := new(uploadRawDataKV)
		//处理该命令字下的所有返回码
		for retCode, contStruc := range retmap {
			//总访问量
			allsum += contStruc.total
			//Key大小分布统计
			if 0 == retCode {
				ge1KB += contStruc.distri1 + contStruc.distri2 + contStruc.distri3
				ge5KB += contStruc.distri2 + contStruc.distri3
				ge10KB += contStruc.distri3
				//平均值
				tempElem.KVAvgByte = contStruc.avg / contStruc.accNum
				//最大值
				tempElem.KVMaxByte = contStruc.max
			}
		}
		//针对该命令字的封装:tempElem结构
		tempElem.KVTotal = allsum / 60.0
		tempElem.KVGT1KBRate = (ge1KB / allsum) * 100.0
		tempElem.KVGT5KBRate = (ge5KB / allsum) * 100.0
		tempElem.KVLT10KBRate = (1 - ge10KB/allsum) * 100.0
		myUploadKV[appidcmd] = tempElem
	}
	return &myUploadKV
}
