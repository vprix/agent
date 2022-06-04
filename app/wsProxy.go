package app

import (
	"errors"
	"fmt"
	"github.com/gogf/gf/util/gconv"
	"github.com/vprix/vncproxy/encodings"
	"github.com/vprix/vncproxy/handler"
	"github.com/vprix/vncproxy/messages"
	"github.com/vprix/vncproxy/rfb"
	"github.com/vprix/vncproxy/security"
	"github.com/vprix/vncproxy/session"
	"golang.org/x/net/websocket"
	"net"
	"time"
)

type WSVncProxy struct {
	rfbSvrCfg              *rfb.ServerConfig      // proxy服务端监听vnc客户端的配置信息
	rfbCliCfg              *rfb.ClientConfig      // proxy客户端连接vnc服务端的配置信息
	svrSession             *session.ServerSession // vnc客户端连接到proxy的会话
	cliSession             *session.ClientSession // 链接到vnc服务端的会话
	closed                 chan struct{}
	errorCh                chan error             // 错误通道
	proxyCli2VncSvrMsgChan chan rfb.ClientMessage // proxy客户端发送给vnc服务端的消息通道
	vncSvr2ProxyMsgChan    chan rfb.ServerMessage // vnc服务端发送给proxy客户端的消息通道
	vncCli2ProxyMsgChan    chan rfb.ClientMessage // vnc客户端发送给proxy服务端的消息通道
	proxySvr2VncCliMsgChan chan rfb.ServerMessage // proxy服务端发送给vnc客户端的消息通道
	vncConnParams          *VncConnParams
}

// NewWSVncProxy 生成vnc proxy服务对象
func NewWSVncProxy(svrCfg *rfb.ServerConfig, cliCfg *rfb.ClientConfig, vncConnParams *VncConnParams) *WSVncProxy {
	errorChan := make(chan error, 32)
	vncProxy := &WSVncProxy{
		errorCh:                errorChan,
		closed:                 make(chan struct{}),
		proxyCli2VncSvrMsgChan: make(chan rfb.ClientMessage),
		vncSvr2ProxyMsgChan:    make(chan rfb.ServerMessage),
		vncCli2ProxyMsgChan:    make(chan rfb.ClientMessage),
		proxySvr2VncCliMsgChan: make(chan rfb.ServerMessage),
		vncConnParams:          vncConnParams,
	}

	vncProxy.rfbSvrCfg = svrCfg
	vncProxy.rfbCliCfg = cliCfg
	return vncProxy
}

// Start 启动
func (that *WSVncProxy) Start(ws *websocket.Conn) {
	ws.PayloadType = websocket.BinaryFrame
	that.rfbSvrCfg.Input = that.vncCli2ProxyMsgChan
	that.rfbSvrCfg.Output = that.proxySvr2VncCliMsgChan
	that.rfbSvrCfg.ErrorCh = make(chan error)
	if len(that.rfbSvrCfg.Messages) <= 0 {
		that.rfbSvrCfg.Messages = messages.DefaultClientMessage
	}

	that.rfbSvrCfg.Handlers = []rfb.IHandler{
		&handler.ServerVersionHandler{},
		&handler.ServerSecurityHandler{},
		that, // 把链接到vnc服务端的逻辑加入
		&handler.ServerClientInitHandler{},
		&handler.ServerServerInitHandler{},
		&handler.ServerMessageHandler{},
	}
	err := session.NewServerSession(ws, that.rfbSvrCfg).Server()
	if err != nil {
		that.errorCh <- err
	}
	return
}

func (that *WSVncProxy) handleIO() {

	for {
		select {
		case msg := <-that.rfbCliCfg.ErrorCh:
			// 如果链接到vnc服务端的会话报错，则需要把链接到proxy的vnc客户端全部关闭
			_ = that.svrSession.Close()
			that.errorCh <- msg
		case msg := <-that.rfbSvrCfg.ErrorCh:
			//  链接到proxy的vnc客户端链接报错，则把错误转发给vnc proxy
			that.errorCh <- msg
		case msg := <-that.vncSvr2ProxyMsgChan:
			// 收到vnc服务端发送给proxy客户端的消息，转发给proxy服务端, proxy服务端内部会把该消息转发给vnc客户端
			sSessCfg := that.svrSession.Config().(*rfb.ServerConfig)
			disabled := false
			// 如果该消息禁用，则跳过不转发该消息
			for _, t := range sSessCfg.DisableMessageType {
				if t == msg.Type() {
					disabled = true
					break
				}
			}
			if !disabled {
				sSessCfg.Output <- msg
			}
		case msg := <-that.vncCli2ProxyMsgChan:
			// vnc客户端发送消息到proxy服务端的时候,需要对消息进行检查
			// 有些消息不支持转发给vnc服务端
			switch msg.Type() {
			case rfb.SetPixelFormat:
				// 发现是设置像素格式的消息，则忽略
				//that.rfbCliCfg.PixelFormat = msg.(*messages.SetPixelFormat).PF
				_ = that.cliSession.SetPixelFormat(msg.(*messages.SetPixelFormat).PF)
				that.proxyCli2VncSvrMsgChan <- msg
				continue
			case rfb.SetEncodings:
				// 设置编码格式的消息
				var encTypes []rfb.EncodingType
				// 判断编码是否再支持的列表
				for _, s := range that.cliSession.Encodings() {
					for _, cEnc := range msg.(*messages.SetEncodings).Encodings {
						if cEnc == s.Type() {
							encTypes = append(encTypes, s.Type())
						}
					}
				}
				// 发送编码消息给vnc服务端
				that.proxyCli2VncSvrMsgChan <- &messages.SetEncodings{EncNum: gconv.Uint16(len(encTypes)), Encodings: encTypes}
			default:
				cliCfg := that.cliSession.Config().(*rfb.ClientConfig)
				disabled := false
				for _, t := range cliCfg.DisableMessageType {
					if t == msg.Type() {
						disabled = true
						break
					}
				}
				if !disabled {
					that.proxyCli2VncSvrMsgChan <- msg
				}
			}
		}
	}
}

// Handle 建立远程链接
func (that *WSVncProxy) Handle(sess rfb.ISession) error {
	timeout := 10 * time.Second

	addr := fmt.Sprintf("%s:%d", that.vncConnParams.Host, that.vncConnParams.Port)

	clientConn, err := that.newTarget(addr, timeout)
	if err != nil {
		return err
	}
	// 配置链接信息
	if that.rfbCliCfg == nil {
		that.rfbCliCfg = &rfb.ClientConfig{
			SecurityHandlers: []rfb.ISecurityHandler{&security.ClientAuthVNC{Password: gconv.Bytes(that.vncConnParams.VncPasswd)}},
			Encodings:        encodings.DefaultEncodings,
			ErrorCh:          make(chan error),
			Input:            that.vncSvr2ProxyMsgChan,
			Output:           that.proxyCli2VncSvrMsgChan,
			Handlers:         session.DefaultClientHandlers,
			Messages:         messages.DefaultServerMessages,
		}
	}
	that.cliSession, err = session.NewClient(clientConn, that.rfbCliCfg)
	if err != nil {
		return err
	}
	err = that.cliSession.Connect()
	if err != nil {
		return err
	}
	that.svrSession = sess.(*session.ServerSession)
	that.svrSession.SetWidth(that.cliSession.Width())
	that.svrSession.SetHeight(that.cliSession.Height())
	that.svrSession.SetDesktopName(that.cliSession.DesktopName())
	_ = that.svrSession.SetPixelFormat(that.cliSession.PixelFormat())
	go that.handleIO()
	return nil
}

func (that *WSVncProxy) newTarget(addr string, dialTimeout time.Duration) (net.Conn, error) {
	if len(addr) == 0 {
		return nil, errors.New("addr is empty")
	}
	if dialTimeout <= 0 {
		dialTimeout = 10 * time.Second
	}
	clientConn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return nil, err
	}
	return clientConn, nil
}

func (that *WSVncProxy) Close() {
	that.closed <- struct{}{}
	close(that.proxySvr2VncCliMsgChan)
	close(that.proxyCli2VncSvrMsgChan)
	close(that.vncCli2ProxyMsgChan)
	close(that.vncSvr2ProxyMsgChan)
}

func (that *WSVncProxy) Error() <-chan error {
	return that.errorCh
}
