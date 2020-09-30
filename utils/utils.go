package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// GetMd5 .
func GetMd5(strData string) (string, error) {
	md5Ctx := md5.New()
	_, err := md5Ctx.Write([]byte(strData))
	if err != nil {
		return "", err
	}
	cipher := md5Ctx.Sum(nil)

	return hex.EncodeToString(cipher), nil
}
