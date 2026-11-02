package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	memory      uint32 = 19 * 1024 // 19MiB (argon2 uses memory to calculate each hash, make large scale password guess expensive)
	iterations  uint32 = 2         // (argon2 perform how many pass over memory)
	parallelism uint8  = 1         // (how many lanes for computation)
	saltLength         = 16        // (random value added to password before hashing)
	keyLength   uint32 = 32        // 32 bytes = 256 bits (lenght of hash password)
)

func HashPassword(plainPassword string) (string, error) {

	// make a slice of bytes of salthLength(16)
	salt := make([]byte, saltLength)

	// fill the slice salt with random bytes
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(plainPassword),
		salt,
		iterations,
		memory,
		parallelism,
		keyLength,
	)

	encoded := fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory,
		iterations,
		parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

func VerifyPassword(plainPassword string, HashPassword string) bool {

	// split the hash password
	parts := strings.Split(HashPassword, "$")
	/*
		 * parts := []string{
		    "",
		    "argon2id",
		    "v=19",
		    "m=19456,t=2,p=1",
		    "SALT",
		    "HASH",
			}
		Why is the first element empty?

		Because our string starts with $.
	*/

	// verify the parts
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false
	}

	// mem -> memory, time -> iteration
	var mem, time uint32
	// threads -> parallelism
	var threads uint8

	n, err := fmt.Sscanf(
		parts[3],
		"m=%d,t=%d,p=%d",
		&mem,
		&time,
		&threads,
	)

	if err != nil || n != 3 {
		return false
	}

	if mem < 8*1024 || mem > 256*1024 || time == 0 || time > 10 || threads == 0 || threads > 10 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 16 || len(salt) > 64 {
		return false
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expectedHash) != int(keyLength) {
		return false
	}

	actualHash := argon2.IDKey(
		[]byte(plainPassword),
		salt,
		time,
		mem,
		threads,
		uint32(len(expectedHash)),
	)

	return subtle.ConstantTimeCompare(expectedHash, actualHash) == 1
}
