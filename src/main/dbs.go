/* 数据库操作 */
package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/* 定义结构体 保存数据库信息 */
type DBInfo struct {
	session *mgo.Session //数据库操作会话
	dbName  string       //数据库名
}

/* 设置集合常量 */
const (
	CollectionCounters   string = "counters"
	CollectionCompanies  string = "companies"
	CollectionCustomers  string = "customers"
)

func ConnToDB(cfg *Config) (*DBInfo, error) {
	session, err := mgo.Dial(cfg.DBUrl)
	if err != nil {

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

/* 检查访问令牌 */
func (p *DBInfo) checkAccessToken(userid,accesstoken string) bool {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionCustomers)
	//useridObId := bson.ObjectIdHex(userid)
	count,err := c.Find(bson.M{"id": userid,"accesstoken": accesstoken}).Count()
	if count == 0 || err != nil {
		return false
	}
	return true
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

/* 查看公司 */
func (p *DBInfo) findCompanies() ([]Company, error) {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionCompanies)
	var companies []Company
	err := c.Find(nil).All(&companies)
	if err != nil {
		fmt.Printf("查询公司信息错误,err:%v\n",err)
		return nil, err
	}
	return companies, err
}

/* 添加公司 */
func (p *DBInfo) insertCompany(company *Company) error {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionCompanies)
	err := c.Insert(company)
	if err != nil {
		fmt.Printf("添加公司失败,err:%v\n",err)
		return err
	}
	return err
}

/* 编辑公司 */
func (p *DBInfo) updateCompany(company *Company) error {
	s := p.session.Copy()
	defer s.Close()
	c := s.DB(p.dbName).C(CollectionCompanies)
	err := c.Update(bson.M{"cid": company.Cid},bson.M{"$set":
		bson.M{"name": company.Name,"phone": company.Phone,"emial": company.Email,"manager": company.Manager}})
	if err != nil {
		fmt.Printf("编辑公司失败,err:%v\n",err)
		return err
	}
	return err
}

