package game

import (
	"fmt"
	"math/rand"
)

var passwordChars = []rune("abcdefghijklmnopqrstuvwxyz" + "0123456789")

func genRandomPassword(n int) string {
	password := []rune{}
	for range n {
		password = append(password, passwordChars[rand.Intn(len(passwordChars))])
	}
	return string(password)
}

func (t *Team) QuickLoginURL() string {
	return fmt.Sprintf("/quick-login?l=%s&p=%s", t.Login, t.Password)
}
