package parseConfig

import (
	"statUpload/isFile"

	"github.com/echou/toml"
)

var StatConfig = make(map[string]string)

func ParseFile(cfgFile string) {

	if !isFile.IsFileExist(cfgFile) {
		panic("file is not exist.\n")
	}

	_, err := toml.DecodeFile(cfgFile, StatConfig)
	if err != nil {
		panic(err.Error())
	}
}
