package app

import (
	"github.com/gogf/gf/os/glog"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"github.com/vprix/vncproxy/session"
	"github.com/vprix/vncproxy/vnc"
	"golang.org/x/net/websocket"
	"io"
	"net"
	"time"
)

type WSVncProxy struct {
	vncConnParams *VncConnParams
}

// NewWSVncProxy 生成vnc proxy服务对象
func NewWSVncProxy(vncConnParams *VncConnParams) *WSVncProxy {
	vncProxy := &WSVncProxy{
		vncConnParams: vncConnParams,
	}
	return vncProxy
}

// Start 启动
func (that *WSVncProxy) Start(ws *websocket.Conn) {
	ws.PayloadType = websocket.BinaryFrame
	securityHandlers := []rfb.ISecurityHandler{
		&security.ServerAuthNone{},
	}
	targetCfg := rfb.TargetConfig{
		Network:  "tcp",
		Timeout:  10 * time.Second,
		Host:     that.vncConnParams.Host,
		Port:     that.vncConnParams.Port,
		Password: []byte(that.vncConnParams.VncPasswd),
	}
	var err error
	svrSess := session.NewServerSession(
		rfb.OptDesktopName([]byte("Vprix VNC Proxy")),
		rfb.OptHeight(768),
		rfb.OptWidth(1024),
		rfb.OptSecurityHandlers(securityHandlers...),
		rfb.OptGetConn(func(sess rfb.ISession) (io.ReadWriteCloser, error) {
			return ws, nil
		}),
	)
	cliSess := session.NewClient(
		rfb.OptSecurityHandlers([]rfb.ISecurityHandler{&security.ClientAuthVNC{Password: targetCfg.Password}}...),
		rfb.OptGetConn(func(sess rfb.ISession) (io.ReadWriteCloser, error) {
			return net.DialTimeout(targetCfg.GetNetwork(), targetCfg.Addr(), targetCfg.GetTimeout())
		}),
	)
	p := vnc.NewVncProxy(cliSess, svrSess)
	err = p.Start()
	if err != nil {
		glog.Warning(err)
		return
	}
	for {
		select {
		case err = <-p.Error():
			glog.Warning(err)
			return
		case <-p.Wait():
			return
		}
	}
}
