package parseConfig

import (
	"os"

	"github.com/echou/toml"
)

var StatConfig = make(map[string]string)

func isFileExist(fileName string) bool {

	if len(fileName) < 1 {
		return false
	}
	_, err := os.Stat(fileName)
	return err == nil || os.IsExist(err)
}

func ParseFile(cfgFile string) {

	if !isFileExist(cfgFile) {
		return
	}

	_, err := toml.DecodeFile(cfgFile, StatConfig)
	if err != nil {
		return
	}
}
