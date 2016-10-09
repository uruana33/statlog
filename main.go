package main

import (
	//_ "net/http/pprof"
	"os"
	"statUpload/cron"
	"statUpload/parseConfig"
	"statUpload/zkconfig"
	"time"
)

func main() {

	argNum := len(os.Args)
	if argNum < 2 {
		panic("Not enough parameters. Please run prog like this :./myprog ./xxx.cfg\n")
	}
	cfgFile := os.Args[1]
	parseConfig.ParseFile(cfgFile)
	time.Sleep(time.Microsecond * 10)
	go zkconfig.GetZKConfig()
	time.Sleep(time.Microsecond * 10)
	cron.StartCronJobs()
	select {}
}
