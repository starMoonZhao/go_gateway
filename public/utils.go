package public

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
)

// 生成加盐密码
func EncodeSaltPassword(salt, password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	//转换为十六进制
	hexPassord := fmt.Sprintf("%x", hash.Sum(nil))
	//加盐
	hash = sha256.New()
	hash.Write([]byte(hexPassord + salt))
	hexPassord = fmt.Sprintf("%x", hash.Sum(nil))
	return hexPassord
}

// MD5 md5加密
func MD5(s string) string {
	h := md5.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}
