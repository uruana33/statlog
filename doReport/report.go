package doReport

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"statlog/goLog"
	"statlog/parseConfig"
	"statlog/zkconfig"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Report() {

	for bizkey, cfgMap := range zkconfig.ServiceConfMap {
		mysqlSW := cfgMap["mysqlSwitch"]
		if "1" != mysqlSW {
			msg := "upload to mysql disabled. "
			goLog.SendLog(msg, "INFO", bizkey)
			return
		}
		user := cfgMap["userName"]
		passWD := cfgMap["passWord"]
		addr := cfgMap["sqlServer"]
		dbName := cfgMap["dbName"]
		//tbName := cfgMap["tableName"]
		connStr := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8", user, passWD, addr, dbName)
		db, err := sql.Open("mysql", connStr)
		defer db.Close()
		if err != nil {
			goLog.SendLog(err.Error(), "ERROR", bizkey)
		}
		//与数据库建立实际的连接是通过Ping方法完成
		err = db.Ping()
		if err != nil {
			goLog.SendLog(err.Error(), "ERROR", bizkey)
		}
		//insert操作
		statInsert(db, awkResult(bizkey))
	}
}

func statInsert(db *sql.DB, res string) {

	indb, _ := db.Begin()
	//获取当前时间
	timeLayOut := "2006-01-02 15:04"
	tNowMin := time.Now().Unix()
	date := time.Unix(tNowMin, 0).Format(timeLayOut)
	host, _ := os.Hostname()

	resBuff := strings.Split(res, "\n")
	for _, resLine := range resBuff {
		aList := strings.Fields(strings.TrimSpace(resLine))
		if 2 != len(aList) {
			continue
		}
		appidCMD := aList[0]
		total := aList[1]

		bList := strings.SplitN(appidCMD, "_", 2)
		if 2 != len(bList) {
			continue
		}
		appid := bList[0]
		cmd := bList[1]
		indb.Exec("insert into appid_cmd_info(datestr,hostname,appid,cmdword,reqtotal) values(?,?,?,?,?);", date, host, appid, cmd, total)
	}
	indb.Commit()
}

func awkResult(bizkey string) (resultStr string) {
	//获取前一天的日期
	timeLayOut := "2006-01-02"
	tPreDay := time.Now().AddDate(0, 0, -1).Format(timeLayOut)
	runLogDir := parseConfig.StatConfig["logBaseDir"]
	rawStatFile := fmt.Sprintf("%s%s/%s_%s", runLogDir, bizkey, bizkey, parseConfig.StatConfig["rawStatLog"])
	cmdStr := fmt.Sprintf("cat %s | grep '%s' | grep -v ALL| awk -F'|' '$2 ~ / [0-9]+_/{a[$2]+=$5}END{for(ii in a)print ii,a[ii]}'", rawStatFile, tPreDay)
	f, err := exec.Command("/bin/sh", "-c", cmdStr).Output()
	if err != nil {
		msg := "Runing shell command AWK failed!!!"
		goLog.SendLog(msg, "INFO", bizkey)
		return
	}
	resultStr = string(f)
	return
}
