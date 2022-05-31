package desktop

import (
	"github.com/gogf/gf/os/genv"
	"single-agent/pkg/customexec"
)

// 命令说明 http://linux.51yip.com/search/xset

type XSet struct{}

func NewXSet() *XSet {
	return &XSet{}
}

// B 打开和关闭电脑的嘟嘟的提示音，比如我们打开文件的是否，出错的时候发出的声音。但是听音乐还是可以照常听的
func (that *XSet) B(open bool) error {
	var b = "off"
	if open {
		b = "on"
	}
	cmd := customexec.Command("/usr/bin/xset", "b", b)
	cmd.Env = genv.All()
	cmd.SetUser()
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// BC -bc 关闭调试版本兼容机制
//     bc 打开调试版本兼容机制
func (that *XSet) BC(open bool) error {
	var b = "-bc"
	if open {
		b = "bc"
	}
	cmd := customexec.Command("/usr/bin/xset", b)
	cmd.Env = genv.All()
	cmd.SetUser()
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

//C c 控制键盘的按键声
//   关闭/打开
func (that *XSet) C(open bool) error {
	var c = "off"
	if open {
		c = "on"
	}
	cmd := customexec.Command("/usr/bin/xset", "c", c)
	cmd.Env = genv.All()
	cmd.SetUser()
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// 命令	                        描述
//xset s off	            禁用屏保清空
//xset s 3600 3600	        将清空时间设置到 1 小时
//xset -dpms	            关闭 DPMS
//xset s off -dpms	        禁用 DPMS 并阻止屏幕清空
//xset dpms force off	    立即关闭屏幕
//xset dpms force standby	待机界面
//xset dpms force suspend	休眠界面

// DpmsOpen 打开电源之星，主要用来省电的
func (that *XSet) DpmsOpen() error {
	cmd := customexec.Command("/usr/bin/xset", "+dpms")
	cmd.Env = genv.All()
	cmd.SetUser()
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// DpmsClose 关闭电源之星，不用省电
func (that *XSet) DpmsClose() error {
	cmd := customexec.Command("/usr/bin/xset", "-dpms")
	cmd.Env = genv.All()
	cmd.SetUser()
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// FP 字体搜索
// 增加一个字体搜寻目录。
// 删除一个字体搜寻目录。
// 重新设置字体搜寻目录。
// $ xset +fp /usr/local/fonts/Type1
// $ xset fp+ /usr/local/fonts/bitmap
func (that *XSet) FP(dir string, opt int) error {
	fp := "fp="
	if opt == 0 {
		fp = "fp="
	} else if opt > 0 {
		fp = "fp+"
	} else if opt < 0 {
		fp = "fp-"
	}
	cmd := customexec.Command("/usr/bin/xset", fp, dir)
	cmd.Env = genv.All()
	cmd.SetUser()
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// NoBlank 屏保后画面是一个图案)
func (that *XSet) NoBlank() error {
	cmd := customexec.Command("/usr/bin/xset", "s", "noblank")
	cmd.Env = genv.All()
	cmd.SetUser()
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// Blank 屏保后画面为黑色的
func (that *XSet) Blank() error {
	cmd := customexec.Command("/usr/bin/xset", "s", "blank")
	cmd.Env = genv.All()
	cmd.SetUser()
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// SOff 禁用屏保清空
func (that *XSet) SOff() error {
	cmd := customexec.Command("/usr/bin/xset", "s", "off")
	cmd.Env = genv.All()
	cmd.SetUser()
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
