package main

import (
	"fmt"
	"github.com/amocea/go-im-chat/api"
	"log"
	"net/http"
	"strings"
	"text/template"
)

/*
完成对于模板文件的访问行为
*/

// TemplateCache 模板文件缓存结构
type TemplateCache struct {
	m map[string]*template.Template
}

// RegisterAll 注册模板引擎的注册行为
func (t *TemplateCache) RegisterAll() {
	tpl, err := template.ParseGlob("templates/**/*")
	if err != nil {
		log.Println("模板渲染失败")
		log.Fatalln(err)
	}
	r := api.Group("")
	for _, v := range tpl.Templates() {
		tn := v.Name() // 获得模板引擎的名字
		if !strings.HasPrefix(tn, "/") {
			tn = fmt.Sprintf("/%s", tn)
		}
		log.Printf("模板渲染：文件：[templates%s] -> url:[%s] \n", tn, tn)
		r.Register(http.MethodGet, tn, TemplateHandler(v, tn))
	}
}

func init() {
	tc := &TemplateCache{
		m: make(map[string]*template.Template),
	}
	tc.RegisterAll()
}

func TemplateHandler(t *template.Template, u string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := t.ExecuteTemplate(w, u, nil)
		if err != nil {
			http.Error(w, "页面渲染错误", http.StatusInternalServerError)
			log.Fatalln(err)
		}
	}
}
