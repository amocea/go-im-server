package defs

import (
	"encoding/json"
	"log"
	"net/http"
)

// Response 响应实体类对象
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}
type PageResponse struct {
	Code  int         `json:"code"`
	Rows  interface{} `json:"rows"`
	Total int         `json:"total"`
}

// RespJSON 将数据以 JSON 的方式返回
func RespJSON(w http.ResponseWriter, sc int, code int, msg string, data interface{}) {
	r := Response{
		Code: code,
		Msg:  msg,
		Data: data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(sc)

	if v, err := json.Marshal(r); err != nil {
		log.Fatalln(err)
	} else {
		_, _ = w.Write(v)
	}
}

// RespPageJSON 将数据以 JSON 的方式返回
func RespPageJSON(w http.ResponseWriter, sc int, code int, total int, rows interface{}) {
	r := PageResponse{
		Code:  code,
		Rows:  rows,
		Total: total,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(sc)

	if v, err := json.Marshal(r); err != nil {
		log.Fatalln(err)
	} else {
		_, _ = w.Write(v)
	}
}
