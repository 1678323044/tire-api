/* 处理业务逻辑 */
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type OperateHttpListener struct {
	config  *Config
	dbInfo  *DBInfo
	sysPath string
}

func ListenHttpService(cfg *Config, dbo *DBInfo) {
	/* 启动静态服务 */
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	listenter := OperateHttpListener{}

	execpath, _ := os.Executable()
	syspath := filepath.Dir(execpath)
	listenter.sysPath = syspath

	listenter.dbInfo = dbo
	listenter.config = cfg
	listenter.listen(cfg.HttpPort)
}

/* 将端口号和函数绑定 封装监听功能 */
func (p *OperateHttpListener) listen(port int) {
	sPort := fmt.Sprintf(":%d", port)
	http.HandleFunc("/", p.Router)
	http.ListenAndServe(sPort, nil)
}

/* 加载静态文件路径 */
func (p *OperateHttpListener)StaticPath(file string) string {
	path := filepath.Join(p.sysPath, file)
	return path
}

/* 写入模板的数据 */
type TempData struct {
	Username     string
	Password     string
	Checked      bool
}

/* 返回给前端的公共字段 */
type ReturnPublic struct {
	Id         string `json:"id"`
	Statuscode int    `json:"statuscode"`
	Error      string `json:"error"`
}

/* 通用检查 */
func (p *OperateHttpListener)checkFields(fields ...string) bool {
	for _,val := range fields{
		if val == ""{
			return false
		}
	}
	return true
}

/* 显示用户登录页面 */
func (p *OperateHttpListener) showLogin(w http.ResponseWriter, r *http.Request) {
	path := p.StaticPath("view/login.html")
	t, err := template.ParseFiles(path)
	if err != nil {
		w.Write([]byte("页面发生异常，请稍后..."))
		return
	}
	//获取cookie记住密码
	var Data TempData
	username, err1 := r.Cookie("username")
	passwrod, err2 := r.Cookie("password")
	if err1 == nil && err2 == nil {
		Data.Username = username.Value
		Data.Password = passwrod.Value
		checked := false
		if len(username.Value) != 0 && len(passwrod.Value) != 0 {
			checked = true
		}
		Data.Checked = checked
	}
	t.Execute(w, Data)
}

/* 处理用户登录功能 */
func (p *OperateHttpListener) handleLogin(w http.ResponseWriter, r *http.Request) {
	//合法性验证
	username := strings.TrimSpace(r.FormValue("username"))
	password := strings.TrimSpace(r.FormValue("password"))
	remember := r.FormValue("remember")
	checkResult := p.checkFields(username,password)
	if checkResult == false{
		w.Write([]byte("输入框不能为空"))
		return
	}
	//设置cookie完成登录
	var returnPublic ReturnPublic
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	count := p.dbInfo.findLoginCount(username, password)
	if count == 0 {
		returnPublic.Statuscode = 1001
		returnPublic.Error = "用户名或密码错误"
		buf, err := json.Marshal(returnPublic)
		if err != nil {
			debugLog.Printf("JSON parsing error when user login error,Error reason:%v\n,IP address:%v\n", err,r.RemoteAddr)
			return
		}
		w.Write(buf)
	} else {
		if remember == "on" {
			nameCookie := http.Cookie{
				Name:   "username",
				Value:  username,
				Path:   "/",
				MaxAge: 1 << 30,
			}
			pwdCookie := http.Cookie{
				Name:   "password",
				Value:  password,
				Path:   "/",
				MaxAge: 1 << 30,
			}
			http.SetCookie(w, &nameCookie)
			http.SetCookie(w, &pwdCookie)
		} else {
			nameCookie := http.Cookie{
				Name:   "username",
				Value:  username,
				Path:   "/",
				MaxAge: -1,
			}
			pwdCookie := http.Cookie{
				Name:   "password",
				Value:  password,
				Path:   "/",
				MaxAge: -1,
			}
			http.SetCookie(w, &nameCookie)
			http.SetCookie(w, &pwdCookie)
		}
		returnPublic.Statuscode = 0
		buf, err := json.Marshal(returnPublic)
		if err != nil {
			debugLog.Printf("JSON parsing error when user login success,error reason:%v\n,IP address:%v\n", err,r.RemoteAddr)
			return
		}
		w.Write(buf)
	}
}

/* 侧边栏 */
type CompAndTeam struct {
	Cid   int    `json:"cid"`
	Name  string `json:"name"`
	Teams []Team `json:"teams" bson:"teams"`
}

/* 回传侧边栏数据 */
func (p *OperateHttpListener) returnSidebarData (w http.ResponseWriter, r *http.Request){
	var CompAndTeams []CompAndTeam
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	CompAndTeams, err01 := p.dbInfo.findCompAndTeams()
	if err01 != nil{
		debugLog.Println("Find companies and teams error on sidebar pages,Error reason",err01)
		return
	}
	buf,err02 := json.Marshal(CompAndTeams)
	if err02 != nil{
		debugLog.Println("JSON parsing error on sidebar pages,Error reason:",err02)
		return
	}
	w.Write(buf)
}

/* 公司 */
type Company struct {
	Cid     int      `json:"cid"`
	Name    string   `json:"name"`
	Manager string   `json:"manager"`
	Phone   string   `json:"phone"`
	Email   string   `json:"email"`
	Stat    int      `json:"stat"`
}

/* 显示管理公司页面 */
func (p *OperateHttpListener) showCompanies(w http.ResponseWriter, r *http.Request) {
	path1 := p.StaticPath("view/public/header.html")
	path2 := p.StaticPath("view/companies.html")
	t, err := template.ParseFiles(path1, path2)
	if err != nil {
		w.Write([]byte("页面发生异常，请稍后..."))
		return
	}
	t.ExecuteTemplate(w, "companies", nil)
}

/* 回传公司数据 */
func (p *OperateHttpListener)returnCompaniesData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	companies,_ := p.dbInfo.findCompanies()
	buf,_ := json.Marshal(companies)
	w.Write(buf)
}

/* 显示添加公司页面 */
func (p *OperateHttpListener) showAddCompanies(w http.ResponseWriter, r *http.Request) {
	path1 := p.StaticPath("view/public/header.html")
	path2 := p.StaticPath("view/addcompanies.html")
	t, err := template.ParseFiles(path1, path2)
	if err != nil {
		w.Write([]byte("页面发生异常，请稍后..."))
		return
	}
	t.ExecuteTemplate(w, "addcompanies", nil)
}

/* 处理公司添加功能 */
func (p *OperateHttpListener) handleAddCompies(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	email := strings.TrimSpace(r.FormValue("email"))
	manager := strings.TrimSpace(r.FormValue("manager"))

	checkResult := p.checkFields(name,phone,email,manager)
	if checkResult == false{
		w.Write([]byte("输入框不能为空"))
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	cid := p.dbInfo.getNextId("companiesId")
	company := Company{
		Cid:     cid,
		Name:    r.FormValue("name"),
		Phone:   r.FormValue("phone"),
		Email:   r.FormValue("email"),
		Manager: r.FormValue("manager"),
	}
	err01 := p.dbInfo.insertCompany(&company)
	if err01 != nil {
		debugLog.Println("Addition company error,Error reason:",err01)
		return
	}
	var returnPublic ReturnPublic
	returnPublic.Statuscode = 0
	buf, err02 := json.Marshal(returnPublic)
	if err02 != nil {
		debugLog.Println("JSON parsing error when addition company,Error reason:",err02)
		return
	}
	w.Write(buf)
}

/* 处理公司删除功能 */
func (p *OperateHttpListener) handleDelCompanies(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if len(query) == 0 {
		debugLog.Println("Getting url param error on del company pages,Error reason:")
		return
	}
	cid := query["cid"][0]
	iCid, err := strconv.Atoi(cid)
	if err != nil{
		debugLog.Println("Conversion parameter type error on del company pages,Error reason",err)
		return
	}

	err01 := p.dbInfo.delCompany(iCid)
	if err01 != nil {
		debugLog.Println("deletion company error,Error reason:",err01)
		return
	}
	http.Redirect(w, r, "/companies", http.StatusFound)
}

/* 显示公司编辑页面 */
func (p *OperateHttpListener) showEditCompanies(w http.ResponseWriter, r *http.Request) {
	path1 := p.StaticPath("view/public/header.html")
	path2 := p.StaticPath("view/editcompanies.html")
	t, err := template.ParseFiles(path1, path2)
	if err != nil {
		w.Write([]byte("页面发生异常，请稍后..."))
		return
	}
	t.ExecuteTemplate(w, "editcompanies", nil)
}

/* 回传公司编辑数据 */
func (p *OperateHttpListener) returnEditCompanies(w http.ResponseWriter, r *http.Request){
	query := r.URL.Query()
	if len(query) == 0 {
		debugLog.Println("Getting url param error on editing company pages")
		return
	}
	cid := query["cid"][0]
	iCid, _ := strconv.Atoi(cid)
	company, err02 := p.dbInfo.findCompanyCid(iCid)
	if err02 != nil {
		debugLog.Println("find company error on editing company pages,Error reason:",err02)
		return
	}
	buf,err := json.Marshal(company)
	if err != nil{
		debugLog.Println("JSON parsing error when editing company,Error reason:",err)
		return
	}
	w.Write(buf)
}

/* 处理公司编辑功能 */
func (p *OperateHttpListener) handleEditCompanies(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	phone :=r.FormValue("phone")
	email := r.FormValue("email")
	manager := r.FormValue("manager")

	checkResult := p.checkFields(name,phone,email,manager)
	if checkResult == false {
		w.Write([]byte("输入框不能为空！"))
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	iCid, err := strconv.Atoi(r.FormValue("cid"))
	if err != nil {
		return
	}
	err01 := p.dbInfo.updateACompany(iCid, name, phone, email, manager)
	if err01 != nil {
		debugLog.Println("Update company error,Error reason:",err01)
		return
	}
	var returnPublic ReturnPublic
	returnPublic.Statuscode = 0
	buf, err02 := json.Marshal(returnPublic)
	if err02 != nil {
		debugLog.Println("JSON parsing error on update company pages,Error reason:",err02)
		return
	}
	w.Write(buf)
}

/* 显示车队管理页面 */
func (p *OperateHttpListener) showTeams(w http.ResponseWriter, r *http.Request){
	path1 := p.StaticPath("view/public/header.html")
	path2 := p.StaticPath("view/teams.html")
	t, err01 := template.ParseFiles(path1, path2)
	if err01 != nil {
		w.Write([]byte("页面显示错误"))
		return
	}
	t.ExecuteTemplate(w,"teams",nil)
}

type ReturnTeamsData struct {
	Cid      string  `json:"cid"`
	CompanyD Company `json:"company"`
	Teams    []Team  `json:"teams"`
}

/* 车队 */
type Team struct {
	Oid    bson.ObjectId   `bson:"_id"`
	Id     string          `json:"id"`
	Cid    int		       `json:"cid"`
	Pid    string		   `json:"pid"`
	Name   string          `json:"name"`
	Leader string          `json:"leader"`
}

/* 回传车队管理数据 */
func (p *OperateHttpListener) returnTeamsData(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if len(query) == 0 {
		debugLog.Println("Getting url param error on teams pages")
		return
	}
	cid := query["cid"][0]
	iCid, _ := strconv.Atoi(cid)
	company, err02 := p.dbInfo.findCompanyCid(iCid)
	if err02 != nil {
		debugLog.Println("find company error on teams pages,Error reason:",err02)
		return
	}
	teams, err03 := p.dbInfo.findTeamsCid(iCid)
	if err03 != nil {
		debugLog.Println("find teams error on teams pages,Error reason",err03)
		return
	}
	for i := 0; i < len(teams); i ++ {
		teams[i].Id = teams[i].Oid.Hex()
	}
	var returnTeamsData ReturnTeamsData
	returnTeamsData.Cid = cid
	returnTeamsData.CompanyD = company
	returnTeamsData.Teams = teams
	buf,err := json.Marshal(returnTeamsData)
	if err != nil{
		debugLog.Println("JSON parsing error on teams pages,Error reason:",err)
		return
	}
	w.Write(buf)
}

/* 显示车队添加页面 */
func (p *OperateHttpListener) showAddTeams(w http.ResponseWriter, r *http.Request) {
	path1 := p.StaticPath("view/public/header.html")
	path2 := p.StaticPath("view/addteams.html")
	t, err := template.ParseFiles(path1, path2)
	if err != nil {
		w.Write([]byte("页面发生异常，请稍后..."))
		return
	}
	t.ExecuteTemplate(w, "addteams", nil)
}

type ReturnAddTeam struct {
	ReturnPublic
}

/* 处理车队添加功能 */
func (p *OperateHttpListener) handleAddTeams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	name := r.FormValue("name")
	leader := r.FormValue("leader")
	cid := r.FormValue("cid")

	checkResult := p.checkFields(name,leader,cid)
	if checkResult == false{
		w.Write([]byte("输入框不能为空！"))
		return
	}

	iCid, _ := strconv.Atoi(cid)
	team := Team{
		Oid:    bson.NewObjectId(),
		Cid:    iCid,
		Name:   r.FormValue("name"),
		Leader: r.FormValue("leader"),
	}
	err01 := p.dbInfo.insertTeam(&team)
	if err01 != nil {
		debugLog.Println("addition team error,Error reason:",err01)
		return
	}

	var returnAddTeam ReturnAddTeam
	returnAddTeam.ReturnPublic.Statuscode = 0
	buf, err := json.Marshal(returnAddTeam)
	if err != nil {
		debugLog.Println("JSON parsing error on addition team page,Error reason:",err)
		return
	}
	w.Write(buf)
}

/* 处理车队删除功能 */
func (p *OperateHttpListener) handleDelTeams(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if len(query) == 0 {
		debugLog.Println("Getting url param error on del team pages")
		return
	}
	tid := query["tid"][0]
	iTid := bson.ObjectIdHex(tid)
	cid := query["cid"][0]

	err := p.dbInfo.delATeam(iTid)
	if err != nil {
		debugLog.Println("Deletion team error,Error reason:",err)
		return
	}
	http.Redirect(w, r, "/teams?cid="+cid, http.StatusFound)
}

/* 显示编辑车队页面 */
func (p *OperateHttpListener) showEditTeams(w http.ResponseWriter, r *http.Request) {
	path1 := p.StaticPath("view/public/header.html")
	path2 := p.StaticPath("view/editteams.html")
	t, err01 := template.ParseFiles(path1, path2)
	if err01 != nil {
		w.Write([]byte("页面显示错误"))
		return
	}
	t.ExecuteTemplate(w,"editteams",nil)
}

type ReturnEditTeams struct {
	ReturnPublic
	TeamD Team      `json:"team"`
}

/* 回传编辑车队数据 */
func (p *OperateHttpListener) returnEditTeamsData(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if len(query) == 0 {
		debugLog.Println("Getting param errors on editing team pages")
		return
	}
	tid := query["tid"][0]
	oid := bson.ObjectIdHex(tid)
	team, err := p.dbInfo.findTeamTid(oid)
	if err != nil {
		debugLog.Println("find team error on editing team pages,Error reason:",err)
		return
	}
	var returnEditTeams ReturnEditTeams
	returnEditTeams.TeamD = team
	returnEditTeams.ReturnPublic.Statuscode = 0
	buf,err01 := json.Marshal(returnEditTeams)
	if err01 != nil{
		debugLog.Println("JSON parsing error when editing team,Error reason:",err01)
		return
	}
	w.Write(buf)
}

type HandleEditTeams struct {
	ReturnPublic
	Cid        string  `json:"cid"`
}

/* 处理编辑车队功能 */
func (p *OperateHttpListener) handleEditTeams(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	leader := r.FormValue("leader")
	cid := r.FormValue("cid")
	tid := r.FormValue("tid")

	checkResult := p.checkFields(name,leader,cid,tid)
	if checkResult == false{
		w.Write([]byte("输入框不能为空！"))
		return
	}
	oid := bson.ObjectIdHex(tid)
	err01 := p.dbInfo.editATeam(oid, name, leader)
	if err01 != nil {
		debugLog.Println("Update team error,Error reason:",err01)
		return
	}

	var handleEditTeams HandleEditTeams
	handleEditTeams.Cid = cid
	handleEditTeams.ReturnPublic.Statuscode = 0
	buf, err02 := json.Marshal(handleEditTeams)
	if err02 != nil {
		debugLog.Println("JSON parsing error on update team pages,Error reason:", err02)
	}
	w.Write(buf)
}

/* 接收器 */
type Receiver struct {
	Rid        string  `json:"rid"`
	Bid        int     `json:"bid"`
	Model      string  `json:"model"`
	Ct         int64   `bson:"ct"`
	Ctt		   string  `json:"ct"`
	At         int64   `bson:"at"`
	Att        string  `json:"at"`
	Bt         int64   `bson:"bt"`
	Btt        string  `json:"bt"`
	Cnum       string  `json:"cnum"`
	Binded     bool    `json:"binded"`
	Type       int     `json:"type"`
	Serverip   string  `json:"serverip"`
	Registered bool    `json:"registered"`
	Authcode   string  `json:"authcode"`
	Tma        float64 `json:"tma"`
	Ma         float64 `json:"ma"`
	Dur        int     `json:"dur"`
	Lt         int     `json:"lt"`
	Status     int     `json:"status"`
}

/* 显示设备列表页面 */
func (p *OperateHttpListener) showReceivers(w http.ResponseWriter, r *http.Request) {
	path1 := p.StaticPath("view/public/header.html")
	path2 := p.StaticPath("view/receivers.html")
	t, err := template.ParseFiles(path1, path2)
	if err != nil {
		w.Write([]byte("页面显示错误"))
		return
	}
	t.ExecuteTemplate(w,"receivers",nil)
}

type ReturnReceiversData struct {
	PageIndex int        `json:"pageIndex"`
	Receivers []Receiver `json:"receivers"`
	PageCount int        `json:"pageCount"`
}

/* 回传设备数据 */
func (p *OperateHttpListener) handleReceivers(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	count, err01 := p.dbInfo.findReceiversCount()
	if err01 != nil {
		debugLog.Println("find receivers count error,Error reason:",err01)
		return
	}
	pageSize := 4
	pageCount := int(math.Ceil(float64(count) / float64(pageSize)))
	query := r.URL.Query()
	var pageIndex int
	if query.Get("pageIndex") == ""{
		pageIndex = 1
	}else {
		pageIndex,_ = strconv.Atoi(query.Get("pageIndex"))
	}

	start := pageSize * (pageIndex - 1)
	receivers, err02 := p.dbInfo.findReceivers(start, pageSize)
	if err02 != nil {
		debugLog.Println("find receivers error on receivers pages,Error reason:",err02)
		return
	}
	for i := 0; i < len(receivers); i++ {
		receivers[i].Ctt = TimeStampToStr(receivers[i].Ct)
		receivers[i].Att = TimeStampToStr(receivers[i].At)
		receivers[i].Btt = TimeStampToStr(receivers[i].Bt)
	}

	var returnReceiversData ReturnReceiversData
	returnReceiversData.PageCount = pageCount
	returnReceiversData.PageIndex = pageIndex
	returnReceiversData.Receivers = receivers
	buf,err := json.Marshal(returnReceiversData)
	if err != nil{
		debugLog.Println("JSON parsing error")
		return
	}
	w.Write(buf)
}

/* 显示设备添加页面 */
func (p *OperateHttpListener) showAddReceivers(w http.ResponseWriter, r *http.Request) {
	path1 := p.StaticPath("view/public/header.html")
	path2 := p.StaticPath("view/addreceivers.html")
	t, err := template.ParseFiles(path1, path2)
	if err != nil{
		w.Write([]byte("模板加载错误，请重试"));
		return
	}
	t.ExecuteTemplate(w,"addreceivers",nil)
}

/* 回传设备添加数据 */
func (p *OperateHttpListener) returnAddReceivers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	companies, err := p.dbInfo.findCompaniesCidName()
	if err != nil{
		debugLog.Println("find companies error on add receiver pages,Error reason:",err)
		return
	}
	buf,err01 := json.Marshal(companies)
	if err01 != nil{
		debugLog.Println("JSON parsing error on add receiver page,Error reason:",err01)
		return
	}
	w.Write(buf)
}

type ReturnAddReceivers struct {
	ReturnPublic
}

/* 处理设备添加功能 */
func (p *OperateHttpListener) handleAddReceivers(w http.ResponseWriter, r *http.Request) {
	//合法性验证
	rids := r.FormValue("rids")
	cnum := r.FormValue("cnum")
	cid := r.FormValue("cid")
	rtype := r.FormValue("type")

	checkResult := p.checkFields(rids,cnum,cid,rtype)
	if checkResult == false{
		w.Write([]byte("请填写输入框！"))
		return
	}

	iCid, _ := strconv.Atoi(cid)
	res := p.dbInfo.findCompanyCount(iCid)
	if res == false {
		w.Write([]byte("公司id不存在！"))
		return
	}
	iType, _ := strconv.Atoi(rtype)
	if iType != 1 && iType != 2 && iType != 3 && iType != 4 {
		w.Write([]byte("设备类型不存在！"))
		return
	}

	//实现设备添加功能
	nowTime := time.Now().Unix()
	for i := 0; i < len(rids); i += 12 {
		rid := string([]byte(rids)[i : i+12])
		receiver := Receiver{
			Rid:  rid,
			Cnum: cnum,
			Type: iType,
			Ct:   nowTime,
		}
		err01 := p.dbInfo.insertReceivers(&receiver)
		if err01 != nil {
			debugLog.Printf("Add receivers error,Error reason:%v\n,IP address:%v\n",err01,r.RemoteAddr)
			return
		}
	}
	var returnAddReceivers ReturnAddReceivers
	returnAddReceivers.ReturnPublic.Statuscode = 0
	buf,err01 := json.Marshal(returnAddReceivers)
	if err01 != nil{
		debugLog.Printf("Add receivers JSON parsing error,Error reason:%v\n,IP address:%v",err01,r.RemoteAddr)
		return
	}
	w.Write(buf)
}

/* 处理设备删除功能 */
func (p *OperateHttpListener) handleDelReceivers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if len(query) == 0 {
		debugLog.Println("Error getting URL parameters when deleting receiver")
		return
	}
	rid := query["rid"][0]
	err := p.dbInfo.delAReceiver(rid)
	if err != nil {
		debugLog.Printf("Error deleting receiver,Error reason:%v\n,IP address:%v",err,r.RemoteAddr)
		return
	}
	http.Redirect(w, r, "/receivers", http.StatusFound)
}

/* 显示设备编辑页面 */
func (p *OperateHttpListener) showEditReceivers(w http.ResponseWriter, r *http.Request){
	path1 := p.StaticPath("view/public/header.html")
	path2 := p.StaticPath("view/editreceivers.html")
	t, err := template.ParseFiles(path1, path2)
	if err != nil {
		w.Write([]byte("页面发生异常，请稍后..."))
		return
	}
	t.ExecuteTemplate(w,"editreceivers",nil)
}

type EditReceiversData struct {
	ReceiverD  Receiver  `json:"receiver"`
	Companies  []Company `json:"companies"`
	ReturnPublic
}

/* 回传设备编辑页面数据 */
func (p *OperateHttpListener) returnEditReceiversData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	query := r.URL.Query()
	if len(query) == 0{
		debugLog.Printf("Error getting URL parameters when editing receiver pages,IP address:%v\n",r.RemoteAddr)
		return
	}
	rid := query["rid"][0]
	receiver, err01 := p.dbInfo.findAReceiver(rid)
	if err01 != nil{
		debugLog.Printf("find receiver error when editing receiver pages,Error reason:%v\n,IP address:%v\n",err01,r.RemoteAddr)
		return
	}
	companies, err02 := p.dbInfo.findCompaniesCidName()
	if err02 != nil{
		debugLog.Printf("find companies error when editing receiver pages,Error reason:%v\n,IP address:%v\n",err02,r.RemoteAddr)
		return
	}
	var editReceiversData EditReceiversData
	editReceiversData.ReceiverD = receiver
	editReceiversData.Companies = companies
	buf,err03 := json.Marshal(editReceiversData)
	if err03 != nil{
		debugLog.Println("JSON parsing error when editing receiver pages,Error reason:",err03)
		return
	}
	w.Write(buf)
}

/* 处理设备编辑功能 */
func (p *OperateHttpListener) handleEditReceivers(w http.ResponseWriter, r *http.Request) {
	//合法性验证
	cid := r.FormValue("cid")
	rType := r.FormValue("type")
	cnum := r.FormValue("cnum")
	rids := r.FormValue("rids")

	checkResult := p.checkFields(cid,rType,cnum,rids)
	if checkResult == false{
		w.Write([]byte("输入框不能为空！"))
		return
	}

	rid := r.FormValue("rid")
	iType,_ := strconv.Atoi(rType)
	err := p.dbInfo.updateAReceiver(rid, rids, cnum, iType)
	if err != nil {
		debugLog.Printf("Update receiver error,Error reason:%v\n,IP address:%v\n",err,r.RemoteAddr)
		return
	}
	var editReceiversData EditReceiversData
	editReceiversData.ReturnPublic.Statuscode = 0
	buf,err01 := json.Marshal(editReceiversData)
	if err01 != nil{
		debugLog.Println("JSON parsing error when update receiver,Error reason:",err)
		return
	}
	w.Write(buf)
}

/* 最新胎压数据 */
type Lp struct {
	T      int64   `bson:"t"`
	TT     string  `json:"t"`
	Sid    string  `json:"sid"`
	Cardid int     `json:"cardid"`
	Sn     int     `json:"sn"`
	Vol    float64 `json:"vol"`
	Pres   float64 `json:"pres"`
	Temp   int     `json:"temp"`
}

/* 显示设备最新状态页面 */
func (p *OperateHttpListener) showLastStatuses(w http.ResponseWriter, r *http.Request) {
	path1 := p.StaticPath("view/public/header.html")
	path2 := p.StaticPath("view/laststatuses.html")
	t, err01 := template.ParseFiles(path1, path2)
	if err01 != nil {
		w.Write([]byte("页面发生异常，请稍后..."))
		return
	}
	t.ExecuteTemplate(w, "laststatuses", nil)
}

/* 最新状态 */
type Laststatus struct {
	Rid string `json:"rid"`
	Lps []Lp   `json:"lp" bson:"lp"`
}

/* 回传设备最新状态数据 */
func (p *OperateHttpListener) returnLastStatuses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	query := r.URL.Query()
	if len(query) == 0 {
		debugLog.Println("Getting url param errors on receiver Latest status pages,IP address:",r.RemoteAddr)
		return
	}
	rid := query["rid"][0]
	laststatus, err02 := p.dbInfo.findLastStatus(rid)
	if err02 != nil {
		debugLog.Printf("find receiver latest status error,Error reason:%v\n,IP address:%v\n",err02,r.RemoteAddr)
		return
	}
	for i := 0; i < len(laststatus.Lps); i++ {
		laststatus.Lps[i].TT = FormattingTime(laststatus.Lps[i].T)
	}
	buf,err := json.Marshal(laststatus)
	if err != nil{
		debugLog.Println("JSON parsing error on receiver latest status pages,Error reason:",err)
		return
	}
	w.Write(buf)
}

/* 处理删除最新状态轮位 */
func (p *OperateHttpListener) handleDelLastStatus(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if len(query) == 0 {
		debugLog.Println("Getting url param errors when del receiver latest status data")
		return
	}
	cardId := query["cardid"][0]
	iCardId, _ := strconv.Atoi(cardId)
	sn := query["sn"][0]
	iSn, _ := strconv.Atoi(sn)
	rid := query["rid"][0]

	err := p.dbInfo.delANewStatu(rid, iCardId, iSn)
	if err != nil {
		debugLog.Println("deletion receiver latest status data error")
		return
	}
	http.Redirect(w, r, "/laststatuses?rid="+rid, http.StatusFound)
}

/* 原始数据 */
type RawData struct {
	Id    bson.ObjectId `bson:"_id"`
	Imei  string        `json:"imei"`
	Dt    string        `json:"dt"`
	T     int64
	TT    string        `json:"t"`
	Data  string        `json:"rawdata"`
	Rdata string
	Stat  int           `json:"stat"`
	Err   string
}

/* int64时间格式化 */
func FormattingTime(t int64) string {
	if t == 0{
		return ""
	}
	str := strconv.FormatInt(t,10)
	year := str[0:2]
	month := str[2:4]
	day := str[4:6]
	hour := str[6:8]
	minute := str[8:10]
	second := str[10:12]
	date := year + "-" + month + "-" + day
	time1 := hour+":"+minute+":"+second
	return date+" "+time1
}

/* 时间戳时间格式化 */
func TimeStampToStr(t int64) string{
	ts:= time.Unix(t,0)
	s := ts.Format("2006-01-02 15:04:05")
	return s
}

/* 显示设备原始数据 */
func (p *OperateHttpListener) showRawDatas (w http.ResponseWriter, r *http.Request){
	path1 := p.StaticPath("view/public/header.html")
	path2 := p.StaticPath("view/rawdatas.html")
	t ,err := template.ParseFiles(path1, path2)
	if err != nil{
		w.Write([]byte("页面加载错误"))
		return
	}
	t.ExecuteTemplate(w,"rawdatas",nil)
}

type ReturnRawdata struct {
	Count     int       `json:"count"`
	PageIndex int       `json:"pageIndex"`
	Rawdata   RawData   `json:"rawdata"`
	PageCount int	    `json:"pagecount"`
	Rawdatas  []RawData `json:"rawdatas"`
}

/* 回传设备原始数据 */
func (p *OperateHttpListener) returnRawDatas(w http.ResponseWriter, r *http.Request) {
	//判断是否存在查询条件
	match := bson.M{}
	imei := r.FormValue("imei")
	dt := r.FormValue("dt")
	firstDate := r.FormValue("firstDate")
	lastDate := r.FormValue("lastDate")
	if imei != "" {
		match["imei"] = imei
	}
	if dt != "" {
		match["dt"] = dt
	}
	if firstDate != "" && lastDate != "" {
		match["t"] = bson.M{"$gte": firstDate, "$lte": lastDate}
	}
	rawdata := RawData{
		Imei: imei,
		Dt:   dt,
	}

	//实现翻页功能
	count := p.dbInfo.findRawDatasCount(match)
	pageSize := 4
	pageCount := int(math.Ceil(float64(count) / float64(pageSize)))
	var pageIndex int
	if r.FormValue("pageIndex") == "" {
		pageIndex = 1
	} else {
		pageIndex, _ = strconv.Atoi(r.FormValue("pageIndex"))
	}
	start := pageSize * (pageIndex - 1)
	rawdatas,err01 := p.dbInfo.findRawDatasMatch(pageSize, start, match)
	if err01 != nil{
		debugLog.Println("find raw data error on raw data pages,Error reason:",err01)
		return
	}
	for i := 0; i < len(rawdatas); i++ {
		rawdatas[i].TT = TimeStampToStr(rawdatas[i].T)
	}

	var returnRawdata ReturnRawdata
	returnRawdata.Count = count
	returnRawdata.PageIndex = pageIndex
	returnRawdata.Rawdata = rawdata
	returnRawdata.PageCount = pageCount
	returnRawdata.Rawdatas = rawdatas
	buf,err := json.Marshal(returnRawdata)
	if err != nil{
		debugLog.Println("JSON parsing error on raw data pages,Error reason:",err)
		return
	}
	w.Write(buf)
}

/* 显示远程升级页面 */
func (p *OperateHttpListener)showRemoteUpgrade(w http.ResponseWriter, r *http.Request){
	path1 := p.StaticPath("view/public/header.html")
	path2 := p.StaticPath("view/remoteupgrade.html")
	t,err := template.ParseFiles(path1,path2)
	if err != nil{
		w.Write([]byte("页面加载错误"))
		return
	}
	t.ExecuteTemplate(w,"remoteupgrade",nil)
}

/* 显示远程关闭页面 */
func (p *OperateHttpListener)showRemoteShutdown(w http.ResponseWriter, r *http.Request){
	path1 := p.StaticPath("view/public/header.html")
	path2 := p.StaticPath("view/remoteshutdown.html")
	t,err := template.ParseFiles(path1,path2)
	if err != nil{
		w.Write([]byte("页面加载失败！"))
		return
	}
	t.ExecuteTemplate(w,"remoteshutdown",nil)
}