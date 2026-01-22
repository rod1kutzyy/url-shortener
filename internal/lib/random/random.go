package random

import (
	"crypto/rand"
	"math/big"
)

func NewRandomString(size int) string {

	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	buf := make([]rune, size)
	for i := range buf {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			panic(err)
		}
		buf[i] = chars[n.Int64()]
	}

	return string(buf)
}
