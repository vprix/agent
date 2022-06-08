package desktop

import (
	"agent/env"
	"fmt"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gstr"
	"github.com/osgochina/dmicro/logger"
	"github.com/osgochina/dmicro/supervisor/process"
)

// Xfce 启动xfce桌面
type Xfce struct {
	// 显示器编号
	DisplayNumber int
	// 计算机名
	Hostname string
	// 桌面名称
	DesktopName string

	//显示地址
	Display string

	// home目录
	Home string

	// 进程运行用户
	User string

	// 运行日志文件路径
	LogPath string
}

func NewXfce() *Xfce {
	return &Xfce{}
}

// XStartup 启动Xface桌面
func (that *Xfce) XStartup() (*process.ProcEntry, error) {
	xStartup := NewXStartup()
	if gfile.Exists(fmt.Sprintf(":%d", that.DisplayNumber)) ||
		gfile.Exists(fmt.Sprintf("/usr/spool/sockets/X11/%d", that.DisplayNumber)) {
		that.Display = fmt.Sprintf(":%d", that.DisplayNumber)
	} else {
		that.Display = fmt.Sprintf("%s:%d", that.Hostname, that.DisplayNumber)
	}

	xStartup.SetDisplay(that.Display)
	xStartup.SetVncDesktop(that.DesktopName)
	xStartup.InitEnv()
	xSet := NewXSet()
	err := xSet.DpmsClose()
	if err != nil {
		return nil, gerror.Newf("Dpms Close:%v", err)
	}
	err = xSet.NoBlank()
	if err != nil {
		return nil, gerror.Newf("NoBlank:%v", err)
	}
	err = xSet.SOff()
	if err != nil {
		return nil, gerror.Newf("SOff:%v", err)
	}
	proc := process.NewProcEntry("/usr/bin/dbus-launch", []string{"startxfce4"})
	proc.SetDirectory(that.Home)
	//proc.SetUser(that.User)
	proc.SetUser(env.User())
	// 初始化工具要使用的环境变量，后期需要放到一个统一的地方
	proc.SetEnvironment(genv.All())
	proc.SetAutoReStart("false")
	proc.SetRedirectStderr(true)
	proc.SetStdoutLogfile(that.LogPath)
	proc.SetStdoutLogFileMaxBytes("100MB")
	proc.SetStderrLogFileBackups(10)
	logger.Info("启动xfce命令: /usr/bin/dbus-launch startxfce4")
	logger.Info("启动环境变量：", gstr.Implode(";", genv.All()))

	return proc, nil
}
