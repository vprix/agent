package app

import (
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/genv"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/text/gstr"
)

type FileInfo struct {
	Path  string `json:"path"`
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	IsDir bool   `json:"isDir"`
}

// FileList 展示文件列表信息
func (that *ControllerApiV1) FileList(r *ghttp.Request) {
	rootDir := fmt.Sprintf("%s/Downloads", genv.Get("HOME", "/home/vprix-user"))
	path := r.GetString("path", "/")
	path = fmt.Sprintf("%s/%s", rootDir, gstr.TrimLeft(path, "/"))
	path = gfile.Abs(path)
	if gstr.Pos(path, rootDir) != 0 {
		FailJson(true, r, "path not found")
	}
	// 扫描目录
	paths, err := gfile.ScanDirFunc(path, "*", false, func(path string) string {
		if gstr.Pos(gfile.Basename(path), ".") == 0 {
			return ""
		}
		return path
	})
	if err != nil {
		FailJson(true, r, err.Error())
	}
	if len(paths) == 0 {
		SusJson(true, r, "ok", g.Map{})
	}
	// 迭代路径，生成指定的目录
	var FileInfos []*FileInfo
	for _, p := range paths {
		fileInfo := new(FileInfo)
		fileInfo.Path = gstr.Replace(p, rootDir, "")
		fInfo, e := gfile.Info(p)
		if e != nil {
			FailJson(true, r, e.Error())
		}
		fileInfo.IsDir = fInfo.IsDir()
		fileInfo.Name = fInfo.Name()
		fileInfo.Size = fInfo.Size()
		FileInfos = append(FileInfos, fileInfo)
	}
	SusJson(true, r, "ok", FileInfos)
}

// Download 下载文件
func (that *ControllerApiV1) Download(r *ghttp.Request) {
	rootDir := fmt.Sprintf("%s/Downloads", genv.Get("HOME", "/home/vprix-user"))
	path := r.GetString("path")
	path = fmt.Sprintf("%s/%s", rootDir, gstr.TrimLeft(path, "/"))
	path = gfile.Abs(path)
	if gstr.Pos(path, rootDir) != 0 {
		r.Response.WriteHeader(404)
		r.Exit()
	}
	if !gfile.Exists(path) || gfile.IsDir(path) {
		r.Response.WriteHeader(404)
		r.Exit()
	}

	r.Response.ServeFileDownload(path)
}
