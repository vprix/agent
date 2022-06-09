package env

import "github.com/gogf/gf/os/genv"

const (
	VprixAgentLoginUsername = "VPRIX_AGENT_LOGIN_USERNAME"
	VprixAgentLoginPassword = "VPRIX_AGENT_LOGIN_PASSWORD"
)

// UserName 获取网页端登录的用户名
func UserName() string {
	return genv.Get(VprixAgentLoginUsername, "vprix")
}

// Password 获取网页端登录的密码
func Password() string {
	return genv.Get(VprixAgentLoginPassword, "vprix")
}

// User 获取登录桌面的用户名
func User() string {
	return genv.Get("VPRIX_USER", "vprix-user")
}

// Home 获取家目录
func Home() string {
	return genv.Get("HOME", "/home/vprix-user")
}

// Display 显示端口
func Display() string {
	return genv.Get("DISPLAY", ":0")
}

// VprixPort 获取服务监听端口
func VprixPort() int {
	return genv.GetVar("VPRIX_PORT", 8080).Int()
}

// VprixAgentPath 获取agent的安装路径
func VprixAgentPath() string {
	return genv.Get("VPRIX_AGENT_PATH", "/usr/share/vprix_agent")
}

// VncColDepth 获取vnc 界面的色深
func VncColDepth() int {
	return genv.GetVar("VNC_COL_DEPTH", 24).Int()
}

// VncResolution 获取vnc的默认分辨率
func VncResolution() string {
	return genv.Get("VNC_RESOLUTION", "1280x1024")
}

// VncPassword 获取设置的vnc的密码，注意，最多只有8位起效
func VncPassword() string {
	return genv.Get("VNC_PASSWORD", "vprix.com")
}

// VncOptions 获取vnc的设置
func VncOptions() string {
	return genv.Get("VNC_OPTIONS", "")
}

func IsUbuntu() bool {
	return genv.GetVar("DISTRO", "ubuntu").String() == "ubuntu"
}
