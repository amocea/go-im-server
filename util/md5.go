package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

/*
关于 md5 加密的内容
*/

// Md5Encode 将明文密码 md5 加密为密文
func Md5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))

	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

// MD5Encode 调用 Md5Encode() 后再进行密文全大写处理
func MD5Encode(data string) string {
	return strings.ToUpper(Md5Encode(data))
}

// ValidatePassword 将明文 和 盐值 进行相加后跟密文字符进行比较
func ValidatePassword(plain, salt, epwd string) bool {
	return MD5Encode(plain+salt) == epwd
}

// EncryptPassword 将 plain + salt 加密后输出为加密密文
func EncryptPassword(plain, salt string) string {
	return MD5Encode(plain + salt)
}

// GenerateSalt 随机生成的盐值，盐值的选择：时间戳 + 选择数
func GenerateSalt() string {
	unix := int(time.Now().Unix()) // 进行转化
	r := rand.Intn(10000)
	return fmt.Sprintf("%d%5d", unix, r)
}
