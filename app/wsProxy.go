package proxy

import (
	"bitcloud-proxy/library"
	"bitcloud-proxy/vncproxy/encodings"
	"bitcloud-proxy/vncproxy/handler"
	"bitcloud-proxy/vncproxy/messages"
	"bitcloud-proxy/vncproxy/pkg/dbuffer"
	"bitcloud-proxy/vncproxy/rfb"
	"bitcloud-proxy/vncproxy/security"
	"bitcloud-proxy/vncproxy/session"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/net/gtcp"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gres"
	"github.com/gogf/gf/util/gconv"
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
	vncConnParams          *library.VncConnParams
}

// NewWSVncProxy 生成vnc proxy服务对象
func NewWSVncProxy(svrCfg *rfb.ServerConfig, cliCfg *rfb.ClientConfig, vncConnParams *library.VncConnParams) *WSVncProxy {
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
	that.rfbSvrCfg.ClientMessageCh = that.vncCli2ProxyMsgChan
	that.rfbSvrCfg.ServerMessageCh = that.proxySvr2VncCliMsgChan
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
				sSessCfg.ServerMessageCh <- msg
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
	// 写入认证信息
	bb := dbuffer.GetByteBuffer()
	defer dbuffer.ReleaseByteBuffer(bb)
	jsonStr, _ := gjson.New(that.vncConnParams).ToJsonString()
	//写入长度
	_ = binary.Write(bb, binary.BigEndian, uint32(len(jsonStr)+2))
	//写入版本
	_ = binary.Write(bb, binary.BigEndian, uint16(1))
	//写入参数
	_, _ = bb.WriteString(jsonStr)
	_, err = clientConn.Write(bb.Bytes())
	if err != nil {
		return fmt.Errorf("send packet to vnc target failed: %v", err)
	}
	// 配置链接信息
	if that.rfbCliCfg == nil {
		that.rfbCliCfg = &rfb.ClientConfig{
			SecurityHandlers: []rfb.ISecurityHandler{&security.ClientAuthVNC{Password: gconv.Bytes(that.vncConnParams.VncPasswd)}},
			Encodings:        encodings.DefaultEncodings,
			ErrorCh:          make(chan error),
			ServerMessageCh:  that.vncSvr2ProxyMsgChan,
			ClientMessageCh:  that.proxyCli2VncSvrMsgChan,
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
	//clientConn, err := net.DialTimeout("tcp", addr, dialTimeout)
	//if err != nil {
	//	return nil, err
	//}
	//return clientConn, nil
	if dialTimeout <= 0 {
		dialTimeout = 10 * time.Second
	}
	cert, err := tls.X509KeyPair(gres.Get("cert/client.pem").Content(), gres.Get("cert/client.key").Content())
	if err != nil {
		return nil, err
	}
	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(gres.Get("cert/client.pem").Content())
	if !ok {
		return nil, gerror.New("failed to parse root certificate")
	}
	conf := &tls.Config{
		RootCAs:            clientCertPool,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
		GetConfigForClient: func(clientHello *tls.ClientHelloInfo) (*tls.Config, error) {
			// 检查底层是否使用的是tcp链接，如果是tcp链接，则为该链接设置
			if tcpConn, ok := clientHello.Conn.(*net.TCPConn); ok {
				if err = tcpConn.SetKeepAlive(true); err != nil {
					glog.Println("Could not set keep alive", err)
				}
				if err = tcpConn.SetKeepAlivePeriod(69 * time.Second); err != nil {
					glog.Println("Could not set keep alive period", err)
				} else {
					glog.Println("update keep alive period")
				}
			} else {
				glog.Println("TLS over non-TCP connection")
			}
			return nil, nil
		},
	}
	c, err := gtcp.NewNetConnTLS(addr, conf, dialTimeout)
	//c, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return nil, gerror.Wrap(err, "cannot connect to vnc backend")
	}

	return c, nil
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
