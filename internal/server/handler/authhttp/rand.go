package authhttp

import "crypto/rand"

// defaultRand 读 crypto/rand。
func defaultRand(p []byte) (int, error) { return rand.Read(p) }
