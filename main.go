package main

import (
	"github.com/amocea/go-im-chat/api"
	_ "github.com/amocea/go-im-chat/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
)

// IgnorePath 文件服务器访问文件时，一些文件不需要对外暴露
func IgnorePath(fs http.Handler, exts ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, ext := range exts {
			if strings.HasSuffix(r.URL.Path, ext) {
				// 表示带有路径
				http.NotFound(w, r)
				return
			}
		}
		fs.ServeHTTP(w, r)
	}
}

func main() {
	api.RegisterUserHandler()
	api.RegisterContactHandlers()
	api.RegisterChatHandlers()
	api.RegisterAttachHandler()

	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// api 服务器已打开
		// 下面的话仅允许访问 .html 后缀的
		// http.Handle("/", IgnorePath(http.FileServer(http.Dir("/templates")), []string{".go", ".mod"}...))
		// http.Dir(.) 表示将当前文件所在的目录作为静态文件服务器的根目录
		http.Handle("/asset/", http.FileServer(http.Dir(".")))
		http.Handle("/mnt/", http.FileServer(http.Dir(".")))
		log.Fatalln(http.ListenAndServe(":8080", nil))
	}()
	ei := api.GetEvictInterrupt()
	<-interrupt // 阻塞

	ei <- struct{}{}
	log.Println("websocket 淘汰监听程序正在关闭中...")
	runtime.Gosched()
	log.Println("服务器已关闭")
}
