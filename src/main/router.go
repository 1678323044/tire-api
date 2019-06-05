package main

import "net/http"

/* 判断用户提交的地址并调用对应的处理函数 */
func (p *HttpListener) Router(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/companies"{
		p.handleCompanies(w,r)
	}
	if r.URL.Path == "/addCompany" {
		p.handleAddCompany(w,r)
	}
	if r.URL.Path == "/editCompany" {
		p.handleEditCompany(w,r)
	}
	if r.URL.Path == "/rawdatas"{
		p.handleRawDatas(w,r)
	}
}