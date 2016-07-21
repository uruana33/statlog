package assemble

var BizKey string

func PackageData(buffCont string, bizKey string) (jsonData []byte) {
	BizKey = bizKey
	bulking(buffCont)
	writeRawLog()
	trimming()
	jsonData = format2JSON()
	clearBuffer()
	return
}
