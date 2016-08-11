package assemble

var BizKey string

func PackageData(buffCont []byte, kind string, biz string) (databyte []byte) {
	BizKey = biz
	records := bulking(buffCont)
	//由于encapTypeNormal中会对map进行write操作,writeRawLog则是读操作
	//会报错:[fatal error: concurrent map read and map write],目前不加锁,暂时不采用gorountie
	//go writeRawLog(records, kind)
	writeRawLog(records, kind)
	switch kind {
	case "TypeNormal":
		needUpload := encapTypeNormal(records)
		databyte = falconFormat(needUpload)
	case "TypeSpecial":
		needUploadKV := encapTypeSpecial(records)
		databyte = falconFormat(needUploadKV)
	default:
		databyte = nil
	}
	return
}
