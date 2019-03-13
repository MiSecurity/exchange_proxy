package util

import (
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"time"
)

// md5
func MD5(s string) (m string) {
	h := md5.New()
	_, _ = io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// 生成固定的随机字符串
func RandomString(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

// 生成激活码
func GenerateAtiveCode(username, deviceId string) (code string) {
	randString := RandomString(10)
	t := fmt.Sprintf("%v%v%v%v", username, deviceId, randString, time.Now().UnixNano())
	code = MD5(t)[5:17]
	return code

}
