package auth

import (
	"time"

	"lumium/lib/config"
)

// Config is the configuration wrapper for authentication
type Config struct {
	JWTSecret           []byte
	JWTIssuer           string
	CoreMFAEnabled      bool
	AccessTTL           time.Duration
	RefreshTTL          time.Duration
	RefreshCookieName   string
	RefreshCookieSecure bool

	ArgonMemKiB   uint32
	ArgonIter     uint32
	ArgonParallel uint8
	ArgonSaltLen  uint32
	ArgonKeyLen   uint32
}

// LoadConfig returns the configuration wrapper for authentication
func LoadConfig() Config {
	c := Config{
		JWTSecret:           []byte(config.MustString("JWT_SECRET")),
		JWTIssuer:           config.MustString("JWT_ISSUER"),
		CoreMFAEnabled:      config.MayBool("CORE_MFA_ENABLED", false),
		AccessTTL:           time.Duration(config.MayInt("AUTH_ACCESS_TTL_SECONDS", 600)) * time.Second,
		RefreshTTL:          time.Duration(config.MayInt("AUTH_REFRESH_TTL_SECONDS", 30*24*60*60)) * time.Second,
		RefreshCookieName:   config.MayString("REFRESH_COOKIE_NAME", "refresh_token"),
		RefreshCookieSecure: config.MayBool("REFRESH_COOKIE_SECURE", true),

		ArgonMemKiB:   uint32(config.MayInt("ARGON2_MEM_KIB", 64*1024)),
		ArgonIter:     uint32(config.MayInt("ARGON2_ITER", 3)),
		ArgonParallel: uint8(config.MayInt("ARGON2_PAR", 1)),
		ArgonSaltLen:  uint32(config.MayInt("ARGON2_SALT_LEN", 16)),
		ArgonKeyLen:   uint32(config.MayInt("ARGON2_KEY_LEN", 32)),
	}
	normalizeArgon(&c)
	return c
}

func normalizeArgon(c *Config) {
	if c.ArgonMemKiB == 0 || c.ArgonMemKiB > 262144 { // cap at 256 MB, default 64 MB
		c.ArgonMemKiB = 64 * 1024
	}
	if c.ArgonIter == 0 {
		c.ArgonIter = 3
	}
	if c.ArgonParallel == 0 {
		c.ArgonParallel = 1 // safer default in containers
	}
	if c.ArgonSaltLen == 0 {
		c.ArgonSaltLen = 16
	}
	if c.ArgonKeyLen == 0 {
		c.ArgonKeyLen = 32
	}
}
