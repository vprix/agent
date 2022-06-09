package desktop

//type XVnc struct {
//	opts          *XVncOpts
//	DisplayNumber int
//	Dir           string
//	User          string
//	LogPath       string
//}
//
//func NewXVnc(opts *XVncOpts) *XVnc {
//	return &XVnc{
//		opts: opts,
//	}
//}
//
//// NewXVncProcess 创建xvnc的进程
//func (that *XVnc) NewXVncProcess() (*process.ProcEntry, error) {
//	proc := process.NewProcEntry("xinit")
//	proc.SetArgs(append([]string{"/usr/bin/dbus-launch", "startxfce4", "--", "/usr/bin/Xvnc", fmt.Sprintf(":%d", that.DisplayNumber)}, that.opts.Array()...))
//	//proc.SetArgs(append([]string{"/etc/X11/Xsession", "startxfce4", "--", "/usr/bin/Xvnc", fmt.Sprintf(":%d", that.DisplayNumber)}, that.opts.Array()...))
//	proc.SetDirectory(that.Dir)
//	proc.SetUser(env.User())
//	proc.SetAutoReStart("false")
//	proc.SetRedirectStderr(true)
//	proc.SetStdoutLogfile(that.LogPath)
//	proc.SetStdoutLogFileMaxBytes("100MB")
//	proc.SetStderrLogFileBackups(10)
//	proc.SetEnvironment(genv.All())
//	logger.Info("启动xvnc命令: ", "xinit ", gstr.Implode(" ", proc.Args()))
//	logger.Info("启动环境变量：", gstr.Implode(";", genv.All()))
//	return proc, nil
//}
