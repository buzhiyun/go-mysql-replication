package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
)

func Base64UrlSafeEncode(source []byte) string {
	b64 := base64.StdEncoding.EncodeToString(source)
	//safeB64 := strings.Replace(string(b64), "/", "_", -1)
	//safeB64 = strings.Replace(safeB64, "+", "-", -1)
	//safeB64 = strings.Replace(safeB64, "=", "", -1)
	safeB64 := url.QueryEscape(b64)

	return safeB64
}

func GethmacSha256(data string, secret string) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	//return hex.EncodeToString(h.Sum(nil))
	//return Base64UrlSafeEncode(h.Sum(nil))
	return h.Sum(nil)
}
