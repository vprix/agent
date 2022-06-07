package app

import (
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gcache"
	"github.com/gogf/gf/os/genv"
	"github.com/osgochina/dmicro/easyservice"
	"github.com/osgochina/dmicro/logger"
	"github.com/osgochina/dmicro/supervisor/process"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"golang.org/x/net/websocket"
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
	sandbox.svr.SetPort(8080)
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
			svrCfg := &rfb.ServerConfig{
				Encodings:   encodings.DefaultEncodings,
				DesktopName: []byte("VNC Proxy"),
				Width:       1024,
				Height:      768,
				SecurityHandlers: []rfb.ISecurityHandler{
					&security.ServerAuthNone{},
				},
				//DisableMessageType: []rfb.ServerMessageType{rfb.ServerCutText},
			}
			vncConnParams := val.(*VncConnParams)
			fmt.Println(vncConnParams)
			vncProxy := NewWSVncProxy(svrCfg, nil, vncConnParams)
			h := websocket.Handler(vncProxy.Start)
			h.ServeHTTP(r.Response.Writer, r.Request)
		})
	})
	path := genv.GetVar("VPRIX_VNC_PATH", "/usr/share/vprixvnc")
	that.svr.SetIndexFolder(false)
	that.svr.AddStaticPath("/static", path.String()+"/assets")
	//that.svr.Group("/static", func(group *ghttp.RouterGroup) {
	//	group.GET("/", func(r *ghttp.Request) {
	//		r.Response.Write(gres.Get(r.RequestURI))
	//	})
	//})
	that.vncSvr = NewVncServer()
	go func() {
		_, err := that.vncSvr.VncStart(&StartUser{UserName: "vprix-user", GroupName: "vprix-user", VncPasswd: GetVncPassword()})
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
