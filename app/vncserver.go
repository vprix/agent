package app

import (
	"agent/env"
	"agent/pkg/customexec"
	"agent/pkg/desktop"
	"agent/pkg/vncpasswd"
	"bytes"
	"fmt"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/os/gcache"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"github.com/gogf/gf/util/grand"
	"github.com/osgochina/dmicro/logger"
	"github.com/osgochina/dmicro/supervisor/process"
	"os"
	"time"
)

const StartWorkSpaceVncKey = "Start_WorkSpace_Vnc_Key"

// vnc服务
type vncServer struct {
	home  string
	user  string
	group string
	uid   int
	gid   int
	// vnc的链接密码
	vncPasswd string
	// 链接vnc的端口
	vncPort int
	//当前计算机名
	hostname string
	// 显示器编号
	displayNumber int
	//显示面板
	display string
	// 权限文件存放地址 "$ENV{HOME}/.Xauthority"
	xAuthorityFile string
	//桌面名称
	desktopName string
	// 桌面启动的log
	desktopLog string
	xInit      *desktop.XInit
}

// NewVncServer 创建vnc服务
func NewVncServer() *vncServer {
	hostname, _ := os.Hostname()
	return &vncServer{
		hostname: hostname,
	}
}

// VncStart 启动vnc服务
func (that *vncServer) VncStart(user *StartUser) (*VncConnParams, error) {
	val, err := gcache.Get(StartWorkSpaceVncKey)
	if err != nil {
		return nil, err
	}
	if val != nil {
		vncSvr := val.(*VncConnParams)
		return vncSvr, nil
	}

	if len(user.VncPasswd) <= 0 {
		return nil, gerror.New("vnc passwd not found.")
	}

	if len(user.Home) <= 0 {
		user.Home = env.Home()
	}

	if len(user.UserName) <= 0 {
		user.UserName = env.User()
	}

	if user.UserId == 0 {
		user.UserId = 1000
	}
	if user.GroupId == 0 {
		user.GroupId = 1000
	}
	that.init(user)
	err = that.startXInit()
	if err != nil {
		return nil, err
	}
	time.Sleep(time.Second * 1)
	that.beforeStart()
	conn := &VncConnParams{
		VncPasswd: user.VncPasswd,
		Host:      "127.0.0.1",
		Port:      5900,
	}
	_ = gcache.Set(StartWorkSpaceVncKey, conn, 0)
	return conn, nil
}

// Init 初始化参数
func (that *vncServer) init(user *StartUser) {
	that.home = user.Home
	that.user = user.UserName
	that.group = user.GroupName
	that.uid = int(user.UserId)
	that.gid = int(user.GroupId)
	that.vncPasswd = user.VncPasswd

	that.xAuthorityFile = fmt.Sprintf("%s/.Xauthority", that.home)
	that.vncPort = 5900
	//桌面名
	that.desktopName = fmt.Sprintf("%s:%d-%s", that.hostname, that.displayNumber, that.user)
	// 启动桌面的log地址
	that.desktopLog = fmt.Sprintf("%s/.vnc/%s:%d.log", that.home, that.hostname, that.displayNumber)
}

// CreateVncPasswd 创建vnc密码
func (that *vncServer) createVncPasswd(password string) error {
	fmt.Println("password:", password)
	passwd := vncpasswd.AuthVNCEncode(gconv.Bytes(password))
	path := fmt.Sprintf("%s/.vnc/passwd", that.home)
	err := gfile.PutBytes(path, passwd)
	if err != nil {
		return err
	}
	osFile, err := gfile.Open(fmt.Sprintf("%s/.vnc", that.home))
	if err != nil {
		return err
	}
	err = osFile.Chown(that.uid, that.gid)
	if err != nil {
		return err
	}
	return nil
}

// 生成随机cookie
func (that *vncServer) mCookie() string {
	cmd := customexec.Command("mcookie")
	cmd.SetUser(that.user)
	cookie, _ := cmd.Output()
	if len(cookie) > 0 {
		return string(cookie)
	}
	return grand.S(32)
}

// CreateXAuth 创建连接X服务器的认证信息。
func (that *vncServer) createXAuth() error {
	if gfile.Exists(that.xAuthorityFile) {
		return nil
	}
	mCookie := that.mCookie()
	tmpXAuthorityFile := "/tmp/tmpXAuthorityFile"
	err := gfile.PutContents(tmpXAuthorityFile,
		fmt.Sprintf("add %s:%d . %sadd %s/unix:%d . %s",
			that.hostname, that.displayNumber, mCookie, that.hostname, that.displayNumber, mCookie),
	)
	if err != nil {
		return err
	}
	_, err = gfile.Create(that.xAuthorityFile)
	if err != nil {
		return err
	}
	// 修改文件的权限
	_ = os.Chown(that.xAuthorityFile, that.uid, that.gid)
	cmd := customexec.Command("/usr/bin/xauth", "-f", that.xAuthorityFile, "source", tmpXAuthorityFile)
	var in bytes.Buffer
	cmd.Stdin = &in
	cmd.Env = genv.All()
	cmd.SetUser(that.user)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// 启动桌面进程
//func (that *vncServer) startDesktop() (err error) {
//	xfce := desktop.NewXfce()
//	xfce.User = that.user
//	xfce.Home = that.home
//	xfce.DisplayNumber = that.displayNumber
//	xfce.Hostname = that.hostname
//	xfce.DesktopName = that.desktopName
//	xfce.LogPath = that.desktopLog
//	var procEntry *process.ProcEntry
//	if env.IsUbuntu() {
//		procEntry, err = xfce.XStartupUbuntu()
//	} else {
//		procEntry, err = xfce.XStartup()
//	}
//	if err != nil {
//		return err
//	}
//	that.display = xfce.Display
//	entry, err := sandbox.ProcManager().NewProcessByEntry(procEntry)
//	if err != nil {
//		return err
//	}
//	go entry.Start(true)
//	return nil
//}

// 启动vnc进程
func (that *vncServer) startXInit() error {
	// 创建vnc密码
	err := that.createVncPasswd(that.vncPasswd)
	if err != nil {
		logger.Error(err)
		return err
	}
	//创建连接X服务器的认证信息
	err = that.createXAuth()
	if err != nil {
		logger.Error(err)
		return err
	}
	// 启动xvnc
	opts := desktop.NewXVncOpts()
	opts.Desktop = gstr.Trim(that.desktopName, "'", "\"")
	opts.RfbAuth = fmt.Sprintf("%s/.vnc/passwd", that.home)
	opts.RfbPort = that.vncPort
	that.xInit = desktop.NewXInit(opts)
	that.xInit.SetUser(that.user)
	that.xInit.SetChroot(that.home)
	that.xInit.SetDisplayNumber(that.displayNumber)
	that.xInit.SetLogPath(that.desktopLog)
	procEntry, err := that.xInit.NewProcess()
	if err != nil {
		logger.Error(err)
		return err
	}
	entry, err := sandbox.ProcManager().NewProcessByEntry(procEntry)
	if err != nil {
		logger.Error(err)
		return err
	}
	entry.Start(false)
	return nil
}

// 启动xinit后，执行自定义脚本
func (that *vncServer) beforeStart() {
	if !gfile.Exists("/dockerstartup/custom_startup.sh") {
		return
	}
	proc := process.NewProcEntry("/dockerstartup/custom_startup.sh")
	proc.SetUser(that.user)
	proc.SetAutoReStart("false")
	proc.SetRedirectStderr(true)
	proc.SetEnvironment(genv.All())
	entry, err := sandbox.ProcManager().NewProcessByEntry(proc)
	if err != nil {
		logger.Error(err)
		return
	}
	entry.Start(false)
}
