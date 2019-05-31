package main

import "net/http"

/* 判断用户提交的地址并调用对应的处理函数 */
func (p *OperateHttpListener) Router(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/login" {
		p.showLogin(w, r)
	}
	if r.URL.Path == "/sidebar"{
		p.returnSidebarData(w,r)
	}
	if r.URL.Path == "/handlelogin" {
		p.handleLogin(w, r)
	}
	if r.URL.Path == "/companies" {
		p.showCompanies(w, r)
	}
	if r.URL.Path == "/returncompaniesdata"{
		p.returnCompaniesData(w,r)
	}
	if r.URL.Path == "/addcompanies" {
		p.showAddCompanies(w, r)
	}
	if r.URL.Path == "/handleaddcompies" {
		p.handleAddCompies(w, r)
	}
	if r.URL.Path == "/delcompanies" {
		p.handleDelCompanies(w, r)
	}
	if r.URL.Path == "/editcompanies" {
		p.showEditCompanies(w, r)
	}
	if r.URL.Path == "/returneditcompanies"{
		p.returnEditCompanies(w,r)
	}
	if r.URL.Path == "/handleeditcompanies" {
		p.handleEditCompanies(w, r)
	}
	if r.URL.Path == "/teams" {
		p.showTeams(w, r)
	}
	if r.URL.Path == "/returnteamsdata"{
		p.returnTeamsData(w,r)
	}
	if r.URL.Path == "/addteams" {
		p.showAddTeams(w, r)
	}
	if r.URL.Path == "/handleaddteams" {
		p.handleAddTeams(w, r)
	}
	if r.URL.Path == "/delteams" {
		p.handleDelTeams(w, r)
	}
	if r.URL.Path == "/editteams" {
		p.showEditTeams(w, r)
	}
	if r.URL.Path == "/returneditteamsdata"{
		p.returnEditTeamsData(w,r)
	}
	if r.URL.Path == "/handleeditteams" {
		p.handleEditTeams(w, r)
	}
	if r.URL.Path == "/receivers" {
		p.showReceivers(w, r)
	}
	if r.URL.Path == "/handlereceivers"{
		p.handleReceivers(w,r)
	}
	if r.URL.Path == "/addreceivers" {
		p.showAddReceivers(w, r)
	}
	if r.URL.Path == "/returnaddreceivers"{
		p.returnAddReceivers(w,r)
	}
	if r.URL.Path == "/handleaddreceivers"{
		p.handleAddReceivers(w,r)
	}
	if r.URL.Path == "/delreceivers" {
		p.handleDelReceivers(w, r)
	}
	if r.URL.Path == "/editreceivers" {
		p.showEditReceivers(w, r)
	}
	if r.URL.Path == "/handleeditreceivers"{
		p.handleEditReceivers(w,r)
	}
	if r.URL.Path == "/returneditreceiversdata"{
		p.returnEditReceiversData(w,r)
	}
	if r.URL.Path == "/laststatuses" {
		p.showLastStatuses(w, r)
	}
	if r.URL.Path == "/returnlaststatuses"{
		p.returnLastStatuses(w,r)
	}
	if r.URL.Path == "/dellaststatus" {
		p.handleDelLastStatus(w, r)
	}
	if r.URL.Path == "/rawdatas" {
		p.showRawDatas(w, r)
	}
	if r.URL.Path == "/handlerawdatas"{
		p.returnRawDatas(w,r)
	}
	if r.URL.Path == "/remoteupgrade"{
		p.showRemoteUpgrade(w,r)
	}
	if r.URL.Path == "/remoteshutdown"{
		p.showRemoteShutdown(w,r)
	}
}