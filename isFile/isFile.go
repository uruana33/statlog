package isFile

import "os"

func IsFileExist(fileName string) bool {

	if len(fileName) < 1 {
		return false
	}
	_, err := os.Stat(fileName)
	return err == nil || os.IsExist(err)
}
