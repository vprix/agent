package customexec

import (
	"github.com/osgochina/dmicro/logger"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

type Cmd struct {
	*exec.Cmd
}

func Command(name string, arg ...string) *Cmd {
	cmd := exec.Command(name, arg...)
	c := &Cmd{}
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	c.Cmd = cmd
	return c
}

func (that *Cmd) SetUser(username ...string) {
	userName := "vprix"
	if len(username) > 0 && len(username[0]) > 0 {
		userName = username[0]
	}
	//判断是否传入了用户组
	pos := strings.Index(userName, ":")
	groupName := ""
	if pos != -1 {
		groupName = userName[pos+1:]
		userName = userName[0:pos]
	}
	u, err := user.Lookup(userName)
	if err != nil {
		logger.Error(err)
		return
	}
	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		logger.Error(err)
		return
	}
	gid, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil && groupName == "" {
		logger.Error(err)
		return
	}
	if groupName != "" {
		g, err := user.LookupGroup(groupName)
		if err != nil {
			logger.Error(err)
			return
		}
		gid, err = strconv.ParseUint(g.Gid, 10, 32)
		if err != nil {
			logger.Error(err)
			return
		}
	}
	that.Cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid), NoSetGroups: true}
}

func (that *Cmd) Start() error {
	return that.Cmd.Start()
}
