/* 数据库操作 */
package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
)

/* 定义结构体 保存数据库信息 */
type DBInfo struct {
	session *mgo.Session //数据库操作会话
	dbName  string       //数据库名
}

/* 设置集合常量 */
const (
	CollectionCounters   string = "counters"
	CollectionLogin      string = "login"
	CollectionTeams      string = "teams"
	CollectionCompanies  string = "companies"
	CollectionReceivers  string = "receivers"
	CollectionLastStatus string = "laststatuses"
	collectionRawDatas   string = "rawdatas"
)

func ConnToDB(cfg *Config) (*DBInfo, error) {
	session, err := mgo.Dial(cfg.DBUrl)
	if err != nil {
		debugLog.Panic("Failure to connect database")
	}
	//设置连接缓冲池的最大值
	session.SetPoolLimit(100)
	//给结构体初始化赋值 保存会话对象和数据库名
	dbo := DBInfo{session, cfg.DBName}
	return &dbo, nil
}

/* 关闭会话 */
func (p *DBInfo) Close() {
	p.session.Close()
}

/* ID自增 原子操作 */
func (p *DBInfo) getNextId(sequenceName string) int {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionCounters)
	IdInt := struct {
		Value int
	}{}
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"value": 1}},
		Upsert:    true,
		ReturnNew: true,
	}
	//findAndModify命令是原子性的，Apply可以实现findAndModify功能
	c.Find(bson.M{"_id": sequenceName}).Apply(change, &IdInt)
	return IdInt.Value
}

/* 查看用户名密码 返回bool类型 */
func (p *DBInfo) findLoginCount(username, password string) int {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionLogin)
	count,err := c.Find(bson.M{"username": username, "password": password}).Count()
	if err != nil {
		debugLog.Println("find user name error,error reason:",err)
	}
	return count
}

/* 查看公司车队 聚合查询 */
func (p *DBInfo) findCompAndTeams() ([]CompAndTeam, error) {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionCompanies)
	var CompAndTeams []CompAndTeam
	pipe := c.Pipe([]bson.M{{"$lookup": bson.M{"from": "teams", "localField": "cid", "foreignField": "cid", "as": "teams"}}})
	err := pipe.All(&CompAndTeams)
	if err != nil {
		debugLog.Printf("find companies and teams error on sidebar,error reason:%v\n", err)
		return CompAndTeams, err
	}
	return CompAndTeams, err
}


/* 查看公司 */
func (p *DBInfo) findCompanies() ([]Company, error) {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionCompanies)
	var companies []Company
	err := c.Find(nil).All(&companies)
	if err != nil {
		debugLog.Printf("find companies error,error reason:%v\n", err)
		return nil, err
	}
	return companies, err
}

/* 查看公司 限定字段cid=? */
func (p *DBInfo) findCompanyCid(iCid int) (Company, error) {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionCompanies)
	var company Company
	err := c.Find(bson.M{"cid": iCid}).One(&company)
	if err != nil {
		debugLog.Printf("find company error by cid,error reason:%v\n", err)
		return company, err
	}
	return company, err
}

/* 查看公司 过滤字段cid,name */
func (p *DBInfo) findCompaniesCidName() ([]Company, error) {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionCompanies)
	var companies []Company
	err := c.Find(nil).Select(bson.M{"cid": 1, "name": 1}).All(&companies)
	if err != nil {
		debugLog.Printf("select cid and name find companies error,Error reason:%v\n", err)
		return nil, err
	}
	return companies, err
}

/* 查看公司 返回bool */
func (p *DBInfo) findCompanyCount(iCid int) bool {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionCompanies)
	var company Company
	err := c.Find(bson.M{"cid": iCid}).One(&company)
	if err != nil {
		debugLog.Printf("find company error by cid,Error reason:%v\n", err)
		return false
	}
	return true
}

/* 添加公司 */
func (p *DBInfo) insertCompany(company *Company) error {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionCompanies)
	err := c.Insert(company)
	if err != nil {
		debugLog.Printf("addition company error,error reason:%v\n", err)
		return err
	}
	return err
}

/* 删除公司 限定字段cid=? */
func (p *DBInfo) delCompany(iCid int) error {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionCompanies)
	err := c.Remove(bson.M{"cid": iCid})
	if err != nil {
		debugLog.Printf("deletion company error by cid,error reason:%v\n", err)
		return err
	}
	return err
}

/* 修改公司 限定字段cid */
func (p *DBInfo) updateACompany(iCid int, name, phone, email, manager string) error {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionCompanies)
	err := c.Update(bson.M{"cid": iCid}, bson.M{"$set": bson.M{"name": name, "phone": phone, "email": email, "manager": manager}})
	if err != nil {
		debugLog.Printf("update company error by cid,error reason:%v\n", err)
		return err
	}
	return err
}

/* 查看车队 限定字段cid=? */
func (p *DBInfo) findTeamsCid(iCid int) ([]Team, error) {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionTeams)
	var teams []Team
	err := c.Find(bson.M{"cid": iCid}).All(&teams)
	if err != nil {
		debugLog.Printf("find teams error by cid,error reason:%v\n", err)
		return nil, err
	}
	return teams, err
}

/* 查看车队 限定字段tid=? */
func (p *DBInfo) findTeamTid(oid bson.ObjectId) (Team, error) {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionTeams)
	var team Team
	err := c.FindId(oid).One(&team)
	if err != nil {
		debugLog.Printf("find teams eror by objectid,error reason:%v\n", err)
		return team, err
	}
	return team, err
}

/* 添加车队 */
func (p *DBInfo) insertTeam(team *Team) error {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionTeams)
	err := c.Insert(team)
	if err != nil {
		debugLog.Printf("insert team error,error reason:%v\n", err)
		return err
	}
	return err
}

/* 删除车队 限定字段tid=? */
func (p *DBInfo) delATeam(iTid bson.ObjectId) error {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionTeams)
	err := c.RemoveId(iTid)
	if err != nil {
		debugLog.Printf("del team error by objectid,error reason:%v\n", err)
		return err
	}
	return err
}

/* 修改车队 限定字段tid=? */
func (p *DBInfo) editATeam(oid bson.ObjectId ,name, leader string) error {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionTeams)
	err := c.Update(bson.M{"_id": oid}, bson.M{"$set": bson.M{"name": name, "leader": leader}})
	if err != nil {
		debugLog.Printf("update team error by objectid,error reason:%v\n", err)
		return err
	}
	return err
}

/* 查看设备 返回所有记录 */
func (p *DBInfo) findReceivers(start, pageSize int) ([]Receiver, error) {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionReceivers)
	var receivers []Receiver
	err := c.Find(nil).Select(bson.M{"_id":0,"serverip":0,"vid":0,"authcode":0,"registered":0}).Sort("-_id").Skip(start).Limit(pageSize).All(&receivers)
	if err != nil {
		debugLog.Printf("find receivers eror,error reason:%v\n", err)
		return nil, err
	}
	return receivers, err
}

/* 查看设备 返回总记录数 */
func (p *DBInfo) findReceiversCount() (int, error) {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionReceivers)
	count, err := c.Count()
	if err != nil {
		debugLog.Printf("find receivers count error,error reason:%v\n", err)
		return 0, err
	}
	return count, err
}

/* 查看设备 限定字段rid=? */
func (p *DBInfo) findAReceiver(rid string) (Receiver, error) {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionReceivers)
	var receiver Receiver
	err := c.Find(bson.M{"rid": rid}).One(&receiver)
	if err != nil {
		debugLog.Printf("find receiver error by rid,error reason:%v\n", err)
		return receiver, err
	}
	return receiver, err
}

/* 添加设备 */
func (p *DBInfo) insertReceivers(receiver *Receiver) error {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionReceivers)
	err := c.Insert(receiver)
	if err != nil {
		debugLog.Printf("insert receivers error,error reason:%v\n", err)
		return err
	}
	return err
}

/* 删除设备 限定字段bid=? */
func (p *DBInfo) delAReceiver(rid string) error {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionReceivers)
	err := c.Remove(bson.M{"rid": rid})
	if err != nil {
		debugLog.Printf("remove receiver erorr,error reason:%v\n", err)
		return err
	}
	return err
}

/* 修改设备 限定字段rid=? */
func (p *DBInfo) updateAReceiver(rid, rids, cnum string,iType int) error {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionReceivers)
	err := c.Update(bson.M{"rid": rid}, bson.M{"$set": bson.M{"rid": rids,"type": iType, "cnum": cnum}})
	if err != nil {
		debugLog.Printf("update receiver error by rid,error reason:%v\n", err)
		return err
	}
	return err
}

/* 查看设备最新状态 */
func (p *DBInfo) findLastStatus(rid string) (Laststatus, error) {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionLastStatus)
	var laststatus Laststatus
	err := c.Find(bson.M{"rid": rid}).Select(bson.M{"rid": 1, "lp": 1, "_id": 0}).One(&laststatus)
	if err != nil {
		debugLog.Printf("find receiver last status error by rid,error reason:%v\n", err)
		return laststatus, err
	}
	fmt.Fprintf(os.Stdout,"结果：%v\n",laststatus)
	return laststatus, err
}

/* 删除最新状态/轮胎 限定字段rid=? cardid=? sn=? */
func (p *DBInfo) delANewStatu(rid string, iCardId, iSn int) error {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionLastStatus)
	err := c.Update(bson.M{"rid": rid}, bson.M{"$pull": bson.M{"lp": bson.M{"cardid": iCardId, "sn": iSn}}})
	if err != nil {
		debugLog.Printf("del receiver last status error by rid,error reason:%v\n", err)
		return err
	}
	return err
}

/* 查看设备日志 */
func (p *DBInfo) findRawDatasMatch(pageSize, start int, match bson.M) ([]RawData, error) {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(collectionRawDatas)
	var rawdatas []RawData
	err := c.Find(match).Select(bson.M{"_id":0,"rdata":0}).Sort("-_id").Skip(start).Limit(pageSize).All(&rawdatas)
	if err != nil{
		debugLog.Printf("find raw data error,error reason:%v\n",err)
		return nil,err
	}
	return rawdatas,err
}

/* 查看设备日志 返回总记录数 */
func (p *DBInfo) findRawDatasCount(match bson.M) int {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(collectionRawDatas)
	count, err := c.Find(match).Count()
	if err != nil {
		debugLog.Printf("find raw data count error,error reason:%v\n", err)
		return 0
	}
	return count
}
