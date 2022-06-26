package app

import (
	"fmt"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/test/gtest"
	"github.com/gogf/gf/text/gstr"
	"testing"
)

func TestControllerApiV1_FileList(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		rootDir := fmt.Sprintf("%s/Downloads", genv.Get("HOME", "/home/vprix-user"))
		path := "/"
		path = fmt.Sprintf("%s/%s", rootDir, gstr.TrimLeft(path, "/"))
		path = gfile.Abs(path)
		t.Assert(gstr.Pos(path, rootDir), 0)
		paths, err := gfile.ScanDirFunc(path, "*", false, func(path string) string {
			if gstr.Pos(gfile.Basename(path), ".") == 0 {
				return ""
			}
			return path
		})

		t.Assert(err, nil)
		for _, path := range paths {
			t.Log(path)
			fileInfo, e := gfile.Info(path)
			t.Assert(e, nil)
			t.Log(fileInfo)
		}
	})
}
