package base62

import (
	"errors"
	"strings"
)

// 62 个可用字符字典（0-9, a-z, A-Z）
const base62Str = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Encode 将 10 进制无符号整数转换为 62 进制短码字符串
func Encode(num uint64) string {
	if num == 0 {
		return string(base62Str[0])
	}
	var sb strings.Builder
	for num > 0 {
		remainder := num % 62
		sb.WriteByte(base62Str[remainder])
		num /= 62
	}
	// 翻转字符串（因为取余数是从低位到高位）
	bytes := []byte(sb.String())
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}
	return string(bytes)
}

// Decode 将 62 进制短码字符串还原回 10 进制无符号整数
func Decode(str string) (uint64, error) {
	var num uint64
	for _, char := range str {
		idx := strings.IndexRune(base62Str, char)
		if idx == -1 {
			return 0, errors.New("invalid base62 character")
		}
		num = num*62 + uint64(idx)
	}
	return num, nil
}
