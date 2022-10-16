package hash

import "golang.org/x/crypto/argon2"

const (
	// OWASP Password Storage Cheat Sheet
	// SRC: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
	// SRC: https://argon2-cffi.readthedocs.io/en/stable/parameters.html
	ARGON2ID_MEMORY      uint32 = 37 * 1024 // 37 MiB
	ARGON2ID_ITERATIONS  uint32 = 1
	ARGON2ID_PARALLELISM uint8  = 1
	ARGON2ID_HASH_LENGTH uint32 = 16
)

type HasherArgon2id struct {
}

func (HasherArgon2id) Hash(password string, salt []byte) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		ARGON2ID_ITERATIONS,
		ARGON2ID_MEMORY,
		ARGON2ID_PARALLELISM,
		ARGON2ID_HASH_LENGTH)
}
