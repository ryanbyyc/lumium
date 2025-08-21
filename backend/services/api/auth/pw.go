package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

var errInvalidPHC = errors.New("invalid argon2id PHC format")

type argon2idPHC struct {
	Version uint32
	MemKiB  uint32
	Iter    uint32
	Par     uint8
	Salt    []byte
	Key     []byte
}

// VerifyPassword compares a plaintext password to a PHC-formatted argon2id hash
func VerifyPassword(plain, phc string) (bool, error) {
	p, err := parsePHCArgon2id(phc)
	if err != nil {
		return false, err
	}
	got := deriveArgon2id(plain, p.Salt, p.Iter, p.MemKiB, p.Par, uint32(len(p.Key)))
	// (optionally zeroize `got` after use)
	return subtle.ConstantTimeCompare(got, p.Key) == 1, nil
}

// HashPassword produces a PHC-formatted argon2id hash using cfg parameters
func HashPassword(plain string, cfg Config) (string, error) {
	salt := make([]byte, cfg.ArgonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	key := deriveArgon2id(plain, salt, cfg.ArgonIter, cfg.ArgonMemKiB, cfg.ArgonParallel, cfg.ArgonKeyLen)
	p := &argon2idPHC{
		Version: 19,
		MemKiB:  cfg.ArgonMemKiB,
		Iter:    cfg.ArgonIter,
		Par:     cfg.ArgonParallel,
		Salt:    salt,
		Key:     key,
	}
	return formatPHCArgon2id(p), nil
}

func deriveArgon2id(plain string, salt []byte, iter, memKiB uint32, par uint8, keyLen uint32) []byte {
	return argon2.IDKey([]byte(plain), salt, iter, memKiB, par, keyLen)
}

func parsePHCArgon2id(phc string) (*argon2idPHC, error) {
	// Expected: $argon2id$v=19$m=65536,t=3,p=2$<saltB64>$<keyB64>
	parts := strings.Split(phc, "$")
	if len(parts) < 6 || parts[1] != "argon2id" {
		return nil, errInvalidPHC
	}

	ver, ok := strings.CutPrefix(parts[2], "v=")
	if !ok {
		return nil, errInvalidPHC
	}
	v, err := strconv.ParseUint(ver, 10, 32)
	if err != nil {
		return nil, errInvalidPHC
	}

	// m=...,t=...,p=...
	var m64, t64, p64 uint64
	for _, kv := range strings.Split(parts[3], ",") {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		switch k {
		case "m":
			m64, _ = strconv.ParseUint(v, 10, 32)
		case "t":
			t64, _ = strconv.ParseUint(v, 10, 32)
		case "p":
			p64, _ = strconv.ParseUint(v, 10, 8)
		}
	}
	if m64 == 0 || t64 == 0 || p64 == 0 {
		return nil, errInvalidPHC
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) == 0 {
		return nil, errInvalidPHC
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(key) == 0 {
		return nil, errInvalidPHC
	}

	return &argon2idPHC{
		Version: uint32(v),
		MemKiB:  uint32(m64),
		Iter:    uint32(t64),
		Par:     uint8(p64),
		Salt:    salt,
		Key:     key,
	}, nil
}

func formatPHCArgon2id(p *argon2idPHC) string {
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		p.Version, p.MemKiB, p.Iter, p.Par,
		base64.RawStdEncoding.EncodeToString(p.Salt),
		base64.RawStdEncoding.EncodeToString(p.Key),
	)
}
