package isFile

import "os"

func IsFileExist(fileName string) bool {

	if len(fileName) < 1 {
		return false
	}
	_, err := os.Stat(fileName)
	return err == nil || os.IsExist(err)
}

func IsDirExist(dirName string) bool {

	if len(dirName) < 1 {
		return false
	}

	dir, err := os.Stat(dirName)
	if err != nil {
		//不存在返回false
		return os.IsExist(err)
	} else {
		//判断是否为目录
		return dir.IsDir()
	}
}
