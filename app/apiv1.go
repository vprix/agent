package app

import (
	"agent/env"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gcache"
	"github.com/osgochina/dmicro/logger"
)

type UserCheck struct {
	Username string `json:"username" v:"required#请传入用户账号名称，只包含英文和数字"`                // 用户名，必须是以英文字母开头的字符串，可以包含的字符只有英文和数字
	Password string `json:"password" v:"password2#请传入正确格式的密码，长度6-18之间，必须包含大小写字母和数字"` // 登录密码
}

type ControllerApiV1 struct{}

// Login 登录
func (that *ControllerApiV1) Login(r *ghttp.Request) {

	username := r.GetString("username")
	passwd := r.GetString("passwd")
	if len(username) <= 0 || len(passwd) <= 0 {
		FailJson(true, r, "用户名或密码错误")
		return
	}
	if username != env.UserName() || passwd != env.Password() {
		FailJson(true, r, "用户名或密码错误")
		return
	}
	uid := 1000
	err := r.Session.Set("username", username)
	err = r.Session.Set("uid", uid)
	_ = r.Session.Set("isLogin", true)
	if err != nil {
		FailJson(true, r, err.Error())
		return
	}
	SusJson(true, r, "登录成功")
}

// Logout 用户退出登录
func (that *ControllerApiV1) Logout(r *ghttp.Request) {
	err := r.Session.Clear()
	if err != nil {
		FailJson(true, r, "退出失败")
		return
	}
	SusJson(true, r, "退出成功")
}

// Share 通过分享链接中的token，获取vnc链接信息
func (that *ControllerApiV1) Share(r *ghttp.Request) {
	token := r.GetString("token")
	if len(token) <= 0 {
		FailJson(true, r, "获取token失败")
		return
	}
	data, err := gcache.Get(token)
	if err != nil {
		logger.Warning(err)
		FailJson(true, r, "获取token信息失败")
		return
	}
	if data == nil {
		FailJson(true, r, "获取token信息失败")
		return
	}
	shareConnParams, ok := data.(*ShareConnParams)
	if !ok {
		FailJson(true, r, "获取token信息失败")
		return
	}
	params, err := gcache.Get(shareConnParams.Token)
	if err != nil {
		logger.Warning(err)
		FailJson(true, r, "获取params信息失败")
		return
	}
	if data == nil {
		FailJson(true, r, "获取params信息失败")
		return
	}
	p, ok := params.(*VncConnParams)
	if !ok {
		FailJson(true, r, "获取params信息失败")
		return
	}
	logger.Infof("分享桌面：%v", shareConnParams)
	d := g.Map{
		"token":    p.Token,
		"nickname": shareConnParams.Nickname,
		"host":     g.Cfg().GetString("base_url"),
		"path":     "/proxy/v1/websockify",
		"passwd":   p.VncPasswd,
		"rw":       shareConnParams.RW,
	}
	SusJson(true, r, "ok", d)
}

// Join 加入到实训机
func (that *ControllerApiV1) Join(r *ghttp.Request) {
	uid := r.Session.GetInt("uid", 0)
	if uid <= 0 {
		FailJson(true, r, "请登录")
		return
	}
	userComputerId := r.GetInt("user_computer_id", 0)
	if userComputerId <= 0 {
		FailJson(true, r, "请选择要登录的实训机")
		return
	}
	token := ""
	SusJson(true, r, "ok", g.Map{"token": token, "url": g.Cfg().GetString("base_url") + "/api/v1/manager"})
}

// Stop 强制停止实训机
func (that *ControllerApiV1) Stop(r *ghttp.Request) {
	uid := r.Session.GetInt("uid", 0)
	if uid <= 0 {
		FailJson(true, r, "请登录")
		return
	}
	userComputerId := r.GetInt("user_computer_id", 0)
	if userComputerId <= 0 {
		FailJson(true, r, "请选择要停止的实训机")
		return
	}
	SusJson(true, r, "停止成功")
}

// Clipboard 复制文本内容到实训机
func (that *ControllerApiV1) Clipboard(r *ghttp.Request) {

	text := r.GetString("text")
	if len(text) <= 0 || len(text) > 4*1024 {
		FailJson(true, r, "要复制的内容应该在4K以内")
		return
	}
	SusJson(true, r, "ok", "复制成功")
}
