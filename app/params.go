package app

import "github.com/gogf/gf/os/gtime"

type StartUser struct {
	VncPasswd string
	Home      string
	UserName  string
	GroupName string
	UserId    uint32
	GroupId   uint32
}

type VncConnParams struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Token     string `json:"token"`
	VncPasswd string `json:"vnc_passwd"`
}

// ShareConnParams 分享链接的参数
type ShareConnParams struct {
	RW             string      `json:"rw"`
	Uid            int         `json:"uid"`
	Nickname       string      `json:"nickname"`
	Token          string      `json:"token"`
	ComputerId     int         `json:"computer_id"`
	ShareTime      *gtime.Time `json:"share_time"`
	ExpirationTime *gtime.Time `json:"expiration_time"`
}
