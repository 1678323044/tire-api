package main

import (
	"fmt"
)

func main() {
	/* 启动日志服务 */
	newLog()
	//defer logFile.Close()

	/* 获取系统配置信息 */
	cfg, err := loadConfig("")
	if err != nil {
		fmt.Println(err)
	}

	/* 连接数据库 */
	dbo, err := ConnToDB(cfg)
	if err != nil {
		fmt.Printf("连接数据库失败,%v\n",err)
		return
	}
	defer dbo.Close()

	/* 监听HTTP端口 */
	ListenHttpService(cfg, dbo)
}
