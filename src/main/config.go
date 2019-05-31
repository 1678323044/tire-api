/* 读取配置 */
package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

//定义系统配置信息
type Config struct {
	HttpPort int    //端口
	DBUrl    string //数据库地址
	DBName   string //数据库名称
}

//加载配置信息
func loadConfig(file string) (*Config, error) {
	if file == "" {
		file = os.Args[0] + ".json"
	}
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(buf, &cfg)
	//返回值是结构体的指针类型和错误信息
	return &cfg, err
}
