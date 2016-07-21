package main

import (
	"fmt"

	"github.com/echou/toml"
)

type ignstruct struct {
	cmdName     string
	ignCodeList []int64
}

var whiteList []ignstruct

var tomlConfigMaps = make(map[string]map[string][]int64)

func parseTomlArgToIntArr(cfgFile string) {

	_, err := toml.DecodeFile(cfgFile, tomlConfigMaps)

	if err != nil {
		fmt.Println(err)
		return
	}

	for cmd, ignmap := range tomlConfigMaps {
		var ignTemp ignstruct
		for _, value := range ignmap {
			ignTemp.cmdName = cmd
			ignTemp.ignCodeList = value
		}
		whiteList = append(whiteList, ignTemp)
	}
	fmt.Printf("%+v\n", whiteList)

}

func main() {
	fileName := "./igncode.toml"
	parseTomlArgToIntArr(fileName)
}
