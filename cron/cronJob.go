package cron

import (
	"log"
	"statUpload/doPost"
	"statUpload/doReport"

	"github.com/robfig/cron"
)

type cronJob struct {
	Comment string
	Spec    string
	fun     func()
}

var (
	kCron     = cron.New()
	kCronJobs []*cronJob
)

func addJob(c *cronJob) {
	kCronJobs = append(kCronJobs, c)
}

func (c *cronJob) Run() {
	log.Println(c.Comment)
	c.fun()
}

func StartCronJobs() {

	addJob(&cronJob{
		" stat数据上报(60s)... ",
		"0 * * * * *",
		doPost.Send,
	})

	addJob(&cronJob{
		" 日报数据入库 ",
		"0 0 3 * * *",
		doReport.Report,
	})

	for _, j := range kCronJobs {
		kCron.AddJob(j.Spec, j)
	}
	kCron.Start()
}
