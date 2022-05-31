package app

import (
	"bitcloud-proxy/app/bittopone/dao"
	"bitcloud-proxy/library"
	"fmt"
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gcache"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/grand"
	"github.com/osgochina/dmicro/logger"
	"time"
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
	cli, err := Instance().ServiceManager(r.Session.Id())
	if err != nil {
		JsonExit(r, 1001, err.Error())
		return
	}
	userCheck := &UserCheck{
		Username: username,
		Password: passwd,
	}
	var result *gjson.Json
	stat := cli.Call("/v1/user/check", userCheck, &result).Status()
	if !stat.OK() {
		JsonExit(r, 1002, stat.Msg())
		return
	}
	dataJson := result.GetJson("data")
	uid := dataJson.GetInt("uid", 0)
	if uid <= 0 {
		FailJson(true, r, "用户名或密码错误")
		return
	}
	err = r.Session.Set("uid", uid)
	err = r.Session.Set("isAdmin", false)
	err = r.Session.Set("username", dataJson.GetString("username"))
	err = r.Session.Set("nickname", dataJson.GetString("nickname"))
	_ = r.Session.Set("isLogin", true)
	if err != nil {
		FailJson(true, r, err.Error())
		return
	}
	SusJson(true, r, "登录成功", g.Map{"uid": uid})
}

// AdminLogin 验证管理员登录
func (that *ControllerApiV1) AdminLogin(r *ghttp.Request) {

	Certificate := r.GetString("certificate")
	if len(Certificate) <= 0 {
		FailJson(true, r, "参数有误")
		return
	}
	cli, err := Instance().ServiceManager(r.Session.Id())
	if err != nil {
		JsonExit(r, 1001, err.Error())
		return
	}
	var result *gjson.Json
	stat := cli.Call("/v1/user/check_admin_certificate", Certificate, &result).Status()
	if !stat.OK() {
		JsonExit(r, 1002, stat.Msg())
		return
	}
	dataJson := result.GetJson("data")
	userComputerId := dataJson.GetInt("user_computer_id", 0)
	if userComputerId <= 0 {
		FailJson(true, r, "用户实训机id错误")
		return
	}
	uid := dataJson.GetInt("uid", 0)
	err = r.Session.Set("isAdmin", true)
	err = r.Session.Set("userComputerId", userComputerId)
	err = r.Session.Set("uid", uid)
	if err != nil {
		FailJson(true, r, err.Error())
		return
	}
	token := that.createComputerToken(userComputerId, uid)
	SusJson(true, r, "ok", g.Map{"token": token, "bitsessionid": r.Session.Id(), "url": g.Cfg().GetString("base_url") + "/api/v1/manager"})
}

// UserInfo 获取用户信息
func (that *ControllerApiV1) UserInfo(r *ghttp.Request) {
	uid := r.Session.GetInt("uid", 0)
	if uid <= 0 {
		FailJson(true, r, "请登录")
		return
	}
	username := r.Session.GetString("username")
	nickname := r.Session.GetString("nickname")
	SusJson(true, r, "ok", g.Map{"uid": uid, "username": username, "nickname": nickname})
}

type UserChangePassword struct {
	Uid         int    `json:"uid"`                                // 用户名，必须是以英文字母开头的字符串，可以包含的字符只有英文和数字
	PasswordOld string `json:"password_old" v:"required#请传入旧密码"`   // 登录密码
	Password    string `json:"password" v:"required#请传入要修改密码"`     // 登录密码
	Password2   string `json:"password2" v:"required#请重复传入要修改的密码"` // 登录密码
}

func (that *ControllerApiV1) ChangePassword(r *ghttp.Request) {
	uid := r.Session.GetInt("uid", 0)
	if uid <= 0 {
		FailJson(true, r, "请登录")
		return
	}
	password_old := r.GetString("password_old")
	password := r.GetString("password")
	password2 := r.GetString("password2")
	if password != password2 {
		FailJson(true, r, "两次输入的密码不相等")
		return
	}
	userChangePassword := &UserChangePassword{
		Uid:         uid,
		PasswordOld: password_old,
		Password:    password,
		Password2:   password2,
	}
	cli, err := Instance().ServiceManager(r.Session.Id())
	if err != nil {
		FailJson(true, r, "链接服务失败")
		return
	}
	var result *gjson.Json
	stat := cli.Call("/v1/user/change_password", userChangePassword, &result).Status()
	if !stat.OK() {
		FailJson(true, r, stat.Msg())
		return
	}
	if result.IsNil() {
	}
	SusJson(true, r, "ok")
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
	shareConnParams, ok := data.(*library.ShareConnParams)
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
	p, ok := params.(*library.VncConnParams)
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

type UserComputers struct {
	Uid    int64 `json:"uid"`
	Status int   `json:"status"`
}

type Computers struct {
	CourseName     string `json:"course_name"`
	UserComputerId uint   `json:"user_computer_id"`
	TemplateId     int    `json:"template_id"`
	TemplateName   string `json:"template_name"`
	Cpu            string `json:"cpu"`
	Memory         string `json:"memory"`
	Status         string `json:"status"`
	Img            string `json:"img"`
}

// GetComputers 获取用户实训机列表
func (that *ControllerApiV1) GetComputers(r *ghttp.Request) {
	uid := r.Session.GetInt64("uid", 0)
	if uid <= 0 {
		FailJson(true, r, "请登录")
		return
	}
	cli, err := Instance().ServiceManager(r.Session.Id())
	if err != nil {
		FailJson(true, r, "链接服务失败")
		return
	}
	var userComputers = &UserComputers{
		Uid:    uid,
		Status: 0,
	}
	var result []*Computers
	stat := cli.Call("/v1/user/computers", userComputers, &result).Status()
	if !stat.OK() {
		FailJson(true, r, stat.Msg())
		return
	}
	if len(result) > 0 {
		for k, v := range result {
			if len(v.Img) > 0 {
				v.Img = fmt.Sprintf("data:image/png;base64,%s", v.Img)
				result[k] = v
			}
		}
	}

	SusJson(true, r, "ok", result)
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
	token := that.createComputerToken(userComputerId, uid)
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
	cli, err := Instance().ServiceManager(r.Session.Id())
	if err != nil {
		JsonExit(r, 1001, err.Error())
		return
	}
	params := &CheckComputer{Uid: uid, ComputerId: userComputerId}
	var result *gjson.Json
	stat := cli.Call("/v1/computer/check", params, &result).Status()
	if !stat.OK() {
		FailJson(true, r, stat.Msg())
	}
	params2 := &JoinComputer{ComputerId: userComputerId}
	stat = cli.Call("/v1/computer/stop", params2, &result).Status()
	if !stat.OK() {
		FailJson(true, r, stat.Msg())
		return
	}
	SusJson(true, r, "停止成功")
}

func (that *ControllerApiV1) createComputerToken(userComputerId int, uid int) string {
	token := grand.S(32)
	params := gmap.New(true)
	params.Set("computer_id", userComputerId)
	params.Set("uid", uid)
	gcache.Set(token, params, 300*time.Second)
	return token
}

func (that *ControllerApiV1) GetToken(r *ghttp.Request) {
	host := g.Cfg().GetString("nodeManagerHost")
	cli, stat := Instance().endpoint.Dial(host)
	if !stat.OK() {
		FailJson(true, r, "链接服务失败")
		return
	}
	geometry := r.GetString("geometry", "1224x800")
	args := &library.InitDesktopParams{
		DesktopName: "xfce",
		DesktopDir:  "/desktop",
		Home:        fmt.Sprintf("/home/%s", "bitcloud"),
	}
	var ret int
	stat = cli.Call("/manager/init_desktop", args, &ret).Status()
	if !stat.OK() {
		FailJson(true, r, stat.String())
		return
	}

	token := grand.S(32)
	VncPasswd := "12121212"
	params := &library.VncStartParams{
		Home:        "/home/bitcloud",
		UserName:    "bitcloud",
		UserId:      1000,
		GroupId:     1000,
		VncPasswd:   VncPasswd,
		ProxyId:     grand.S(8),
		ProxyPasswd: grand.S(8),
		ProxyAddr:   "127.0.0.1:5900",
		Geometry:    geometry,
	}
	gcache.Set(token, params, 0)
	var result = new(library.VncConnParams)
	stat = cli.Call("/manager/vnc_start", params, result).Status()
	if !stat.OK() {
		FailJson(true, r, stat.String())
		return
	}

	// 设置监控上报参数，如果失败了继续执行，并不阻止主流程的执行
	var cResult int
	stat = cli.Call("/manager/set_monitor", that.getMonitorParams(1), &cResult).Status()
	if !stat.OK() {
		logger.Warning(stat.String())
	}
	result.Token = token
	result.Host = "127.0.0.1"
	result.Port = 5899
	logger.Info(result)
	gcache.Set(token, result, 0)
	SusJson(true, r, "ok", g.Map{"token": token, "host": g.Cfg().Get("host"), "port": g.Cfg().Get("port"), "path": "proxy/v1/websockify", "passwd": VncPasswd})
}

type SetMonitor struct {
	Enable         bool   `json:"enable"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	ReportInterval int    `json:"report_interval"`
	UserComputerId int    `json:"user_computer_id"`
}

func (that *ControllerApiV1) getMonitorParams(userComputerId uint) *SetMonitor {
	return &SetMonitor{
		Enable:         true,
		Host:           "172.17.0.1",
		Port:           5790,
		ReportInterval: 30,
		UserComputerId: int(userComputerId),
	}
}

// Clipboard 复制文本内容到实训机
func (that *ControllerApiV1) Clipboard(r *ghttp.Request) {
	cli, stat := Instance().endpoint.Dial(g.Cfg().GetString("nodeManagerHost"))
	if !stat.OK() {
		FailJson(true, r, "链接服务失败")
		return
	}
	text := r.GetString("text")
	if len(text) <= 0 || len(text) > 4*1024 {
		FailJson(true, r, "要复制的内容应该在4K以内")
		return
	}
	var result = new(library.VncConnParams)
	stat = cli.Call("/manager/clipboard", text, result).Status()
	if !stat.OK() {
		FailJson(true, r, stat.String())
		return
	}
	SusJson(true, r, "ok", "复制成功")
}

// ShowNotices 展示通知列表
func (that *ControllerApiV1) ShowNotices(r *ghttp.Request) {
	page := r.GetInt("pageNum", 1)
	pageSize := r.GetInt("pageSize", 20)
	total, err := dao.Notice.Ctx(r.Context()).Where(dao.Notice.Columns.Status, 0).WhereGTE(dao.Notice.Columns.PublicTime, gtime.Now()).Count()
	if err != nil {
		FailJson(true, r, "读取通知出错")
	}
	resultList, err := dao.Notice.Ctx(r.Context()).
		Where(dao.Notice.Columns.Status, 0).
		WhereGTE(dao.Notice.Columns.PublicTime, gtime.Now()).
		Fields("id,title,public_time").
		Order("public_time desc,id desc").Page(page, pageSize).All()
	if err != nil {
		FailJson(true, r, "读取通知出错")
	}
	result := g.Map{
		"currentPage": page,
		"total":       total,
		"list":        resultList,
	}
	SusJson(true, r, "ok", result)
}

// Notice 展示单条通知信息
func (that *ControllerApiV1) Notice(r *ghttp.Request) {
	noticeId := r.GetInt("id")
	if noticeId <= 0 {
		FailJson(true, r, "传入的id不正确")
	}
	result, err := dao.Notice.Ctx(r.Context()).Where(dao.Notice.Columns.Id, noticeId).Where(dao.Notice.Columns.Status, 0).Fields("id,title,content,public_time").One()
	if err != nil {
		FailJson(true, r, "读取通知出错")
	}
	SusJson(true, r, "ok", result)
}
