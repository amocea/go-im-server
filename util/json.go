package util

import (
	"encoding/json"
	"log"
)

// Json2Map 将 json 数据转化为 map 集合
func Json2Map(j []byte) map[string]string {
	m := make(map[string]string)

	if err := json.Unmarshal(j, &m); err != nil {
		log.Println(err)
	}

	return m
}

func Json2Map1(j []byte) map[string]interface{} {
	m := make(map[string]interface{})

	if err := json.Unmarshal(j, &m); err != nil {
		log.Println(err)
	}

	return m
}
