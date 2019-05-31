package main

import (
	"fmt"
	"log"
	"os"
)

var debugLog *log.Logger

func newLog(fileName string,prefix string) *log.Logger {
	/*if runtime.GOOS == "window"{
		logFile,err := os.Open(fileName)
		if err != nil{
			log.Fatalln("open file error")
		}
		l := log.New(logFile,prefix, log.Llongfile)
		return l
	}else {
		l,_ := syslog.NewLogger(syslog.LOG_NOTICE, 0)
		return l
	}*/
	logFile,err := os.Open(fileName)
	if err != nil{
		log.Fatalln("open file error")
	}
	l := log.New(logFile,prefix, log.Llongfile)
	return l
}

//全局日志函数
func zLog(args ...interface{})  {
	sAtt := fmt.Sprint(args...)
	debugLog.Println(sAtt)
}
