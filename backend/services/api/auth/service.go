package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	lumErrors "lumium/lib/errors"
	"lumium/lib/store"
)

// Service defines the auth business operations exposed to HTTP handlers.
type Service interface {
	// Config returns the runtime configuration used by the service.
	Config() Config

	// Login authenticates credentials (and optional MFA), returns access & a rotated refresh token
	// If MFA is required and no code is provided, it returns an MFARequired response instead
	Login(ctx context.Context, in LoginInput) (*LoginResult, *MFARequired, error)

	// Signup creates a user (and optionally a tenant/membership), then returns tokens
	Signup(ctx context.Context, in SignupInput) (*SignupResult, error)

	// Refresh rotates the refresh token/session and mints a new access token
	Refresh(ctx context.Context, in RefreshInput) (*RefreshResult, error)

	// Logout revokes the provided refresh token (opaque) if present; idempotent
	Logout(ctx context.Context, refreshOpaque string) error

	// MFAChallenge issues a new MFA challenge for the user and returns the challenge metadata
	MFAChallenge(ctx context.Context, in MFAChallengeInput) (*MFAChallengeResult, error)

	// MFAVerify verifies and consumes an MFA challenge code
	MFAVerify(ctx context.Context, in MFAVerifyInput) (bool, error)

	// Forgot triggers a password-reset token flow (best-effort, non-enumerating)
	Forgot(ctx context.Context, in ForgotInput) error

	// Reset validates a reset token, updates the password, and revokes active sessions
	Reset(ctx context.Context, in ResetInput) error
}

// Login authenticates a user and handles MFA and session creation
func (s *svc) Login(
	ctx context.Context,
	in LoginInput,
) (*LoginResult, *MFARequired, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))

	userID, pwHash, active, err := s.Repo.GetUserByEmail(ctx, s.DB, email)
	if err != nil {
		_ = s.Repo.InsertLoginAttempt(
			ctx, s.DB, nil, email, false, "not_found", in.IP, in.UserAgent,
		)
		return nil, nil, lumErrors.InvalidArgf("invalid credentials")
	}

	if !active {
		_ = s.Repo.InsertLoginAttempt(
			ctx, s.DB, &userID, email, false, "inactive", in.IP, in.UserAgent,
		)
		return nil, nil, lumErrors.InvalidArgf("account disabled")
	}

	ok, _ := VerifyPassword(in.Password, pwHash)
	if !ok {
		_ = s.Repo.InsertLoginAttempt(
			ctx, s.DB, &userID, email, false, "invalid_password", in.IP, in.UserAgent,
		)
		return nil, nil, lumErrors.InvalidArgf("invalid credentials")
	}

	tenantID := strings.TrimSpace(in.TenantID)
	if tenantID == "" {
		tenantID, _ = s.Repo.GetPrimaryTenantID(ctx, s.DB, userID)
	}

	// MFA requirement: core OR tenant flag OR user has factor
	mfaNeeded := s.Cfg.CoreMFAEnabled
	if !mfaNeeded && tenantID != "" {
		if tReq, _ := s.Repo.TenantRequiresMFA(ctx, s.DB, tenantID); tReq {
			mfaNeeded = true
		}
	}
	if !mfaNeeded {
		if has, _ := s.Repo.UserHasMFAFactor(ctx, s.DB, userID); has {
			mfaNeeded = true
		}
	}

	if mfaNeeded && in.MFACode == "" {
		// Create challenge
		code := random6()
		sum := sha256.Sum256([]byte(code))
		codeHash := hex.EncodeToString(sum[:])

		chID, err := s.Repo.CreateMFAChallenge(ctx, s.DB, userID, 10*time.Minute, codeHash)
		if err == nil {
			_ = s.Repo.InsertLoginAttempt(
				ctx, s.DB, &userID, email, false, "mfa_required", in.IP, in.UserAgent,
			)
			// @TODO: deliver `code` via email/SMS
			_ = code
			return nil, &MFARequired{ChallengeID: chID, Factors: []string{"email"}}, nil
		}
	}

	if mfaNeeded && in.MFACode != "" {
		sum := sha256.Sum256([]byte(strings.TrimSpace(in.MFACode)))
		ok, _, _ := s.Repo.VerifyAndConsumeMFA(
			ctx,
			s.DB,
			strings.TrimSpace(in.MFAChallengeID),
			hex.EncodeToString(sum[:]),
		)
		if !ok {
			_ = s.Repo.InsertLoginAttempt(
				ctx, s.DB, &userID, email, false, "mfa_invalid", in.IP, in.UserAgent,
			)
			return nil, nil, lumErrors.InvalidArgf("invalid verification code")
		}
	}

	roles, _ := s.Repo.GetRolesForUserTenant(ctx, s.DB, userID, tenantID)

	access, exp, err := s.Cfg.MintAccess(userID, tenantID, roles)
	if err != nil {
		return nil, nil, lumErrors.DBf("mint access")
	}

	opaque, hash, err := NewOpaque(32)
	if err != nil {
		return nil, nil, lumErrors.DBf("refresh token")
	}
	if err := s.Repo.InsertSession(
		ctx, s.DB, userID, tenantID, hash, in.UserAgent, in.IP, s.Cfg.RefreshTTL,
	); err != nil {
		return nil, nil, lumErrors.DBf("create session")
	}

	_ = s.Repo.InsertLoginAttempt(
		ctx, s.DB, &userID, email, true, "ok", in.IP, in.UserAgent,
	)
	return &LoginResult{
		UserID:     userID,
		TenantID:   tenantID,
		Access:     access,
		ExpiresIn:  int(time.Until(exp).Seconds()),
		RefreshRaw: opaque,
	}, nil, nil
}

// Signup creates a user (and optionally a tenant & admin membership), then creates a session
// and mints an access token
func (s *svc) Signup(ctx context.Context, in SignupInput) (*SignupResult, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	name := strings.TrimSpace(in.Name)
	slug := strings.ToLower(strings.TrimSpace(in.TenantSlug))

	pwHash, err := HashPassword(in.Password, s.Cfg)
	if err != nil {
		return nil, lumErrors.DBf("hash password")
	}

	var userID, tenantID, access string
	var exp time.Time
	var refreshRaw, refreshHash string

	err = store.WithTx(ctx, s.DB, func(q store.Queryer) error {
		// Create user
		id, err := s.Repo.CreateUser(ctx, q, email, pwHash, name)
		if err != nil {
			if c := lumErrors.DBErrorCode(err); c != nil &&
				*c == lumErrors.ErrorCodeDuplicateKey {
				return lumErrors.WithField(
					lumErrors.DuplicateKeyf("email already registered"),
					"email",
				)
			}
			return lumErrors.DBf("create user: %v", err)
		}
		userID = id

		// Optional tenant setup
		if slug != "" {
			tid, err := s.Repo.EnsureTenantBySlug(ctx, q, slug)
			if err != nil {
				return lumErrors.DBf("ensure tenant: %v", err)
			}
			tenantID = tid

			if err := s.Repo.UpsertUserTenantAdmin(ctx, q, userID, tenantID); err != nil {
				return lumErrors.DBf("add membership")
			}
			_ = s.Repo.SetPrimaryTenantIfNull(ctx, q, userID, tenantID)
		}

		// Resolve tenant if still empty
		if tenantID == "" {
			tid, _ := s.Repo.GetPrimaryTenantID(ctx, q, userID)
			tenantID = tid
		}

		// Roles + access
		roles, _ := s.Repo.GetRolesForUserTenant(ctx, q, userID, tenantID)
		acc, e, err := s.Cfg.MintAccess(userID, tenantID, roles)
		if err != nil {
			return lumErrors.DBf("mint access")
		}
		access, exp = acc, e

		// Refresh session
		opaque, hash, err := NewOpaque(32)
		if err != nil {
			return lumErrors.DBf("refresh token")
		}
		refreshRaw, refreshHash = opaque, hash

		if err := s.Repo.InsertSession(
			ctx, q, userID, tenantID, refreshHash, in.UserAgent, in.IP, s.Cfg.RefreshTTL,
		); err != nil {
			return lumErrors.DBf("create session")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &SignupResult{
		UserID:     userID,
		TenantID:   tenantID,
		Access:     access,
		ExpiresIn:  int(time.Until(exp).Seconds()),
		RefreshRaw: refreshRaw,
	}, nil
}

// Refresh rotates the refresh session and returns a new access token and refresh token
func (s *svc) Refresh(ctx context.Context, in RefreshInput) (*RefreshResult, error) {
	sum := sha256.Sum256([]byte(in.RefreshOpaque))
	oldHash := hex.EncodeToString(sum[:])

	var userID, tenantID string
	var access string
	var exp time.Time
	var newOpaque string

	err := store.WithTx(ctx, s.DB, func(q store.Queryer) error {
		var err error
		userID, tenantID, err = s.Repo.GetActiveSessionByHash(ctx, q, oldHash)
		if err != nil {
			return lumErrors.InvalidArgf("unauthorized")
		}

		_ = s.Repo.RevokeSessionByHash(ctx, q, oldHash)

		opaque, newHash, err := NewOpaque(32)
		if err != nil {
			return lumErrors.DBf("opaque")
		}
		newOpaque = opaque

		if err := s.Repo.InsertSession(
			ctx, q, userID, tenantID, newHash, in.UserAgent, in.IP, s.Cfg.RefreshTTL,
		); err != nil {
			return lumErrors.DBf("insert new session")
		}

		roles, _ := s.Repo.GetRolesForUserTenant(ctx, q, userID, tenantID)
		acc, e, err := s.Cfg.MintAccess(userID, tenantID, roles)
		if err != nil {
			return lumErrors.DBf("mint access")
		}
		access, exp = acc, e

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &RefreshResult{
		Access:     access,
		ExpiresIn:  int(time.Until(exp).Seconds()),
		RefreshRaw: newOpaque,
	}, nil
}

// Logout revokes a refresh token by its opaque value
func (s *svc) Logout(ctx context.Context, refreshOpaque string) error {
	if strings.TrimSpace(refreshOpaque) == "" {
		return nil
	}
	sum := sha256.Sum256([]byte(refreshOpaque))
	hash := hex.EncodeToString(sum[:])
	_ = s.Repo.RevokeSessionByHash(ctx, s.DB, hash)
	return nil
}

// Forgot initiates a password reset by creating a one-time token if the user exists
// It does not reveal whether the email exists
func (s *svc) Forgot(ctx context.Context, in ForgotInput) error {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" {
		// Do nothing; handler always returns 202 to avoid enumeration
		return nil
	}

	// Look up user; if not found, quietly return
	uid, err := s.Repo.GetUserIDByEmail(ctx, s.DB, email)
	if err != nil {
		return nil
	}

	opaque, hash, err := NewOpaque(32)
	if err != nil {
		return nil
	}
	_ = s.Repo.InsertPasswordResetToken(ctx, s.DB, uid, hash, in.IP, 45*time.Minute) // 45m TTL

	// @TODO: deliver email containing the opaque token: /reset?token=<opaque>
	_ = opaque

	return nil
}

// Reset validates a reset token, sets a new password hash, consumes the token,
// revokes all active sessions for the user
func (s *svc) Reset(ctx context.Context, in ResetInput) error {
	token := strings.TrimSpace(in.Token)
	if token == "" {
		return lumErrors.InvalidArgf("invalid or expired token")
	}

	sum := sha256.Sum256([]byte(token))
	th := hex.EncodeToString(sum[:])

	return store.WithTx(ctx, s.DB, func(q store.Queryer) error {
		uid, err := s.Repo.LookupResetUserID(ctx, q, th)
		if err != nil {
			return lumErrors.InvalidArgf("invalid or expired token")
		}

		pwHash, err := HashPassword(in.Password, s.Cfg)
		if err != nil {
			return lumErrors.DBf("hash")
		}

		if err := s.Repo.UpdateUserPasswordHash(ctx, q, uid, pwHash); err != nil {
			return lumErrors.DBf("update password")
		}
		_ = s.Repo.MarkPasswordResetUsed(ctx, q, th)
		_ = s.Repo.RevokeAllSessionsForUser(ctx, q, uid)

		return nil
	})
}

// random6 returns a zero-padded 6-digit numeric code. It is suitable for OTP UX
// (the actual code is never stored; only a hash is persisted)
func random6() string {
	b := make([]byte, 2)
	_, _ = rand.Read(b)
	n := int(b[0])<<8 | int(b[1])
	v := 100000 + (n % 900000)
	return fmt.Sprintf("%06d", v)
}
