/* 处理业务逻辑 */
package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"math"
	"net/http"
	"strconv"
)

type HttpListener struct {
	config  *Config
	dbInfo  *DBInfo
	sysPath string
}

func ListenHttpService(cfg *Config, dbo *DBInfo) {
	listenter := HttpListener{}
	listenter.dbInfo = dbo
	listenter.config = cfg
	listenter.listen(cfg.HttpPort)
}

/* 将端口号和函数绑定 封装监听功能 */
func (p *HttpListener) listen(port int) {
	sPort := fmt.Sprintf(":%d", port)
	http.HandleFunc("/", p.Router)
	http.ListenAndServe(sPort, nil)
}

/* 公司信息 */
type Company struct {
	Cid     int      `json:"cid"`
	Name    string   `json:"name"`
	Manager string   `json:"manager"`
	Phone   string   `json:"phone"`
	Email   string   `json:"email"`
}

//返回公司信息
type ReturnCompanies struct {
	Errcode    int     `json:"errcode"`
	Companies  []Company `json:"companies"`
}

/* 处理公司列表 */
func (p *HttpListener)handleCompanies(w http.ResponseWriter, r *http.Request) {
	if _,ok := p.shareCheck(w,r); ok{
		companies,err := p.dbInfo.findCompanies()
		if err != nil {
			sErr := p.makeResultStr(2000,"查询公司列表错误")
			w.Write([]byte(sErr))
			return
		}
		returnCompanies := ReturnCompanies{
			Errcode: 0,
			Companies: companies,
		}
		buf,err01 := json.Marshal(returnCompanies)
		if err01 != nil {
			debugLog.Printf("公司列表的json解析错误,err:%v\n",err01)
			return
		}
		w.Write(buf)
	}
}

/* 处理添加公司功能 */
func (p *HttpListener) handleAddCompany(w http.ResponseWriter, r *http.Request) {
	if _,ok := p.shareCheck(w,r); ok{
		name := r.FormValue("name")
		phone := r.FormValue("phone")
		email := r.FormValue("email")
		manager := r.FormValue("manager")

		checkResult := p.checkFields(name, phone, email, manager)
		if !checkResult {
			sErr := p.makeResultStr(1003,"输入框不能为空")
			w.Write([]byte(sErr))
			return
		}
		cid := p.dbInfo.getNextId("companiesId")
		company := Company{
			Cid:     cid,
			Name:    name,
			Phone:   phone,
			Email:   email,
			Manager: manager,
		}
		err01 := p.dbInfo.insertCompany(&company)
		if err01 != nil {
			sErr := p.makeResultStr(1101,"添加公司失败")
			w.Write([]byte(sErr))
			return
		}
		var returnCompanies ReturnCompanies
		returnCompanies.Errcode = 0
		buf, err02 := json.Marshal(returnCompanies)
		if err02 != nil {
			debugLog.Printf("添加公司解析json失败,err:%v\n",err02)
			return
		}
		w.Write(buf)
	}
}

/* 处理编辑公司功能 */
func (p *HttpListener) handleEditCompany (w http.ResponseWriter, r *http.Request){
	if _,ok := p.shareCheck(w,r); ok{
		name := r.FormValue("name")
		phone := r.FormValue("phone")
		email := r.FormValue("email")
		manager := r.FormValue("manager")
		cid := r.FormValue("cid")
		iCid,_ := strconv.Atoi(cid)

		checkResult := p.checkFields(name,cid)
		if !checkResult {
			sErr := p.makeResultStr(1003,"缺少必要字段")
			w.Write([]byte(sErr))
			return
		}
		company := Company{
			Cid:     iCid,
			Name:    name,
			Phone:   phone,
			Email:   email,
			Manager: manager,
		}
		err := p.dbInfo.updateCompany(&company)
		if err != nil {
			sErr := p.makeResultStr(1101,"编辑公司失败")
			w.Write([]byte(sErr))
			return
		}
		var returnCompanies ReturnCompanies
		returnCompanies.Errcode = 0
		buf,err01 := json.Marshal(returnCompanies)
		if err01 != nil {
			debugLog.Printf("编辑公司json解析失败,err:%v\n",err01)
			return
		}
		w.Write(buf)
	}
}

/* 处理公司删除功能 */
func (p *HttpListener) handleDelCompanies(w http.ResponseWriter, r *http.Request) {
	if _,ok := p.shareCheck(w, r);ok {
		cid := r.FormValue("cid")
		iCid, _ := strconv.Atoi(cid)

		checkResult := p.checkFields(cid)
		if !checkResult {
			sErr := p.makeResultStr(1003,"缺少必要字段")
			w.Write([]byte(sErr))
			return
		}

		err := p.dbInfo.delCompany(iCid)
		if err != nil {
			sErr := p.makeResultStr(1101,"删除公司失败")
			w.Write([]byte(sErr))
			return
		}
		var returnCompanies ReturnCompanies
		returnCompanies.Errcode = 0
		buf,_ := json.Marshal(returnCompanies)
		w.Write(buf)
	}
}

/* 处理原始数据 */
type RawdataT struct {
	Id   bson.ObjectId `bson:"_id"`
	Imei string	 `json:"imei"`
	Dt   string  `json:"dt"`
	T    int64
	TT   string  `json:"t"`
	Data string  `json:"rawdata"`
	Stat int     `json:"stat"`
}

/* 返回的数据 */
type ReturnRawdata struct {
	Errcode    int         `json:"errcode"`
	Count      int         `json:"count"`
	PageIndex  int	       `json:"pageIndex"`
	PageCount  int         `json:"pageCount"`
	Rawdatas   []RawdataT  `json:"rawdatas"`
	Rawdata    RawdataT	   `json:"rawdata"`
}

/* 处理原始数据 */
func (p *HttpListener)handleRawDatas(w http.ResponseWriter, r *http.Request)  {
	if _,ok := p.shareCheck(w,r); ok{
		//判断是否存在查询条件
		match := bson.M{}
		imei := r.FormValue("imei")
		dt := r.FormValue("dt")
		pageIndex := r.FormValue("pageIndex")
		if imei != "" {
			match["imei"] = imei
		}
		if dt != "" {
			match["dt"] = dt
		}
		rawdata := RawdataT{
			Imei: imei,
			Dt:   dt,
		}

		//实现翻页功能
		count := p.dbInfo.findRawDatasCount(match)
		pageSize := 4
		pageCount := int(math.Ceil(float64(count) / float64(pageSize)))
		var iPageIndex int
		if pageIndex == "" {
			iPageIndex = 1
		}else if iPageIndex,_ = strconv.Atoi(pageIndex);iPageIndex > pageCount {
			iPageIndex = pageCount
		}else if iPageIndex,_ = strconv.Atoi(pageIndex);iPageIndex < 1 {
			iPageIndex = 1
		} else{
			iPageIndex,_ = strconv.Atoi(pageIndex)
		}
		start := pageSize * (iPageIndex - 1)
		rawdatas,err := p.dbInfo.findRawDatasMatch(pageSize,start, match)
		if err != nil {
			sErr := p.makeResultStr(2000,"查询原始数据错误")
			w.Write([]byte(sErr))
			return
		}
		for i := 0; i < len(rawdatas); i++ {
			rawdatas[i].TT = timeStampToStr(rawdatas[i].T)
		}
		returnRawdata := ReturnRawdata{
			Errcode: 0,
			Count: count,
			PageIndex: iPageIndex,
			PageCount: pageCount,
			Rawdatas: rawdatas,
			Rawdata: rawdata,
		}
		buf,err := json.Marshal(returnRawdata)
		if err != nil {
			debugLog.Printf("响应的rawdata数据 json解析错误")
			return
		}
		w.Write(buf)
	}
}