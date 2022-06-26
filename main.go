package main

import (
	"agent/app"
	"github.com/gogf/gf/os/gproc"
	"github.com/osgochina/dmicro/easyservice"
	"github.com/osgochina/dmicro/logger"
	"github.com/osgochina/dmicro/supervisor/process"
	"os"
)

func main() {
	//if gproc.IsChild() {
	logger.Infof("子进程%d启动成功", os.Getpid())
	easyservice.Setup(func(svr *easyservice.EasyService) {
		//注册服务停止时要执行法方法
		svr.BeforeStop(func(service *easyservice.EasyService) bool {
			logger.Info("BeforeStop: agent server stop")
			return true
		})
		svr.AddSandBox(app.NewSandBoxServer(svr))
	})
	//} else {
	// 回收僵尸进程
	process.ReapZombie()
	m := gproc.NewManager()
	p := m.NewProcess(os.Args[0], os.Args, os.Environ())
	_, err := p.Start()
	if err != nil {
		logger.Warningf("启动子进程报错:%v", err)
		return
	}
	logger.Infof("父进程%d启动成功", os.Getpid())
	_ = p.Wait()
	//}
}
