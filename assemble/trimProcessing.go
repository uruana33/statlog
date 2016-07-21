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
	GT10msRate    float64           `延时大于10ms比例`
	GT100msRate   float64           `延时大于100ms比例`
	LT500msRate   float64           `延时小于500ms比例`
	RetCodeDetail map[int64]float64 `返回码明细`
}

//注意这个需要清空
var myUpload = make(map[string]*uploadRawData)

//进一步微调数据,
//求平均,求成功率,求各个占比,并打包相应格式
func trimming() {

	newestCodeMap := make(map[string]map[int64]*stFlag)
	newestCodeMap = logHisRetCodeAccess()

	serviceName := zkconfig.ServiceConfMap[BizKey]["srvName"]

	for appidcmd, retmap := range eachRecord {

		var cmd string
		ttList := strings.SplitN(appidcmd, "_", 2)
		if 2 == len(ttList) {
			reg := regexp.MustCompile(`[\d]+`)
			if reg.MatchString(ttList[0]) {
				cmd = ttList[1]
			} else {
				cmd = appidcmd
			}

		} else {
			cmd = appidcmd
		}

		tempElem := new(uploadRawData)
		//返回码明细
		tempElem.RetCodeDetail = make(map[int64]float64)

		var allsum float64 = 0.0
		var errsum float64 = 0.0
		var gt10ms float64 = 0.0
		var gt100ms float64 = 0.0
		var gt500ms float64 = 0.0

		for retCode, contStruc := range retmap {
			//返回码明细
			tempElem.RetCodeDetail[retCode] = float64(newestCodeMap[appidcmd][retCode].stacc) / 60.0

			//总访问量
			allsum += contStruc.total

			//延时分布统计
			if 0 == retCode {
				gt10ms += contStruc.distri1 + contStruc.distri2 + contStruc.distri3
				gt100ms += contStruc.distri2 + contStruc.distri3
				gt500ms += contStruc.distri3
				//平均值
				tempElem.AvgTimeDelay = contStruc.avg / contStruc.accNum
				//最大值
				tempElem.MaxTimeDelay = contStruc.max
			}

			//错误码
			if 0 != retCode {
				//从zk读取白名单配置
				//该命令字是否配有白名单
				if _, ok := zkconfig.WhiteListMap[serviceName]; ok {
					if ignList, ok := zkconfig.WhiteListMap[serviceName][cmd]; ok {
						found := false
						for _, value := range ignList {
							//白名单中是否存在该返回码
							if value == retCode {
								found = true
								break
							}
						}
						//有白名单,但是没加该返回码,所以错误码统计累计
						if !found {
							errsum += contStruc.total
						}
						//没有改命令字的白名单,所以错误码统计累计
					} else {
						errsum += contStruc.total
					}
					//没有改服务的白名单,所以错误码统计累计
				} else {
					errsum += contStruc.total
				}
			}
		}

		//针对该中statlog的封装:tempElem结构
		tempElem.REQTotal = allsum / 60.0
		tempElem.ErrorCount = errsum / 60.0
		tempElem.SuccRate = (1 - errsum/allsum) * 100.0
		tempElem.GT10msRate = (gt10ms / allsum) * 100.0
		tempElem.GT100msRate = (gt100ms / allsum) * 100.0
		tempElem.LT500msRate = (1 - gt500ms/allsum) * 100.0

		myUpload[appidcmd] = tempElem
	}
}
