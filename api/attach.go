package api

import (
	"fmt"
	"github.com/amocea/go-im-chat/defs"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

/*
上传文件
*/
func init() {
	wd, _ := os.Getwd()
	err := os.MkdirAll(fmt.Sprintf("%s\\%s", wd, "mnt"), os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
}

const (
	MAX_SIZE = 100 * (10 << 20)
)

var RegisterAttachHandler = func() {
	r := Group("/attach")
	{
		r.Register(http.MethodPost, "/upload", Upload)
	}
}

// Upload 上传文件 需要确保文件目录已经创建 mnt
func Upload(w http.ResponseWriter, r *http.Request) {
	_, fh, err := r.FormFile("file")
	uid := r.FormValue("userid")
	ftyp := r.FormValue("filetype")
	if err != nil {
		// 表示获取不到
		defs.RespJSON(w, http.StatusBadRequest, -1, "文件信息读取不到", nil)
		return
	}
	// 是否大于 100MB
	if fh.Size > MAX_SIZE {
		defs.RespJSON(w, http.StatusRequestEntityTooLarge, -1, "文件尺寸过大，超过 25 MB", nil)
		return
	}

	var ext string
	if ftyp != "" {
		ext = ftyp[1:]
	} else {
		// 获取到文件名 判断后缀名是否是我们想要的格式
		tmp := strings.Split(fh.Filename, ".")
		ext = tmp[len(tmp)-1]
	}

	if ext == "jpg" || ext == "png" || ext == "jpeg" || ext == "mp3" {
		// 说明是文件格式
		// 生成随机的名字
		nf := fmt.Sprintf("%s(%s1%d).%s", time.Now().Format("20060102150405"), uid, rand.Intn(100000), ext)
		f, err := fh.Open()
		defer f.Close()
		if err != nil {
			defs.RespJSON(w, http.StatusInternalServerError, -1, "文件上传失败", nil)
			return
		}
		nfi, err := os.Create(fmt.Sprintf("%s\\%s", ".\\mnt", nf))
		defer nfi.Close()
		if err != nil {
			defs.RespJSON(w, http.StatusInternalServerError, -1, "文件上传失败", nil)
			return
		}
		if _, err := io.Copy(nfi, f); err != nil {
			defs.RespJSON(w, http.StatusInternalServerError, -1, "文件上传失败", nil)
			return
		}

		defs.RespJSON(w, http.StatusCreated, 0, "文件上传成功", "/mnt/"+nf)
	} else {
		defs.RespJSON(w, http.StatusUnprocessableEntity, -1, "文件格式错误", nil)
	}
}
