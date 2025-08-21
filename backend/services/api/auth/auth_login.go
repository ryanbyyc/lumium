package auth

import (
	"net/http"
	"strings"

	"lumium/lib/lumnet"

	lumErrors "lumium/lib/errors"
)

// Login is the handler endpoint for creating a JWT and authenticating with the app
//
// @Summary     Login
// @Description Authenticate with email/password. On success returns an access token in the body
// @Description and sets a refresh-token HttpOnly cookie.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input  body  LoginDTO  true  "credentials"
// @Success     200    {object}  ResultWire  "OK"
// @Header      200    {string}  Set-Cookie  "HttpOnly refresh token cookie (name & attributes per server config)"
// @Failure     400    {string}  string           "bad request / validation error"
// @Failure     401    {object}  ErrorWire    "invalid credentials"
// @Failure     423    {object}  MFALockedResponse "MFA required; complete challenge before retrying login"
// @Router      /auth/login [post]
func (h *Auth) Login(w http.ResponseWriter, r *http.Request) lumnet.Reply {
	in, err := lumnet.ParseJSON[LoginDTO](r)
	if err != nil {
		return lumnet.ErrorR(err)
	}

	res, mfa, err := h.svc.Login(r.Context(), LoginInput{
		Email:          strings.ToLower(strings.TrimSpace(in.Email)),
		Password:       in.Password,
		TenantID:       strings.TrimSpace(in.TenantID),
		MFAChallengeID: strings.TrimSpace(in.MFAChallengeID),
		MFACode:        strings.TrimSpace(in.MFACode),
		UserAgent:      r.UserAgent(),
		IP:             clientIP(r),
	})

	// MFA path: 423 with structured payload
	if mfa != nil && err == nil {
		return lumnet.JSONStatusR(map[string]any{
			"code":    "mfa_required",
			"message": "Additional verification required",
			"details": map[string]any{
				"challenge_id": mfa.ChallengeID,
				"factors":      mfa.Factors,
			},
		}, http.StatusLocked) // 423
	}

	if err != nil {
		// Map "invalid credentials" to 401 with simple top-level payload
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "invalid credentials") ||
			lumErrors.IsErrorCode(err, lumErrors.ErrorCodeInvalidArgument) {
			return lumnet.JSONStatusR(map[string]any{
				"code":    "invalid_credentials",
				"message": "Invalid email or password.",
			}, http.StatusUnauthorized) // 401
		}
		// Fallback to standard error envelope for everything else
		return lumnet.ErrorR(err)
	}

	setRefreshCookie(w, h.svc.Config(), res.RefreshRaw)
	return lumnet.OKR(ResultWire{
		User: UserPublic{
			ID:              res.UserID,
			Email:           in.Email,
			PrimaryTenantID: nullIfEmpty(res.TenantID),
		},
		AccessToken: res.Access,
		ExpiresIn:   res.ExpiresIn,
	})
}

func setRefreshCookie(w http.ResponseWriter, cfg Config, token string) {
	c := &http.Cookie{
		Name:     cfg.RefreshCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.RefreshCookieSecure,
		SameSite: http.SameSiteStrictMode, // switch to Lax for cross-site flows
		MaxAge:   int(cfg.RefreshTTL.Seconds()),
	}
	http.SetCookie(w, c)
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func clientIP(r *http.Request) string {
	// Chi's middleware.RealIP enabled in the router, this header is reliable
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		if i := strings.IndexByte(ip, ','); i >= 0 {
			return strings.TrimSpace(ip[:i])
		}
		return strings.TrimSpace(ip)
	}
	host := r.RemoteAddr
	if i := strings.LastIndexByte(host, ':'); i >= 0 {
		return host[:i]
	}
	return host
}
