package app

import "github.com/gogf/gf/os/genv"

const (
	VprixAgentLoginUsername = "VPRIX_AGENT_LOGIN_USERNAME"
	VprixAgentLoginPassword = "VPRIX_AGENT_LOGIN_PASSWORD"
	VprixAgentVncPassword   = "VPRIX_AGENT_VNC_PASSWORD"
)

func GetUserName() string {
	return genv.Get(VprixAgentLoginUsername, "vprix")
}

func GetPassword() string {
	return genv.Get(VprixAgentLoginPassword, "vprix")
}
func GetVncPassword() string {
	return genv.Get(VprixAgentVncPassword, "vprix.com")
}
