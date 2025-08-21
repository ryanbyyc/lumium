package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	lumErrors "lumium/lib/errors"
)

// MFAChallenge issues a new MFA challenge (6-digit), store hash, return challenge id + advertised factors.
func (s *svc) MFAChallenge(ctx context.Context, in MFAChallengeInput) (*MFAChallengeResult, error) {
	userID := strings.TrimSpace(in.UserID)
	if userID == "" {
		return nil, lumErrors.InvalidArgf("user_id required")
	}

	code := random6()
	sum := sha256.Sum256([]byte(code))
	hash := hex.EncodeToString(sum[:])

	chID, err := s.Repo.CreateMFAChallenge(ctx, s.DB, userID, 10*time.Minute, hash)
	if err != nil {
		return nil, lumErrors.DBf("create challenge")
	}

	// TODO: deliver `code` via email/SMS/TOTP; never return it to the client.
	_ = code

	return &MFAChallengeResult{
		ChallengeID: chID,
		Factors:     []string{"email"}, // future: include sms/totp if present
	}, nil
}

// MFAVerify verifies an MFA challenge code atomically (increments attempts, fulfills on match).
func (s *svc) MFAVerify(ctx context.Context, in MFAVerifyInput) (bool, error) {
	chID := strings.TrimSpace(in.ChallengeID)
	code := strings.TrimSpace(in.Code)
	if chID == "" || code == "" {
		return false, lumErrors.InvalidArgf("invalid verification payload")
	}

	sum := sha256.Sum256([]byte(code))
	codeHash := hex.EncodeToString(sum[:])

	ok, _, err := s.Repo.VerifyAndConsumeMFA(ctx, s.DB, chID, codeHash)
	if err != nil {
		// invalid/expired challenge, or DB error
		return false, lumErrors.InvalidArgf("invalid or expired challenge")
	}
	if !ok {
		// auto-increment on mismatch is handled in repo.Update; here we just respond
		return false, lumErrors.InvalidArgf("invalid code")
	}
	return true, nil
}
