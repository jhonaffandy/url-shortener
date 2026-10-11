package service

import (
	"crypto/rand"
	"math/big"
	"net/url"
)


const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"


func GenerateShortCode(length int) (string, error) {
	code := make([]byte, length)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[n.Int64()]
	}
	return string(code), nil
}


func IsValidURL(raw string) bool {
	u, err:= url.ParseRequestURI(raw)
	if err != nil{
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https"{
		return false
	}
	if u.Host == ""{
		return false
	}
	return true
}