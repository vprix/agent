package agent_server

import (
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
	"os"
	"single-agent/pkg/customexec"
	"single-agent/pkg/desktop"
	"single-agent/pkg/proto/agent"
	"single-agent/pkg/vncpasswd"
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
	// 要启动的desktop的类型，比如xfce，gnome
	sessionName string
	//桌面名称
	desktopName string
	// 桌面启动的log
	desktopLog string
	xVnc       *desktop.XVnc
}

// 创建vnc服务
func newVncServer() *vncServer {
	hostname, _ := os.Hostname()
	return &vncServer{
		hostname: hostname,
	}
}

// VncStart 启动vnc服务
func (that *vncServer) VncStart(vncConn *agent.VncConn, user *agent.StartUser) (*agent.VncConn, error) {
	val, err := gcache.Get(StartWorkSpaceVncKey)
	if err != nil {
		return nil, err
	}
	if val != nil {
		vncSvr := val.(*agent.VncConn)
		return vncSvr, nil
	}

	if len(vncConn.VncPasswd) <= 0 || len(vncConn.VncPasswd) > 8 {
		return nil, gerror.New("vnc passwd not found.")
	}

	if len(vncConn.ProxyId) <= 0 || len(vncConn.ProxyPasswd) <= 0 {
		return nil, gerror.New("proxy id or proxy passwd not found.")
	}

	if len(user.Home) <= 0 {
		user.Home = "/home/vprix"
	}

	if len(user.UserName) <= 0 {
		user.UserName = "vprix"
	}

	if user.UserId == 0 {
		user.UserId = 1000
	}
	if user.GroupId == 0 {
		user.GroupId = 1000
	}
	that.init(user, vncConn.VncPasswd)
	err = that.startXVnc()
	if err != nil {
		return nil, err
	}
	time.Sleep(time.Second * 1)

	err = that.startDesktop()
	if err != nil {
		return nil, err
	}
	conn := &agent.VncConn{
		VncPasswd:   vncConn.VncPasswd,
		ProxyId:     vncConn.ProxyId,
		ProxyPasswd: vncConn.ProxyPasswd,
		ProxyAddr:   vncConn.ProxyAddr,
	}
	gcache.Set(StartWorkSpaceVncKey, conn, 0)
	gcache.Set(vncConn.ProxyId, vncConn.ProxyPasswd, 0)
	return conn, nil
}

// Init 初始化参数
func (that *vncServer) init(user *agent.StartUser, vncPasswd string) {
	that.home = user.GetHome()
	that.user = user.GetUserName()
	that.group = user.GetGroupName()
	that.uid = int(user.GetUserId())
	that.gid = int(user.GetGroupId())
	that.vncPasswd = vncPasswd

	that.sessionName = genv.Get("VNC_SESSION_NAME", "xfce")
	if that.sessionName == "" {
		logger.Fatalf("VNC SESSION NAME[%s] 不可识别", that.sessionName)
	}
	that.xAuthorityFile = fmt.Sprintf("%s/.Xauthority", that.home)
	that.vncPort = 5900
	//桌面名
	that.desktopName = fmt.Sprintf("%s:%d-%s", that.hostname, that.displayNumber, that.user)
	// 启动桌面的log地址
	that.desktopLog = fmt.Sprintf("%s/.vnc/%s:%d.log", that.home, that.hostname, that.displayNumber)
}

// CreateVncPasswd 创建vnc密码
func (that *vncServer) createVncPasswd(password string) error {
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
func (that *vncServer) startDesktop() error {
	xfce := desktop.NewXfce()
	xfce.User = that.user
	xfce.Home = that.home
	xfce.DisplayNumber = that.displayNumber
	xfce.Hostname = that.hostname
	xfce.DesktopName = that.desktopName
	xfce.LogPath = that.desktopLog
	procEntry, err := xfce.XStartup()
	if err != nil {
		return err
	}
	that.display = xfce.Display
	entry, err := SvrSBox.ProcManager().NewProcessByEntry(procEntry)
	if err != nil {
		return err
	}
	entry.Start(false)
	return nil
}

// 启动vnc进程
func (that *vncServer) startXVnc() error {
	// 创建vnc密码
	err := that.createVncPasswd(that.vncPasswd)
	if err != nil {
		return err
	}
	//创建连接X服务器的认证信息
	err = that.createXAuth()
	if err != nil {
		return err
	}
	// 启动xvnc
	opts := desktop.NewXVncOpts()
	opts.Desktop = gstr.Trim(that.desktopName, "'", "\"")
	opts.RfbAuth = fmt.Sprintf("%s/.vnc/passwd", that.home)
	opts.RfbPort = that.vncPort
	_ = genv.Set("VNC_RESOLUTION", opts.Geometry)
	that.xVnc = desktop.NewXVnc(opts)
	that.xVnc.User = that.user
	that.xVnc.Dir = that.home
	that.xVnc.DisplayNumber = that.displayNumber
	that.xVnc.LogPath = that.desktopLog
	procEntry, err := that.xVnc.NewXVncProcess()
	if err != nil {
		return err
	}
	entry, err := SvrSBox.ProcManager().NewProcessByEntry(procEntry)
	if err != nil {
		return err
	}
	entry.Start(false)

	return nil
}
