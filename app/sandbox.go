package app

import (
	"agent/env"
	"github.com/gobuffalo/packr/v2"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gcache"
	"github.com/osgochina/dmicro/easyservice"
	"github.com/osgochina/dmicro/logger"
	"github.com/osgochina/dmicro/supervisor/process"
	"golang.org/x/net/websocket"
	"net/http"
)

// SandBoxServer 容器启动后对外提供的管理接口，使用加密传输的rpc协议
type SandBoxServer struct {
	id          int
	name        string
	service     *easyservice.EasyService
	svr         *ghttp.Server
	procManager *process.Manager
	vncSvr      *vncServer
}

var _ easyservice.ISandBox = new(SandBoxServer)
var sandbox *SandBoxServer

// NewSandBoxServer 创建服务沙盒对象
func NewSandBoxServer(svc *easyservice.EasyService) *SandBoxServer {
	sandbox = &SandBoxServer{
		id:      easyservice.GetNextSandBoxId(),
		name:    "vprix",
		service: svc,
	}
	sandbox.svr = g.Server("vprix")
	sandbox.svr.SetPort(env.VprixPort())
	sandbox.procManager = process.NewManager()
	return sandbox
}
func MiddlewareCORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}

// Setup 启动服务
func (that *SandBoxServer) Setup() error {
	that.svr.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(MiddlewareCORS)
		group.ALL("/", new(ControllerApiV1))
	})
	that.svr.Group("/websockify", func(group *ghttp.RouterGroup) {
		group.Middleware(MiddlewareCORS)
		group.ALL("/", func(r *ghttp.Request) {

			//if !r.Session.GetVar("isLogin", false).Bool() {
			//	logger.Warningf("用户未登录")
			//	r.Exit()
			//	return
			//}
			val, err := gcache.Get(StartWorkSpaceVncKey)
			if err != nil {
				logger.Error(err)
				r.Exit()
				return
			}
			if val == nil {
				logger.Warningf("vnc服务未启动")
				r.Exit()
				return
			}
			vncConnParams := val.(*VncConnParams)
			vncProxy := NewWSVncProxy(vncConnParams)
			h := websocket.Handler(vncProxy.Start)
			h.ServeHTTP(r.Response.Writer, r.Request)
		})
	})
	box := packr.New("novnc", "../assets/novnc")
	that.svr.Group("/*", func(group *ghttp.RouterGroup) {
		group.GET("/core/", func(r *ghttp.Request) {
			http.FileServer(box).ServeHTTP(r.Response.Writer, r.Request)
		})
		group.GET("/", func(r *ghttp.Request) {
			http.FileServer(box).ServeHTTP(r.Response.Writer, r.Request)
		})
	})

	that.vncSvr = NewVncServer()
	go func() {
		_, err := that.vncSvr.VncStart(&StartUser{UserName: env.User(), GroupName: env.User(), VncPasswd: env.VncPassword()})
		if err != nil {
			logger.Error(err)
		}
	}()

	return that.svr.Start()
}

// Shutdown 关闭服务
func (that *SandBoxServer) Shutdown() error {
	that.procManager.StopAllProcesses()
	return nil
}

func (that *SandBoxServer) ID() int {
	return that.id
}

func (that *SandBoxServer) Name() string {
	return that.name
}

func (that *SandBoxServer) Service() *easyservice.EasyService {
	return that.service
}

func (that *SandBoxServer) ProcManager() *process.Manager {
	return that.procManager
}
