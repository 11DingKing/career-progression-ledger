package auth

import "crypto/sha256"
import "encoding/hex"

func HashPassword(v string) string {
	h := sha256.Sum256([]byte("career-salt:" + v))
	return hex.EncodeToString(h[:])
}
func CheckPassword(hash, plain string) bool { return HashPassword(plain) == hash }
