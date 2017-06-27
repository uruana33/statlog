package zkconfig

import (
	"log"
	"statUpload/parseConfig"
	"strconv"
	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

var WhiteListMap map[string]map[string][]int64
var ServiceConfMap map[string]map[string]string

func connect() *zk.Conn {

	//测试环境ZK地址
	//var zkhosts string = "zk1.staging.srv:2181,zk2.staging.srv:2181,zk3.staging.srv:2181,zk4.staging.srv:2181"
	//线上环境ZK地址
	var zkhosts string = "10.136.5.11:11000,10.136.4.11:11000,10.136.4.34:11000,10.136.4.54:11000,10.136.4.50:11000"
	var servers []string = strings.Split(zkhosts, ",")
	conn, _, err := zk.Connect(servers, time.Second*3)
	if err != nil {
		panic("Can not Connect to ZK Server !!!")
	}
	return conn
}

func mirror(conn *zk.Conn, path string) (chan []byte, chan error) {
	snapshots := make(chan []byte)
	errors := make(chan error)
	go func() {
		for {
			snapshot, _, events, err := conn.GetW(path)
			if err != nil {
				errors <- err
				return
			}
			snapshots <- snapshot
			evt := <-events
			if evt.Err != nil {
				errors <- evt.Err
				return
			}
		}
	}()
	return snapshots, errors
}

func GetZKConfig() {

	zkWhilteListURL := parseConfig.StatConfig["zkWhiteList"]
	zkServiceURL := parseConfig.StatConfig["zkServices"]

	conn1 := connect()
	defer conn1.Close()

	conn2 := connect()
	defer conn2.Close()

	snapshots1, errs1 := mirror(conn1, zkWhilteListURL)
	snapshots2, errs2 := mirror(conn2, zkServiceURL)

	for {
		select {
		case snap1 := <-snapshots1:
			WhiteListMap = getWhiteListMap(snap1)
		case err1 := <-errs1:
			panic(err1)

		case snap2 := <-snapshots2:
			ServiceConfMap = getServiceConf(snap2)
		case err2 := <-errs2:
			panic(err2)
		}
	}
}

func getWhiteListMap(thisWhiteList []byte) (myWhiteListMap map[string]map[string][]int64) {

	myWhiteListMap = make(map[string]map[string][]int64, 0)
	var wStr string

	if 0 == len(thisWhiteList) {
		log.Println("whiteList configuration is NULL on ZK !!!")
		return
	}

	zkLine := strings.Split(string(thisWhiteList), "\r\n")
	for _, zkLineCont := range zkLine {

		if strings.HasPrefix(zkLineCont, "#") {
			continue
		}

		if 2 > len(zkLineCont) {
			continue
		} else {

			if strings.Contains(zkLineCont, "[") && strings.Contains(zkLineCont, "]") {

				wStr = strings.TrimSpace(strings.Trim(strings.Trim(zkLineCont, "["), "]"))

				if 0 == len(wStr) {
					log.Println("Configuration is Error on ZK miss service Name!!!")
					return
				}

				if _, ok := myWhiteListMap[wStr]; !ok {
					myWhiteListMap[wStr] = make(map[string][]int64)
				}
				continue
			}

			if !strings.Contains(zkLineCont, "=") {
				continue
			}
		}

		if myWhiteListMap[wStr] == nil {
			log.Println("Configuration is Error on ZK miss service Name!!!")
			return
		}

		aList := strings.SplitN(zkLineCont, "=", 2)
		if 2 != len(aList) {
			log.Println("Configuration is Error on ZK with option service !!!")
			continue
		}
		k := aList[0]
		v := aList[1]

		option := strings.TrimSpace(k)
		if _, ok := myWhiteListMap[wStr][option]; !ok {
			myWhiteListMap[wStr][option] = make([]int64, 0)
		}

		for _, aa := range strings.Split(v, ",") {
			//展开白名单范围
			if strings.Contains(aa, "@") {
				bList := strings.SplitN(strings.TrimSpace(aa), "@", 2)
				if 2 != len(bList) {
					log.Println("whiteList code configuration's format is Error !!!")
					continue
				}
				start, _ := strconv.ParseInt(bList[0], 10, 64)
				length, _ := strconv.ParseInt(bList[1], 10, 64)
				var inc int64
				for inc = 0; inc < length; inc++ {
					myWhiteListMap[wStr][option] = append(myWhiteListMap[wStr][option], start+inc)
				}

			} else {
				aaInt64, _ := strconv.ParseInt(strings.TrimSpace(aa), 10, 64)
				myWhiteListMap[wStr][option] = append(myWhiteListMap[wStr][option], aaInt64)
			}
		}
	}
	return
}

func getServiceConf(thisServiceCont []byte) (myServiceConf map[string]map[string]string) {

	if 0 == len(thisServiceCont) {
		log.Println("Service configuration is NULL on ZK !!!")
		return
	}

	var bizStr string
	myServiceConf = make(map[string]map[string]string)
	srvContLine := strings.Split(string(thisServiceCont), "\r\n")
	for _, srvCont := range srvContLine {

		if strings.HasPrefix(srvCont, "#") {
			continue
		}

		if 2 > len(srvCont) {
			continue
		} else {
			if strings.HasPrefix(srvCont, "[") && strings.HasSuffix(srvCont, "]") {
				bizStr = strings.TrimSpace(strings.Trim(strings.Trim(srvCont, "["), "]"))
				if 0 == len(bizStr) {
					log.Println("Configuration is Error on ZK miss service Name!!!")
					return
				}
				if _, ok := myServiceConf[bizStr]; !ok {
					myServiceConf[bizStr] = make(map[string]string)
				}
				continue
			}
			if !strings.Contains(srvCont, "=") {
				continue
			}
		}

		if myServiceConf[bizStr] == nil {
			log.Println("Configuration is Error on ZK miss service Name!!!")
			return
		}

		aList := strings.SplitN(srvCont, "=", 2)
		k := aList[0]
		v := aList[1]
		option := strings.TrimSpace(k)
		value := strings.TrimSpace(v)
		myServiceConf[bizStr][option] = value
	}
	return
}
