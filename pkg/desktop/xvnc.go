package desktop

import (
	"agent/env"
	"fmt"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/text/gstr"
	"github.com/osgochina/dmicro/logger"
	"github.com/osgochina/dmicro/supervisor/process"
)

type XVnc struct {
	opts          *XVncOpts
	DisplayNumber int
	Dir           string
	User          string
	LogPath       string
}

func NewXVnc(opts *XVncOpts) *XVnc {
	return &XVnc{
		opts: opts,
	}
}

// NewXVncProcess 创建xvnc的进程
func (that *XVnc) NewXVncProcess() (*process.ProcEntry, error) {
	proc := process.NewProcEntry("/usr/bin/Xvnc")
	proc.SetArgs(append([]string{fmt.Sprintf(":%d", that.DisplayNumber)}, that.opts.Array()...))
	proc.SetDirectory(that.Dir)
	proc.SetUser(env.User())
	proc.SetAutoReStart("false")
	proc.SetRedirectStderr(true)
	proc.SetStdoutLogfile(that.LogPath)
	proc.SetStdoutLogFileMaxBytes("100MB")
	proc.SetStderrLogFileBackups(10)
	proc.SetEnvironment(genv.All())
	logger.Info("启动xvnc命令: ", "/usr/bin/Xvnc ", gstr.Implode(" ", proc.Args()))
	logger.Info("启动环境变量：", gstr.Implode(";", genv.All()))
	return proc, nil
}
