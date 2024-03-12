package util

import (
	"encoding/json"
	"fmt"
	"github.com/amocea/go-im-chat/defs"
	"io"
	"net/http"
	"strings"
)

const (
	CommonForm      = "application/x-www-form-urlencoded"
	MultiForm       = "multipart/form-data"
	ApplicationJson = "application/json"
)

// Validator 参数校验器/*
type Validator struct {
}

type IllegalArgErr struct {
	Path   string
	Arg    string
	Method string
}

func (i IllegalArgErr) Error() string {
	return fmt.Sprintf("Illegal Argument: [%s] {%s} ，缺失 (%s) 参数", i.Method, i.Path, i.Arg)
}

// Validate 参数校验处理
func (v *Validator) Validate(w http.ResponseWriter,
	r *http.Request,
	j2m interface{},
	params ...string) error {
	// 进行参数校验处理
	if len(params) == 0 {
		// 表示没有任何需要进行参数校验的处理
		return nil
	}

	var err error
	switch r.Method {
	case http.MethodGet:
		q := r.URL.Query()
		for _, p := range params {
			if q.Get(p) == "" {
				// 表示是零值
				err = &IllegalArgErr{r.URL.Path, p, r.Method}
				break
			}
		}
	// 说明参数是在请求路径后面的
	case http.MethodPost, http.MethodPut, http.MethodDelete:

		// 执行判断 判断 enctype 的类型是什么
		ct := r.Header.Get("Content-Type")

		switch {
		case strings.Contains(ct, CommonForm):
			// 解析请求体
			if err := r.ParseForm(); err != nil {
				// 表示解析失败
				err = &IllegalArgErr{r.URL.Path, "ParseForm() 解析失败", r.Method}
				return err
			}
			// 表示普通的 form 表单 直接获取即可
			for _, v := range params {
				if val := r.PostFormValue(v); val == "" {
					// 表示没有数据存在 即可以判断
					err = &IllegalArgErr{r.URL.Path, v, r.Method}
					break
				}
			}
		case strings.Contains(ct, MultiForm):
			// 即是 multipart/form-data 若是文件资源的话，无法判定，由另外的程序判定
			// 解析请求体 100MB 的限制
			if err := r.ParseMultipartForm(10 * (1 << 20)); err != nil {
				// 表示解析失败
				err = &IllegalArgErr{r.URL.Path, "ParseMultiForm() 解析失败", r.Method}
				break
			}
			vals := r.MultipartForm.Value
			for _, v := range params {
				if _, ok := vals[v]; !ok {
					// 表示没有数据存在 即可以判断
					err = &IllegalArgErr{r.URL.Path, v, r.Method}
					break
				}
			}
		case strings.Contains(ct, ApplicationJson):
			// ContentType: application/json 的数据 // 进行赋值 然后进行操作
			body, err := io.ReadAll(r.Body)
			if err != nil {
				err = &IllegalArgErr{r.URL.Path, "请求体数据读取失败", r.Method}
				break
			}

			m := Json2Map1(body)
			for _, p := range params {
				if v, ok := (m)[p]; !ok || v == nil {
					err = &IllegalArgErr{r.URL.Path, p, r.Method}
					return err
				}
			}

			if j2m != nil {
				// 进行输出
				if v, ok := j2m.(*map[string]interface{}); ok {
					// 说明是 map 类型
					*v = m
				} else {
					// 说明是 其他类型
					_ = json.Unmarshal(body, j2m)
				}
			}
		default:
			err = &IllegalArgErr{r.URL.Path, "非法的 Content-Type", r.Method}
		}
	case http.MethodOptions, http.MethodHead:
		// Option 请求通常表示的是跨域请求进行域名的确认
	}
	return err
}

// ErrOutput 将错误信息输出
func (v *Validator) ErrOutput(w http.ResponseWriter, err error) {
	defs.RespJSON(w, http.StatusBadRequest, -1, err.Error(), nil)
}
