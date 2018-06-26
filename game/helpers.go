package game

import (
	"crypto/rand"
	"encoding/base64"
)

func genRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func genRandomString(s int) (string, error) {
	b, err := genRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}
