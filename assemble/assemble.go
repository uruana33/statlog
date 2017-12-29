package assemble

func PackageData(buffCont []byte, kind string, biz string, falconData chan<- FalconStruct, done chan<- struct{}) {

	records := bulking(biz, buffCont)
	//由于encapTypeNormal中会对map进行write操作,writeRawLog则是读操作
	//会报错:[fatal error: concurrent map read and map write],目前不加锁,暂时不采用gorountie
	//go writeRawLog(records, kind)
	//writeRawLog(records, kind, biz)

	switch kind {
	case "TypeNormal":
		//对传过来的map更新
		logHisRetCodeAccess(records, biz)
		//通过该chan,将records中的数据传到falconFormat中
		statNormalBlock := make(chan normalStatPiece)
		go func() {
			//通过falconData,将封装好的数据传给post,formatDone是个数据封装结束标志
			falconFormat(statNormalBlock, biz, falconData, done)
		}()
		encapTypeNormal(records, biz, statNormalBlock)
		close(statNormalBlock)

	case "TypeSpecial":
		statSpecialBlock := make(chan specialStatPiece)
		go func() {
			falconFormat(statSpecialBlock, biz, falconData, done)
		}()
		encapTypeSpecial(records, biz, statSpecialBlock)
		close(statSpecialBlock)
	default:
	}


	//手动回收空间
	//type statInfo map[string]map[int64]*retCont
	for appidcmd, contMAP := range *records {
		for retcode, _ := range contMAP {
			delete(contMAP, retcode)
		}
		delete(*records, appidcmd)
	}


	return
}
