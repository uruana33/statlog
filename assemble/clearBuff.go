package assemble

func clearBuffer() {

	for key, _ := range myUpload {
		delete(myUpload, key)
	}

	for cmdKey, retMap := range eachRecord {
		for retKey, _ := range retMap {
			delete(retMap, retKey)
		}
		delete(eachRecord, cmdKey)
	}
}
