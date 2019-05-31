package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	/* 启动日志服务 */
	execpath, _ := os.Executable()
	sysPath := filepath.Dir(execpath)
	path := filepath.Join(sysPath, "zlog.txt")
	debugLog := newLog(path,"[run]")
	//zLog("test zLog")

	/* 获取系统配置信息 */
	cfg, err := loadConfig("")
	if err != nil {
		fmt.Println(err)
	}

	/* 连接数据库 */
	dbo, err := ConnToDB(cfg)
	if err != nil {
		debugLog.Fatal("Database connection error")
		return
	}
	defer dbo.Close()

	/* 监听HTTP端口 */
	ListenHttpService(cfg, dbo)
}
