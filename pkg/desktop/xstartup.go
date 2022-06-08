package desktop

import (
	"agent/pkg/customexec"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/util/gconv"
)

type XStartup struct {
	display    string
	vncDesktop string
}

func NewXStartup() *XStartup {
	return &XStartup{}
}

func (that *XStartup) Display() string {
	return that.display
}

func (that *XStartup) SetDisplay(display string) {
	that.display = display
}

func (that *XStartup) VncDesktop() string {
	return that.vncDesktop
}

func (that *XStartup) SetVncDesktop(vncDesktop string) {
	that.vncDesktop = vncDesktop
}

func (that *XStartup) InitEnv() {
	_ = genv.Remove("SESSION_MANAGER")
	_ = genv.Remove("DBUS_SESSION_BUS_ADDRESS")
	_ = genv.Set("XDG_SESSION_TYPE", "x11")
	_ = genv.Set("XKL_XMODMAP_DISABLE", gconv.String(1))
	_ = genv.Set("DISPLAY", that.display)
	_ = genv.Set("VNCDESKTOP", that.vncDesktop)
}

func (that *XStartup) VncConfig() error {
	cmd := customexec.Command("/usr/bin/vncconfig", "-nowin")
	cmd.Env = genv.All()
	cmd.SetUser()
	err := cmd.Start()
	if err != nil {
		return err
	}
	return nil
}

func (that *XStartup) DbusLaunch(session string) error {
	cmd := customexec.Command("/usr/bin/dbus-launch", session)
	cmd.Env = genv.All()
	cmd.SetUser()
	err := cmd.Start()
	if err != nil {
		return err
	}
	return nil
}
