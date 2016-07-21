package main

import (
	"statUpload/cron"
	"statUpload/zkconfig"
	"time"
)

func main() {
	xboxJobGrpName := "zhibo-mstore-access"
	//xboxJobGrpName := "zhibo-access"
	go zkconfig.GetZKConfig(xboxJobGrpName)
	time.Sleep(time.Second * 3)
	cron.StartCronJobs()

	/*
		fmt.Printf("%+v\n", zkconfig.WhiteListMap)
		fmt.Printf("%+v\n", zkconfig.ServiceConfMap)
		fmt.Printf("%+v\n", zkconfig.InitConfMap)
	*/

	select {}
}
