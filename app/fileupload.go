package app

import (
	"bufio"
	"fmt"
	"github.com/gogf/gf/crypto/gmd5"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gfile"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

type FileUploadResponseInfo struct {
	IsUploaded int `json:"isUploaded"` // 是否已经完成上传了 0:否  1:是 - 秒传
	Merge      int `json:"merge"`      // 是否可以合并了   0：否  1:是
}

func (that *ControllerApiV1) Upload(r *ghttp.Request) {
	// 返回给前端的对象
	resultInfo := FileUploadResponseInfo{}
	// 接收，处理数据
	filename := r.GetString("filename")              // 文件名
	chunkNumber := r.GetInt("chunkNumber")           // 分片编号
	currentChunkSize := r.GetInt("currentChunkSize") // 当前分片的长度
	totalChunks := r.GetInt("totalChunks")           // 总的分片数量
	fileMd5 := r.GetString("identifier")             // 整个文件的md5值
	// 临时的文件名
	tempFileName := fileMd5 + "_" + strconv.Itoa(chunkNumber) + filepath.Ext(filename)

	// 上传文件前，需要做一次校验，使用get请求执行该操作
	// 传入文件的基本信息，服务端根据条件返回不同的操作逻辑个前端
	if r.Method == "GET" {
		// 如果文件长度等于单个分片长度，则告诉前端，可以合并该文件了
		if chunkNumber == totalChunks {
			resultInfo.IsUploaded = 0
			resultInfo.Merge = 1
			SusJson(true, r, "ok", resultInfo)
		}

		resultInfo.IsUploaded = 0
		resultInfo.Merge = 0
		SusJson(true, r, "ok", resultInfo)
	}
	// 保存当前分片
	err := saveChunkToLocalFromMultiPartForm(r, "/tmp", tempFileName, currentChunkSize)
	if err != nil {
		// 告诉前端重传
		r.Response.WriteHeader(500)
		resultInfo.IsUploaded = 1
		resultInfo.Merge = 0
		SusJson(true, r, "ok", resultInfo)
	}

	resultInfo.IsUploaded = 0
	resultInfo.Merge = 0
	SusJson(true, r, "ok", resultInfo)
	return
}

// Merge 合并文件
func (that *ControllerApiV1) Merge(r *ghttp.Request) {
	chunkNumber := r.GetInt("chunkNumber")
	fileMd5 := r.GetString("identifier")
	fileName := r.GetString("fileName")
	targetFileName := "/home/vprix-user/Upload/" + fileName
	f, err := os.OpenFile(targetFileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
	if err != nil {
		FailJson(true, r, fmt.Sprintf("创建合并文件[%s]失败", targetFileName))
	}
	var totalSize int64
	writer := bufio.NewWriter(f)
	for i := 1; i <= chunkNumber; i++ {
		currentChunkFile := "/tmp/" + fileMd5 + "_" + strconv.Itoa(i) + filepath.Ext(fileName) // 当前的分片名
		bytes, err := ioutil.ReadFile(currentChunkFile)
		if err != nil {
			FailJson(true, r, fmt.Sprintf("读取分片文件[%s]失败，err:%s", currentChunkFile, err.Error()))
		}
		num, err := writer.Write(bytes)
		if err != nil {
			FailJson(true, r, fmt.Sprintf("写入分片文件[%s]失败，err:%s", currentChunkFile, err.Error()))
		}
		totalSize += int64(num)
		err = os.Remove(currentChunkFile)
		if err != nil {

		}
	}
	err = writer.Flush()
	if err != nil {
		FailJson(true, r, fmt.Sprintf("写入全部分片失败,%s", err.Error()))
	}
	// 在重新打开文件之间关闭
	_ = f.Close()
	// 计算已合并文件的md5
	md5Str, err := gmd5.EncryptFile(targetFileName)
	if err != nil {
		FailJson(true, r, fmt.Sprintf("计算已合并文件的md5[%s]失败，err:%s", targetFileName, err))
	}
	if fileMd5 != md5Str {
		FailJson(true, r, "md5对比失败")
	}
	SusJson(true, r, "ok")
}

// 保存文件分片
func saveChunkToLocalFromMultiPartForm(r *ghttp.Request, tempDir, tempFileName string, currentChunkSize int) (err error) {
	if !gfile.IsDir(tempDir) {
		err = gfile.Mkdir(tempDir)
		if err != nil {
			return fmt.Errorf("创建临时文件 %s 失败,err:%v", tempDir, err)
		}
		err = gfile.Chmod(tempDir, 0766)
		if err != nil {
			return err
		}
	}
	fileHeader := r.Request.MultipartForm.File["file"][0]
	if fileHeader == nil {
		return fmt.Errorf("fileHeader 为空")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("error : %v", err)
	}
	// 创建临时文件
	myFile, err := gfile.Create(fmt.Sprintf("%s/%s", tempDir, tempFileName))
	if err != nil {
		return fmt.Errorf("error : %v", err)
	}
	// 读取从客户端传过来的文件内容
	buf := make([]byte, currentChunkSize)
	num, err := file.Read(buf)
	if err != nil {
		return fmt.Errorf("error : %v", err)
	}
	if num != currentChunkSize {
		return fmt.Errorf("接收的文件长度[%d]与传入的长度[%d]不一致", num, currentChunkSize)
	}
	num, err = myFile.Write(buf)
	if err != nil {
		return fmt.Errorf("error : %v", err)
	}
	// 关闭文件
	_ = myFile.Close()
	_ = file.Close()

	return nil
}
