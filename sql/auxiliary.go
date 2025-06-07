package sql

import (
	cr "crypto/rand"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"golang.org/x/exp/rand"
)

func generateKey() ([]byte, []byte) {

	privKey, pubKey, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, cr.Reader)
	if err != nil {
		panic(err)
	}
	privKeyByte, _ := crypto.MarshalPrivateKey(privKey)
	pubKeyByte, _ := crypto.MarshalPublicKey(pubKey)
	return privKeyByte, pubKeyByte

}

// The function generates a pseudo random string
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result []rune
	rand.Seed(uint64(time.Now().UnixNano()))

	for i := 0; i < length; i++ {
		randomIndex := rand.Intn(len(charset))
		result = append(result, rune(charset[randomIndex]))
	}
	return string(result)
}
