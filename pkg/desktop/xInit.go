package desktop

import (
	"fmt"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"github.com/osgochina/dmicro/logger"
	"github.com/osgochina/dmicro/supervisor/process"
)

type XInit struct {
	opts          *XVncOpts
	displayNumber int
	chroot        string
	user          string
	logPath       string
}

func NewXInit(opts *XVncOpts) *XInit {
	return &XInit{
		opts: opts,
	}
}
func (that *XInit) Opts() *XVncOpts {
	return that.opts
}

func (that *XInit) SetOpts(opts *XVncOpts) {
	that.opts = opts
}

func (that *XInit) DisplayNumber() int {
	return that.displayNumber
}

func (that *XInit) SetDisplayNumber(displayNumber int) {
	that.displayNumber = displayNumber
}

func (that *XInit) Chroot() string {
	return that.chroot
}

func (that *XInit) SetChroot(chroot string) {
	that.chroot = chroot
}

func (that *XInit) User() string {
	return that.user
}

func (that *XInit) SetUser(user string) {
	that.user = user
}

func (that *XInit) LogPath() string {
	return that.logPath
}

func (that *XInit) SetLogPath(logPath string) {
	that.logPath = logPath
}

func (that *XInit) initEnv() {
	_ = genv.Remove("SESSION_MANAGER")
	_ = genv.Remove("DBUS_SESSION_BUS_ADDRESS")
	_ = genv.Set("XDG_SESSION_TYPE", "x11")
	_ = genv.Set("XKL_XMODMAP_DISABLE", gconv.String(1))
	_ = genv.Set("DISPLAY", fmt.Sprintf(":%d", that.displayNumber))
}

func (that *XInit) NewProcess() (*process.ProcEntry, error) {
	that.initEnv()
	proc := process.NewProcEntry("/usr/bin/xinit")
	// /usr/bin/xinit /usr/bin/dbus-launch startxfce4 --  /usr/bin/Xvnc :0 ......
	proc.SetArgs(append(
		[]string{
			"/usr/bin/startxfce4",
			"--",
			"/usr/bin/Xvnc",
			fmt.Sprintf(":%d", that.displayNumber)}, that.opts.Array()...),
	)
	proc.SetDirectory(that.chroot)
	proc.SetUser(that.user)
	proc.SetAutoReStart("false")
	proc.SetRedirectStderr(true)
	proc.SetStdoutLogfile(that.logPath)
	proc.SetStdoutLogFileMaxBytes("100MB")
	proc.SetStderrLogFileBackups(10)
	proc.SetEnvironment(genv.All())
	logger.Info("启动xinit命令: ", "/usr/bin/xinit ", gstr.Implode(" ", proc.Args()))
	logger.Info("启动环境变量：", gstr.Implode(";", genv.All()))
	return proc, nil
}
