package util

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type ErrBindFailed struct {
	Err error
	Msg string
}

func (e ErrBindFailed) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return e.Err.Error()
}
func NewErr(msg string, err error) ErrBindFailed {
	return ErrBindFailed{
		Err: err,
		Msg: msg,
	}
}

// Bind 将 r 的请求数据绑定到结构体上
func Bind(r *http.Request, target interface{}) error {
	if !IsPointer(target) {
		return NewErr("参数非法：非指针对象", nil)
	}
	// 进行操作
	switch r.Method {
	case http.MethodGet:
		// 读取的是请求路径后的参数内容
		return BindUrlArg(r, target)
	case http.MethodPost, http.MethodDelete, http.MethodPut:
		// 读取的是请求体的内容
		ct := r.Header.Get("Content-Type")
		if strings.Contains(ct, CommonForm) || strings.Contains(ct, MultiForm) {
			// 表示 Content-Type = application/x-www-form-urlencoded | multipart/form-data
			return BindForm(r, ct, target)
		} else if strings.Contains(ct, ApplicationJson) {
			// 表示 Content-Type = application/json
			return BindJSON(r, target)
		} else {
			// 非法的 Content-Type
			return NewErr("参数非法：非法的 Content-Type", nil)
		}
	}
	return nil
}

// BindUrlArg 绑定请求参数
func BindUrlArg(r *http.Request, target interface{}) error {
	q := r.URL.Query()
	if len(q) == 0 {
		return nil // 不需要绑定
	}
	return _bind(q, target)
}

// BindJSON 绑定 JSON 数据
func BindJSON(r *http.Request, target interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return NewErr("读取请求体失败", err)
	}
	if len(body) == 0 {
		// 表示没有任何数据
		return nil
	}

	// 进行解析
	if err := json.Unmarshal(body, target); err != nil {
		return NewErr("", err)
	}
	return nil
}

func _bind(vals url.Values, target interface{}) (err error) {
	// 获取底层使用对象的值和数据类型
	v := reflect.ValueOf(target).Elem()
	t := reflect.TypeOf(target).Elem()
	// 进行赋值行为
	for i := 0; i < t.NumField(); i++ {
		// 获取到属性值
		f := v.Field(i)
		sf := t.Field(i)
		if !f.CanSet() {
			// 表示不允许设置
			continue
		}

		// 设置完毕后，获取 sf 的标签和 f 的类型
		typ := f.Kind()
		var k string
		if k = sf.Tag.Get("form"); k == "" {
			// 没有设置 form tag，那么将结构体属性名全小写
			k = PropertyName2ArgName(sf.Name)
		}
		val := vals.Get(k)
		if val == "" {
			// []string 没有数据
			continue
		}

		switch typ {
		case reflect.Slice:
			// 表示是数组对象
			var s []string
			var ok bool
			if s, ok = vals[k]; !ok || len(s) == 0 {
				// 表示没有存在数组对象 即返回即可
				break
			}
			// 执行赋值
			l := len(s)
			// 获取元素的类型
			etyp := sf.Type.Elem().Kind()
			tmp := reflect.MakeSlice(sf.Type, l, l)
			for i := 0; i < l; i++ {
				if err = Set(etyp, s[i], tmp.Index(i)); err != nil {
					break
				}
			}
		default:
			// 表示可能是另外的值
			// 有可能是 time.Time 类型
			if _, ok := f.Interface().(time.Time); ok {
				// 表示是 time.Time 类型
				err = SetTime(val, f, sf)
			} else {
				err = Set(typ, val, f)
				if err != nil && (err.(ErrBindFailed)).Msg == "Unknown type" {
					// 直接返回结束
					break
				}
			}
		}
		if err != nil {
			break
		}
	}
	return
}

// BindForm 绑定表单数据
func BindForm(r *http.Request, ct string, target interface{}) error {

	var v url.Values
	if strings.Contains(ct, MultiForm) {
		if err := r.ParseMultipartForm(10 * (1 << 20)); err != nil {
			// 获取 10MB 的内容
			return NewErr("", err)
		}
		v = r.MultipartForm.Value
	} else {
		if err := r.ParseForm(); err != nil {
			return NewErr("", err)
		}
		v = r.PostForm
	}
	return _bind(v, target)
}

// ArgName2propertyName 将请求参数的名字转化为符合 Go 规则的名称
// 这里的请求参数名称都是没有带上下划线的，即只需要将第一个字母改成大写即可
func ArgName2propertyName(an string) string {
	s := []byte(an)
	// 将首字母转化为大写形式
	if s[0] >= 65 && s[0] <= 90 {
		s[0] += 32
	}
	f := []byte{s[0]}
	s = append(f, s[1:]...)
	return string(s)
}

func PropertyName2ArgName(an string) string {
	return strings.ToLower(an)
}

// IsPointer 判断是否是指针对象
func IsPointer(target interface{}) bool {
	typ := reflect.TypeOf(target)
	return typ.Kind() == reflect.Pointer
}

func SetInt(val string, bitsize int, field reflect.Value) error {
	tmp, err := strconv.ParseInt(val, 10, bitsize)
	if err != nil {
		return NewErr("", err)
	}
	field.SetInt(tmp)
	return nil
}
func SetUInt(val string, bitsize int, field reflect.Value) error {
	tmp, err := strconv.ParseUint(val, 10, bitsize)
	if err != nil {
		return NewErr("", err)
	}
	field.SetUint(tmp)
	return nil
}

// SetBoolean 设置 boolean 类型
func SetBoolean(val string, field reflect.Value) error {
	tmp, err := strconv.ParseBool(val)
	if err != nil {
		return NewErr("", err)
	}
	field.SetBool(tmp)
	return nil
}

func SetFloat(val string, bitsize int, field reflect.Value) error {
	tmp, err := strconv.ParseFloat(val, bitsize)
	if err != nil {
		return NewErr("", err)
	}
	field.SetFloat(tmp)
	return nil
}

func Set(typ reflect.Kind, val string, f reflect.Value) (err error) {
	switch typ {
	case reflect.String:
		// 直接注入即可
		f.SetString(val) // 设置
	case reflect.Int:
		err = SetInt(val, 0, f)
	case reflect.Int8:
		err = SetInt(val, 8, f)
	case reflect.Int16:
		err = SetInt(val, 16, f)
	case reflect.Int32:
		err = SetInt(val, 32, f)
	case reflect.Int64:
		err = SetInt(val, 64, f)
	case reflect.Uint:
		err = SetUInt(val, 0, f)
	case reflect.Uint8:
		err = SetUInt(val, 8, f)
	case reflect.Uint16:
		err = SetUInt(val, 16, f)
	case reflect.Uint32:
		err = SetUInt(val, 32, f)
	case reflect.Uint64:
		err = SetUInt(val, 0, f)
	case reflect.Bool:
		err = SetBoolean(val, f)
	case reflect.Float32:
		err = SetFloat(val, 32, f)
	case reflect.Float64:
		err = SetFloat(val, 64, f)
	default:
		err = NewErr("Unknown type", nil)
	}
	return
}

func SetTime(val string, f reflect.Value, sf reflect.StructField) (err error) {
	// 获取时间格式 将格式传回即可
	tf := sf.Tag.Get("time_format")
	if tf == "" {
		tf = "2006-01-02 15:04:05"
	}

	t, err := time.Parse(tf, val)
	if err != nil {
		return NewErr("", err)
	}

	f.Set(reflect.ValueOf(t))
	return nil
}
