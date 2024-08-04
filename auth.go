package main

import (
	"crypto/rand"
	"fmt"
	"slices"

	"golang.org/x/crypto/argon2"
)

const ArgonMem = 64 * 1024 /* Memory, in KiB, so this is 64MiB */
const SaltLength = 16

func generateRandomBytes(nbytes int) ([]byte, error) {
	buf := make([]byte, nbytes)
	n, err := rand.Reader.Read(buf)
	if err != nil {
		return buf, err
	}
	if n != nbytes {
		return buf, fmt.Errorf("failed to read enough bytes")
	}
	return buf, err
}

func generateSalt() ([]byte, error) {
	return generateRandomBytes(SaltLength)
}

func VerifyPassword(password string, key, salt []byte) bool {
	key2 := argon2.IDKey([]byte(password), salt, 1, ArgonMem, 4, 64)
	return slices.Equal(key, key2)
}

func CreatePassword(password string) (key []byte, salt []byte, err error) {
	salt, err = generateSalt()
	if err != nil { return }
	key = argon2.IDKey([]byte(password), salt, 1, ArgonMem, 4, 64)
	return
}

