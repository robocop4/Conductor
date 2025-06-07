package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
)

// Load keys from file or generate new ones
func LoadKeyFromFile() (crypto.PrivKey, crypto.PubKey) {
	privKey, err := os.ReadFile("key.priv")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			privKey, pubKey := saveKeyToFile()
			return privKey, pubKey
		} else {
			panic(err)
		}
	}

	pubKey, err := os.ReadFile("key.pub")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			privKey, pubKey := saveKeyToFile()
			return privKey, pubKey
		} else {
			panic(err)
		}
	}
	privKeyByte, err := crypto.UnmarshalPrivateKey(privKey)
	if err != nil {
		panic(err)
	}
	pubKeyByte, err := crypto.UnmarshalPublicKey(pubKey)
	if err != nil {
		panic(err)
	}
	return privKeyByte, pubKeyByte
}

// Save generated keys to files
func saveKeyToFile() (crypto.PrivKey, crypto.PubKey) {
	fmt.Println("[+] Generate keys")
	privKey, pubKey, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		panic(err)
	}
	privKeyByte, _ := crypto.MarshalPrivateKey(privKey)
	pubKeyByte, _ := crypto.MarshalPublicKey(pubKey)
	err = os.WriteFile("key.priv", privKeyByte, 0644)
	if err != nil {
		// Handle error (optional)
	}
	err = os.WriteFile("key.pub", pubKeyByte, 0644)
	if err != nil {
		panic(err)
	}
	return privKey, pubKey
}
