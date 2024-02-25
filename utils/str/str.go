package str

import (
	"crypto/rand"
)

var defaultAlphabet = []rune("_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func UUID() string {
	return NanoId(21)
}

func NanoId(size int) string {
	i := 0
re:
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		i++
		if i > 5 {
			panic(err)
		}
		goto re
	}
	id := make([]rune, size)
	for i := 0; i < size; i++ {
		id[i] = defaultAlphabet[bytes[i]&63]
	}
	return string(id[:size])
}
