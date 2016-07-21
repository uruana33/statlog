package doReport

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"statUpload/assemble"
	"statUpload/mylog"
	"statUpload/zkconfig"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Report() {

	//第一个参数: 数据库引擎
	//第二个参数:数据库DSN配置 Go中没有统一DSN,都是数据库引擎自己定义的,因此不同引擎可能配置不同
	//这里采用go-mysql-driver:https://github.com/go-sql-driver/mysql
	//完整格式:username:password@protocol(address)/dbname?param=value
	//open之后,并没有与数据库建立实际的连接

	//数据库相关参数
	user := zkconfig.InitConfMap["userName"]
	passWD := zkconfig.InitConfMap["passWord"]
	addr := zkconfig.InitConfMap["sqlServer"]
	dbName := zkconfig.InitConfMap["dbName"]
	//tbName := zkconfig.InitConfMap["tableName"]

	mysqlSW := zkconfig.InitConfMap["mysqlSwitch"]

	if "1" != mysqlSW {
		runLog.MyLog(assemble.BizKey, runLog.INFO, "Data land to  mysql disabled!!", 1)
		return
	}

	connStr := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8", user, passWD, addr, dbName)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		runLog.MyLog(assemble.BizKey, runLog.ERROR, "Connect to mysql failed!!!", 1)
	}
	defer db.Close()

	//与数据库建立实际的连接是通过Ping方法完成
	err = db.Ping()
	if err != nil {
		runLog.MyLog(assemble.BizKey, runLog.ERROR, "Connect to mysql failed!!!", 1)
	}
	//insert操作
	statInsert(db, awkResult())

}

func statInsert(db *sql.DB, res string) {

	//stmt, err := db.Prepare("insert into appid_cmd_info(datestr,hostname,appid,cmdword,reqtotal) values(?,?,?,?,?);")
	/*
		stmt, err := db.Prepare("insert into test(datestr,hostname,appid,cmdword,reqtotal) values(?,?,?,?,?);")
		defer stmt.Close()

		if err != nil {
			log.Println(err)
			return
		}
	*/

	indb, _ := db.Begin()

	//获取当前时间
	timeLayOut := "2006-01-02 15:04"
	tNowMin := time.Now().Unix()
	date := time.Unix(tNowMin, 0).Format(timeLayOut)
	host, _ := os.Hostname()

	resBuff := strings.Split(res, "\n")
	for _, resLine := range resBuff {
		//aList := strings.Split(strings.TrimSpace(resLine), " ")
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
		indb.Exec("insert into test(datestr,hostname,appid,cmdword,reqtotal) values(?,?,?,?,?);", date, host, appid, cmd, total)
		//stmt.Exec(date, host, appid, cmd, total)
	}
	indb.Commit()
}

func awkResult() (resultStr string) {

	//获取前一天的日期
	timeLayOut := "2006-01-02"
	tPreDay := time.Now().AddDate(0, 0, -1).Format(timeLayOut)

	runLogDir := zkconfig.InitConfMap["logBaseDir"]
	rawStatFile := fmt.Sprintf("%s%s/%s", runLogDir, assemble.BizKey, zkconfig.InitConfMap["rawStatLog"])

	cmdStr := fmt.Sprintf("cat %s | grep '%s' | grep -v ALL| awk -F'|' '$2 ~ / [0-9]+_/{a[$2]+=$5}END{for(ii in a)print ii,a[ii]}'", rawStatFile, tPreDay)
	f, err := exec.Command("/bin/sh", "-c", cmdStr).Output()
	if err != nil {
		runLog.MyLog(assemble.BizKey, runLog.ERROR, "Runing shell command AWK failed!!!", 1)
		return
	}
	resultStr = string(f)
	return
}
