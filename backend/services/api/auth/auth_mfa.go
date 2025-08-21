package auth

import (
	"net/http"

	"lumium/lib/lumnet"
)

// MFAChallenge starts a one-time MFA challenge for a user
// @Summary     Start MFA challenge
// @Description Issues a one-time MFA challenge for the user. Typically called after a 423 `mfa_required` login response
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input  body  MFAChallengeDTO  true  "user to challenge"
// @Success     200    {object}  MFAChallengeResult  "challenge metadata"
// @Failure     400    {string}  string              "bad request / validation error"
// @Failure     404    {string}  string              "user not found"
// @Router      /auth/mfa/challenge [post]
func (h *Auth) MFAChallenge(w http.ResponseWriter, r *http.Request) lumnet.Reply {
	in, err := lumnet.ParseJSON[MFAChallengeDTO](r)
	if err != nil {
		return lumnet.ErrorR(err)
	}
	res, err := h.svc.MFAChallenge(r.Context(), MFAChallengeInput(in))
	if err != nil {
		return lumnet.ErrorR(err)
	}
	return lumnet.OKR(res)
}

// MFAVerify verifies and consumes an MFA challenge code
// @Summary     Verify MFA code
// @Description Verifies the 6-digit code for a challenge
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input  body  MFAVerifyDTO  true  "verification payload"
// @Success     200    {object}  MFAVerifyOK       "verification status"
// @Failure     400    {string}  string            "bad request / validation error"
// @Failure     401    {object}  ErrorWire         "invalid or expired code"
// @Failure     410    {string}  string            "challenge expired"
// @Router      /auth/mfa/verify [post]
func (h *Auth) MFAVerify(w http.ResponseWriter, r *http.Request) lumnet.Reply {
	in, err := lumnet.ParseJSON[MFAVerifyDTO](r)
	if err != nil {
		return lumnet.ErrorR(err)
	}

	ok, err := h.svc.MFAVerify(r.Context(), MFAVerifyInput(in))
	if err != nil {
		// Keep your centralized error mapping; callers will still see a clean payload via lumnet.ErrorR.
		return lumnet.ErrorR(err)
	}
	return lumnet.OKR(MFAVerifyOK{OK: ok})
}

// MFAVerifyOK is a tiny success envelope for MFA verify
// swagger:model
type MFAVerifyOK struct {
	OK bool `json:"ok" example:"true"`
}
