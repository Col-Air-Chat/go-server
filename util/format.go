package util

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"
)

func IsEmail(email string) bool {
	reg := `^([a-zA-Z0-9_-])+@([a-zA-Z0-9_-])+(.[a-zA-Z0-9_-])+`
	rgx := regexp.MustCompile(reg)
	return rgx.MatchString(email)
}

func IsPhone(phone string) bool {
	reg := `^1[0-9]{10}$`
	rgx := regexp.MustCompile(reg)
	return rgx.MatchString(phone)
}

func IsVaildLength(str string, min, max int) bool {
	reg := `^[\u4e00-\u9fa5a-zA-Z0-9_]{` + string(rune(min)) + `,` + string(rune(max)) + `}$`
	rgx := regexp.MustCompile(reg)
	return rgx.MatchString(str)
}

func Sha256(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func Int64ToString(num int64) string {
	return strconv.FormatInt(num, 10)
}

func StringToInt64(str string) int64 {
	num, _ := strconv.ParseInt(str, 10, 64)
	return num
}

func Float64ToString(num float64) string {
	return strconv.FormatFloat(num, 'f', -1, 64)
}
