package main

import (
	"os"
	"statUpload/cron"
	"statUpload/parseConfig"
	"statUpload/zkconfig"
	"time"
)

func main() {

	cfgFile := os.Args[1]
	//cfgFile := "./localProg.cfg"
	parseConfig.ParseFile(cfgFile)
	time.Sleep(time.Microsecond * 10)
	go zkconfig.GetZKConfig()
	time.Sleep(time.Microsecond * 10)
	cron.StartCronJobs()
	/*
		time.Sleep(time.Second * 1)
		fmt.Printf("%+v\n", zkconfig.WhiteListMap)
		fmt.Printf("%+v\n", zkconfig.ServiceConfMap)
	*/
	select {}
}
