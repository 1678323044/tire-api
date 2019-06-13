package main

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"time"
)

/* 时间戳时间格式化 */
func timeStampToStr(t int64) string{
	ts:= time.Unix(t,0)
	s := ts.Format("2006-01-02 15:04:05")
	return s
}

/* 检查表单提交是否为空 */
func (p *HttpListener)checkFields(fields ...string) bool {
	for _,val := range fields{
		if val == ""{
			return false
		}
	}
	return true
}

/* 检查访问是否携带令牌 */
func (p *HttpListener) shareCheck(w http.ResponseWriter, r *http.Request) (userid string,result bool) {
	//访问控制允许全部来源 允许跨域
	w.Header().Set("Access-Control-Allow-Origin","*")

	query := r.URL.Query()
	userid = query.Get("userid")
	token := query.Get("accesstoken")
	//检查是否缺少字段
	if userid == "" || token == "" {
		s := p.makeResultStr(1003,"缺少必要字段")
		w.Write([]byte(s))
		result = false
		return
	}
	//检查用户令牌是否存在 IsObjectIdHex() 检查是否为objectHex格式的字符串
	if !bson.IsObjectIdHex(userid) || !p.dbInfo.checkAccessToken(userid,token){
		s := p.makeResultStr(1006,"非法访问")
		w.Write([]byte(s))
		result = false
		return
	}
	result = true
	return
}

/* 返回处理结果 字符串类型 */
func (p *HttpListener) makeResultStr(code int,msg string) string {
	if code == 0{
		return fmt.Sprintf(`{"errcode": 0, %s}`, msg)
	}else {
		return fmt.Sprintf(`{"errcode": %d,"msg": "%s"}`,code,msg)
	}
}