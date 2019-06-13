package main

import (
	"log"
	"os"
	"path/filepath"
)

var debugLog *log.Logger

func newLog() *os.File{
	execPath, _ := os.Executable()  //返回.exe文件的路径
	sysPath := filepath.Dir(execPath)  //返回上一级目录
	logPath := filepath.Join(sysPath, "log.txt")
	logFile,_ := os.Create(logPath)
	debugLog = log.New(logFile,"[Info]",log.Llongfile)
	return logFile
}
