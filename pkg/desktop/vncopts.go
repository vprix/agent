package desktop

import (
	"fmt"
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/errors/gerror"
	"github.com/gogf/gf/text/gregex"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
)

// XVncOpts xvnc 服务启动的参数 https://tigervnc.org/doc/Xvnc.html
type XVncOpts struct {
	// 每个桌面都有一个名字，客户端会显示改名字，默认是"x11"
	Desktop string
	// 桌面分辨率 格式是width x height，指定要创建的桌面的大小。默认值为 1024x768。
	Geometry string
	// 指定要创建的桌面位的像素深度。默认值为 24，其他可能值为 16 和 32。任何其他内容都可能导致应用程序出现奇怪的行为，并可能阻止服务器启动。
	Depth int
	// 指定用于服务器的像素格式（BGRnnn 或 RGBnnn）。深度 16 的默认值为 RGB565，深度 24 和 32 的默认值为 RGB888。
	PixelFormat string

	//监听的端口 格式是 interface IP address,所以这里的赋值是 "127.0.0.1"
	Interface string
	// 这极大地改变了Xvnc的行为，使它可以从inetd启动。请参阅下面关于inetd用法的部分。
	//inetd bool
	// 指定xVnc接收客户端链接的tcp端口，默认是5900+displayNumber
	RfbPort int
	//是否使用 ipv4
	UseIPv4 bool
	//是否使用ipv6
	UseIPv6 bool
	// 指定xVnc接收客户端链接地址，可以不适用端口
	RfbUnixPath string
	//unix socket 文件权限默认是 0600
	RfbUnixMode string

	//Time in milliseconds to wait for a viewer which is blocking the server.
	//This is necessary because the server is single-threaded and sometimes blocks
	//until the viewer has finished sending or receiving a message
	//- note that this does not mean an update will be aborted after this time. Default is 20000 (20 seconds).
	// −ClientWaitTimeMillis 这连个参数作用是一样
	RfbWait int
	// 服务器的密码地址，−PasswordFile passwd-file 这两个也是一样的
	RfbAuth string

	//是否支持剪切板，默认是支持
	AcceptCutText bool
	// 剪切板支持的最大长度，默认262144
	MaxCutText int
	// 是否发送剪切内容到客户端，默认是支持
	SendCutText bool

	//Send the primary selection and cut buffer to the server as well as the clipboard selection. Default is on.
	SendPrimary bool

	//接收客户端的鼠标事件 默认打开
	AcceptPointerEvents bool

	// 接收客户端的键盘事件 默认打开
	AcceptKeyEvents bool

	// 接收客户端的桌面尺寸设置 默认打开
	AcceptSetDesktopSize bool

	//如果传入的连接是non-shared，则断开现有客户端。默认的是。
	//如果DisconnectClients为false，则当有一个活跃的客户端时，将拒绝一个新的非共享连接。
	//当与NeverShared组合时，这意味着一次只允许一个客户端。
	DisconnectClients bool

	// 不共享桌面，默认是off
	NeverShared bool

	// 永远共享桌面 默认是off
	AlwaysShared bool

	// 使用 Protocol3.3的协议，默认是off
	Protocol33 bool

	// 每秒发送给客户端的帧数，当桌面更新速度很快的时候，会压缩发送，注意，这只控制最大速率，当资源有限时，客户端可能会获得较低的速率。默认是60。
	FrameRate int

	// 对framebuffer进行像素比较，减少不必要的更新。可以是0(关闭)，1(总是)或2(自动)。默认是2。
	CompareFB int
	//ZRLE编码的Zlib压缩级别(不影响Tight编码)。可接受的值在0到9之间。默认情况下，使用zlib(3)压缩库提供的标准压缩级别。
	ZlibLevel int
	// 使用改进的压缩算法进行Hextile编码，通过略微增加CPU时间的代价获得更好的压缩比。默认的on。
	ImprovedHextile bool

	//指定为传入连接使用哪个安全方案。有效值为None、VncAuth、Plain、TLSNone、TLSVnc、TLSPlain、X509None、X509Vnc和X509Plain，用逗号分隔。
	//默认是VncAuth, TLSVnc。
	SecurityTypes string

	//客户端访问服务器时必须提供的密码的模糊二进制编码。使用该参数是不安全的，请使用PasswordFile参数。
	Password string

	// 一个以逗号分隔的用户名列表，允许通过任何“普通”安全类型(Plain, TLSPlain等)进行身份验证。指定*以允许任何用户使用此安全类型进行身份验证。默认是拒绝所有用户。
	PlainUsers string

	//对使用任何“Plain”安全类型的用户进行身份验证时使用的PAM服务名称。默认是vnc。
	PAMService string

	// 用于所有基于X509的安全类型(X509None, X509Vnc等)的PEM格式的X509证书的路径。
	X509Cert string

	//X509Cert中给出的证书的私钥对应部分。也必须是PEM格式。
	X509Key string

	//GnuTLS优先级字符串，控制TLS session握手算法。有关可能的值，请参阅GnuTLS手册。默认是正常的。
	GnuTLSPriority string

	// 如果主机的身份验证多次失败，则暂时拒绝它的连接。默认on。
	UseBlacklist bool

	// 在任何一台主机被列入黑名单之前，允许从该主机进行未经身份验证的连接尝试的次数。默认是5。
	BlacklistThreshold int

	// 首次加入黑名单时应用的初始超时时间。主机不能重新尝试连接，直到超时过期。默认是10。
	BlacklistTimeout int

	// 闲置VNC连接将被丢弃的秒数。默认值是0，这意味着空闲的连接永远不会被删除。
	IdleTimeout int

	// 当N秒内没有客户端连接时终止。默认值为0。
	MaxDisconnectionTime int

	// 当客户端已连接N秒时终止。默认值为0。
	MaxConnectionTime int

	// 在用户不活动N秒后终止。默认值为0。
	MaxIdleTime int

	// 提示桌面用户显式地接受或拒绝传入连接。默认是off。
	// 为了支持QueryConnect, vncconfig(1)程序必须在桌面上运行。
	QueryConnect bool

	//拒绝连接之前显示接受连接对话框的秒数。默认是10。
	QueryConnectTimeout int

	// 只允许来自同一机器的连接。如果您使用SSH并且想要停止来自任何其他主机的非SSH连接，则此功能非常有用。
	Localhost bool

	// 配置调试日志设置。
	//Dest目前可以是stderr、stdout或syslog,
	//level介于0和100之间，100表示最详细的输出。
	//logname通常是*表示全部，但如果您知道其“LogWriter”的名称，则可以将目标指向特定的源文件。
	//默认是*:stderr: 30。
	// logname:dest:level
	Log string

	// 更改键盘字符的映射
	// Mapping是一个逗号分隔的字符映射字符串，每个字符的形式为char->char或char<>char，其中char是一个十六进制键符号。例如，要交换"和@符号，你需要指定以下参数:
	// RemapKeys=0x22<>0x40
	RemapKeys string

	// 受NumLock影响的键通常需要插入一个假的Shift，以便生成正确的符号。
	// 打开这个选项可以避免这些额外的假Shift事件，但可能会导致一个稍微不同的符号(例如，一个回车而不是键盘上的回车)。
	AvoidShiftNumLock bool

	// 直接发送键盘事件，避免将它们映射到当前的键盘布局。这有效地使键盘根据服务器上配置的布局而不是客户端上配置的布局表现。默认是off。
	RawKeyboard bool

	// 逗号分隔的参数列表，可以使用VNC扩展修改。例如，可以在正在运行的会话中使用vncconfig(1)程序修改参数。
	// 如果Xvnc在不同的用户下运行，而不是在允许覆盖参数的程序下运行，那么允许覆盖参数(如PAMService或PasswordFile)会对安全性产生负面影响。
	// 当设置NoClipboard参数时，允许覆盖SendCutText和AcceptCutText不起作用。
	//默认的是desktop、AcceptPointerEvents SendCutText AcceptCutText, SendPrimary SetPrimary。
	AllowOverride string
}

func NewXVncOpts() *XVncOpts {
	opts := &XVncOpts{}
	opts.Desktop = "x11"
	opts.Geometry = "1024x768"
	opts.Depth = 24
	opts.PixelFormat = "RGB888"
	opts.RfbWait = 30000
	opts.UseIPv4 = true
	opts.UseIPv6 = false
	opts.AcceptCutText = false
	opts.SendCutText = false
	opts.SendPrimary = true
	opts.AcceptPointerEvents = true
	opts.AcceptKeyEvents = true
	opts.AcceptSetDesktopSize = true
	opts.DisconnectClients = true
	opts.NeverShared = false
	opts.Protocol33 = false
	opts.FrameRate = 60
	opts.CompareFB = 2
	opts.ZlibLevel = 3
	opts.ImprovedHextile = true
	opts.SecurityTypes = "VncAuth,TLSVnc"
	opts.UseBlacklist = true
	opts.BlacklistThreshold = 5
	opts.BlacklistTimeout = 10
	opts.IdleTimeout = 0
	opts.MaxDisconnectionTime = 0
	opts.MaxConnectionTime = 0
	opts.MaxIdleTime = 0
	opts.QueryConnect = false
	opts.QueryConnectTimeout = 10
	opts.Localhost = true
	return opts
}

// 判断分辨率和色深是否符合规则
func (that *XVncOpts) checkGeometryAndDepth() error {
	result, err := gregex.Match(`^(\d+)x(\d+)$`, gconv.Bytes(that.Geometry))
	if err != nil {
		return err
	}
	if len(result) != 3 {
		return gerror.Newf("%s 规则不正确", that.Geometry)
	}
	width := gconv.Int(result[1])
	height := gconv.Int(result[2])
	if width < 1 || height < 1 {
		return gerror.Newf("geometry %s is invalid", that.Geometry)
	}
	that.Geometry = fmt.Sprintf("%dx%d", width, height)
	if that.Depth < 8 || that.Depth > 32 {
		return gerror.New("Depth must be between 8 and 32")
	}
	return nil
}

func (that *XVncOpts) Check() error {
	err := that.checkGeometryAndDepth()
	if err != nil {
		return err
	}
	return nil
}

func (that *XVncOpts) Array() []string {
	var args = gmap.NewStrStrMap()
	args.Set("geometry", that.Geometry)
	args.Set("depth", gconv.String(that.Depth))
	args.Set("pixelformat", that.PixelFormat)
	args.Set("desktop", that.Desktop)

	if len(that.Interface) > 0 {
		args.Set("interface", that.Interface)
	}
	if that.RfbPort > 0 {
		args.Set("rfbport", gconv.String(that.RfbPort))
	}
	if !that.UseIPv4 {
		args.Set("UseIPv4", "")
	}

	if !that.UseIPv6 {
		args.Set("UseIPv6", "")
	}

	if len(that.RfbUnixPath) > 0 {
		args.Set("rfbunixpath", that.RfbUnixPath)
	}

	if len(that.RfbUnixMode) > 0 {
		args.Set("rfbunixmode", that.RfbUnixMode)
	}

	if that.RfbWait > 0 {
		args.Set("rfbwait", gconv.String(that.RfbWait))
	}

	if len(that.RfbAuth) > 0 {
		args.Set("rfbauth", that.RfbAuth)
	}
	if !that.AcceptCutText {
		args.Set("AcceptCutText", "")
	}

	if that.MaxCutText > 0 {
		args.Set("MaxCutText", gconv.String(that.MaxCutText))
	}
	if !that.SendCutText {
		args.Set("SendCutText", "")
	}

	if !that.SendPrimary {
		args.Set("SendPrimary", "")
	}

	if !that.AcceptPointerEvents {
		args.Set("AcceptPointerEvents", "")
	}

	if !that.AcceptKeyEvents {
		args.Set("AcceptKeyEvents", "")
	}

	if !that.AcceptSetDesktopSize {
		args.Set("AcceptSetDesktopSize", "")
	}

	if !that.DisconnectClients {
		args.Set("DisconnectClients", "")
	}

	if that.NeverShared {
		args.Set("NeverShared", "")
	}

	if that.AlwaysShared {
		args.Set("AlwaysShared", "")
	}
	if that.Protocol33 {
		args.Set("Protocol3.3", "")
	}

	if that.FrameRate != 60 {
		args.Set("FrameRate", gconv.String(that.FrameRate))
	}

	if that.CompareFB != 2 {
		args.Set("CompareFB", gconv.String(that.CompareFB))
	}
	if that.ZlibLevel != 3 {
		args.Set("ZlibLevel", gconv.String(that.ZlibLevel))
	}
	if !that.ImprovedHextile {
		args.Set("ImprovedHextile", "")
	}
	if len(that.SecurityTypes) > 0 {
		args.Set("SecurityTypes", that.SecurityTypes)
	}
	if len(that.PlainUsers) > 0 {
		args.Set("PlainUsers", that.PlainUsers)
	}
	if len(that.PAMService) > 0 {
		args.Set("PAMService", that.PAMService)
	}
	if len(that.X509Cert) > 0 {
		args.Set("X509Cert", that.X509Cert)
	}
	if len(that.X509Key) > 0 {
		args.Set("X509Key", that.X509Key)
	}
	if len(that.GnuTLSPriority) > 0 {
		args.Set("GnuTLSPriority", that.GnuTLSPriority)
	}
	if !that.UseBlacklist {
		args.Set("UseBlacklist", "")
	}

	if that.BlacklistThreshold != 5 {
		args.Set("BlacklistThreshold", gconv.String(that.BlacklistThreshold))
	}
	if that.BlacklistTimeout != 10 {
		args.Set("BlacklistTimeout", gconv.String(that.BlacklistTimeout))
	}
	if that.IdleTimeout != 0 {
		args.Set("IdleTimeout", gconv.String(that.IdleTimeout))
	}
	if that.MaxDisconnectionTime != 0 {
		args.Set("MaxDisconnectionTime", gconv.String(that.MaxDisconnectionTime))
	}
	if that.MaxConnectionTime != 0 {
		args.Set("MaxConnectionTime", gconv.String(that.MaxConnectionTime))
	}
	if that.MaxIdleTime != 0 {
		args.Set("MaxIdleTime", gconv.String(that.MaxIdleTime))
	}
	if that.QueryConnect {
		args.Set("QueryConnect", "")
	}
	if that.QueryConnectTimeout != 10 {
		args.Set("QueryConnectTimeout", gconv.String(that.QueryConnectTimeout))
	}
	if that.Localhost {
		args.Set("localhost", "")
	}
	if len(that.Log) > 0 {
		args.Set("Log", that.Log)
	}
	if len(that.RemapKeys) > 0 {
		args.Set("RemapKeys", that.RemapKeys)
	}
	if that.AvoidShiftNumLock {
		args.Set("AvoidShiftNumLock", "")
	}
	if that.RawKeyboard {
		args.Set("RawKeyboard", "")
	}
	if len(that.AllowOverride) > 0 {
		args.Set("AllowOverride", that.AllowOverride)
	}
	var params []string
	for k, v := range args.Map() {
		params = append(params, "-"+k)
		if len(v) > 0 {
			params = append(params, v)
		}
	}
	return params
}
func (that *XVncOpts) String() string {
	params := that.Array()
	return gstr.Implode(" ", params)
}
