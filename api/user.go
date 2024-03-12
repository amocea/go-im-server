package api

import (
	"errors"
	"fmt"
	"github.com/amocea/go-im-chat/config"
	"github.com/amocea/go-im-chat/defs"
	"github.com/amocea/go-im-chat/service"
	"github.com/amocea/go-im-chat/util"
	"log"
	"math/rand"
	"net/http"
)

var us *service.UserService

func init() {
	us = new(service.UserService)
	us.DB = config.DB() // 执行 db 的赋值
}

// SetRequestMethod 设置 api 的请求方法
func SetRequestMethod(method string, fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := r.Method
		if m != method {
			// 表示方法错误
			defs.RespJSON(w, http.StatusMethodNotAllowed, -1, "请求方法不支持", nil)
			return
		}
		fn(w, r) // 执行底层逻辑
	}
}

type RouterGroup struct {
	Prefix string
}

func Group(r string) *RouterGroup {
	return &RouterGroup{
		Prefix: r,
	}
}

func (r *RouterGroup) Register(method, url string, handler http.HandlerFunc) {
	u := fmt.Sprintf("%s%s", r.Prefix, url)
	log.Printf("Register url: -> [%s] url=[%s]", method, u)
	http.HandleFunc(u, SetRequestMethod(method, handler))
}

var RegisterUserHandler = func() {
	r := Group("/user")
	{
		r.Register(http.MethodPost, "/login", Login)
		r.Register(http.MethodPost, "/register", Register)
		r.Register(http.MethodPost, "/find", FindById)
	}
}

// Login 用户登录逻辑
func Login(w http.ResponseWriter, r *http.Request) {
	if err := us.Validate(w, r, nil, "mobile", "passwd"); err != nil {
		// 表示参数校验出问题
		us.ErrOutput(w, err) // 进行输出
		return
	}
	if err := r.ParseForm(); err != nil {
		log.Println(err)
		defs.RespJSON(w, http.StatusBadRequest, -1, "非法请求", nil)
		return
	}
	m, pwd := r.PostFormValue("mobile"), r.PostFormValue("passwd")

	user, err := us.Login(m, pwd)
	if err != nil {
		if errors.Is(err, service.ErrUserNotExist) {
			// 用户信息不存在
			defs.RespJSON(w, http.StatusNotFound, -1, "该手机号码未注册", nil)
		} else if errors.Is(err, service.ErrPassword) {
			// 密码错误
			defs.RespJSON(w, http.StatusUnauthorized, -1, "密码错误", nil)
		} else {
			// 内部错误
			defs.RespJSON(w, http.StatusInternalServerError, -1, "内部错误", nil)
		}
		return
	}

	defs.RespJSON(w, http.StatusOK, 0, "登录成功", user)
}

// Register 注册行为
func Register(w http.ResponseWriter, r *http.Request) {
	if err := us.Validate(w, r, nil, "mobile", "passwd"); err != nil {
		// 表示参数校验出问题
		us.ErrOutput(w, err) // 进行输出
		return
	}
	if err := r.ParseForm(); err != nil {
		log.Println(err)
		defs.RespJSON(w, http.StatusBadRequest, -1, "非法请求", nil)
		return
	}

	// 获取表单数据
	mobile := r.PostFormValue("mobile")
	passwd := r.PostFormValue("passwd")

	nickname := fmt.Sprintf("user%-6d", rand.Intn(100000))

	user, err := us.Register(mobile, passwd, nickname, defs.DefaultAvatar, defs.SexUnknown)
	if err != nil {
		if errors.Is(err, service.ErrUserExist) {
			// 表示用户存在
			defs.RespJSON(w, http.StatusOK, -1, err.Error(), nil)
		} else {
			// 内部错误
			defs.RespJSON(w, http.StatusInternalServerError, -1, "注册失败", nil)
		}
		return
	}

	// 没有错误
	defs.RespJSON(w, http.StatusCreated, 0, "注册成功", user)
}

// FindById 根据 id 寻找用户信息
func FindById(w http.ResponseWriter, r *http.Request) {
	var cond defs.User
	if err := us.Validate(w, r, &cond, "id"); err != nil {
		us.ErrOutput(w, err)
		return
	}

	// 执行绑定
	if err := util.Bind(r, &cond); err != nil {
		defs.RespJSON(w, http.StatusInternalServerError, -1, "内部错误", nil)
		return
	}

	u, err := us.FindById(cond.Id)
	if err != nil {
		defs.RespJSON(w, http.StatusInternalServerError, -1, "内部错误", nil)
		return
	}

	defs.RespJSON(w, http.StatusOK, 0, "查询成功", u)
}
