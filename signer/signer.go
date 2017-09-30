package signer

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
)

// PostPresignSignatureV1 - presigned signature for PostPolicy request.
func PostPresignSignatureV1(policyBase64, secretAccessKey string) string {
	hm := hmac.New(sha1.New, []byte(secretAccessKey))
	hm.Write([]byte(policyBase64))
	signature := base64.StdEncoding.EncodeToString(hm.Sum(nil))
	return signature
}
