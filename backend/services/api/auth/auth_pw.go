package auth

import (
	"net/http"
	"strings"

	"lumium/lib/lumnet"
)

// Forgot is the handler endpoint for initiating a forgot password request
//
// @Summary     Forgot password
// @Description Always returns 202 (Accepted) without revealing whether the email exists.
// @Description If the user exists, a reset token is generated and (normally) delivered out-of-band.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input  body  ForgotDTO  true  "email to reset"
// @Success     202    {object}  AcceptedWire  "accepted; no user enumeration"
// @Failure     400    {string}  string        "bad request / validation error"
// @Router      /auth/forgot [post]
func (h *Auth) Forgot(w http.ResponseWriter, r *http.Request) lumnet.Reply {
	in, err := lumnet.ParseJSON[ForgotDTO](r)
	if err != nil {
		return lumnet.ErrorR(err)
	}

	// Always respond 202; do best-effort side effect to avoid user enumeration
	_ = h.svc.Forgot(r.Context(), ForgotInput{
		Email: strings.ToLower(strings.TrimSpace(in.Email)),
		IP:    clientIP(r),
	})

	// Return an explicit 202 ack
	return lumnet.JSONStatusR(map[string]any{
		"code":    "accepted",
		"message": "If an account exists, you'll receive an email with instructions.",
	}, http.StatusAccepted)
}

// Reset is the handler endpoint for resetting a password
//
// @Summary     Reset password
// @Description Validates the reset token and sets a new password. On success clears the refresh cookie and returns 204.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input  body  ResetDTO  true  "reset token + new password"
// @Success     204    "password updated; no content"
// @Header      204    {string}  Set-Cookie  "clears refresh cookie"
// @Failure     400    {string}  string      "bad request / validation error"
// @Failure     422    {object}  ErrorWire   "invalid or expired token"
// @Router      /auth/reset [post]
func (h *Auth) Reset(w http.ResponseWriter, r *http.Request) lumnet.Reply {
	in, err := lumnet.ParseJSON[ResetDTO](r)
	if err != nil {
		return lumnet.ErrorR(err)
	}

	if err := h.svc.Reset(r.Context(), ResetInput{
		Token:    strings.TrimSpace(in.Token),
		Password: in.Password,
	}); err != nil {
		return lumnet.ErrorR(err)
	}

	clearRefreshCookie(w, h.svc.Config())
	return lumnet.NoContentR()
}
