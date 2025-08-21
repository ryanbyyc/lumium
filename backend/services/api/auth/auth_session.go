package auth

import (
	"net/http"
	"strings"

	lumErrors "lumium/lib/errors"
	"lumium/lib/lumnet"
)

// Refresh is the http endpoint for refreshing a JWT
//
// @Summary     Refresh access token
// @Description Rotate the refresh session (from an HttpOnly cookie) and mint a new access token.
// @Description On success, returns a new access token in the body and sets a new refresh-token cookie.
// @Tags        auth
// @Produce     json
// @Success     200 {object}  RefreshWire  "OK"
// @Header      200 {string}  Set-Cookie   "New HttpOnly refresh token cookie (name & attributes per server config)"
// @Failure     422 {object}  ErrorWire    "unauthorized or invalid/expired refresh token"
// @Router      /auth/refresh [post]
func (h *Auth) Refresh(w http.ResponseWriter, r *http.Request) lumnet.Reply {
	cfg := h.svc.Config()
	c, err := r.Cookie(cfg.RefreshCookieName)
	if err != nil || c.Value == "" {
		return lumnet.ErrorR(lumErrors.InvalidArgf("unauthorized"))
	}

	res, err := h.svc.Refresh(r.Context(), RefreshInput{
		RefreshOpaque: c.Value,
		UserAgent:     r.UserAgent(),
		IP:            clientIP(r),
	})
	if err != nil {
		return lumnet.ErrorR(err)
	}

	setRefreshCookie(w, cfg, res.RefreshRaw)
	return lumnet.OKR(RefreshWire{
		AccessToken: res.Access,
		ExpiresIn:   res.ExpiresIn,
	})
}

// Logout is the http endpoint for destroying a session
//
// @Summary     Logout
// @Description Revoke the current refresh session and clear the token cookie.
// @Tags        auth
// @Produce     json
// @Success     204 {string}  string     "No Content"
// @Header      204 {string}  Set-Cookie "Clears refresh token cookie"
// @Router      /auth/logout [post]
func (h *Auth) Logout(w http.ResponseWriter, r *http.Request) lumnet.Reply {
	cfg := h.svc.Config()
	if c, err := r.Cookie(cfg.RefreshCookieName); err == nil && c.Value != "" {
		_ = h.svc.Logout(r.Context(), c.Value)
	}
	clearRefreshCookie(w, cfg)
	return lumnet.NoContentR()
}

// Me is the http endpoint for describing the current user
//
// @Summary     Current user
// @Description Return the current user derived from a Bearer access token
// @Tags        auth
// @Produce     json
// @Param       Authorization  header  string  true  "Bearer {access_token}"
// @Success     200 {object}   UserPublic
// @Failure     422 {object}   ErrorWire "unauthorized"
// @Router      /auth/me [get]
func (h *Auth) Me(w http.ResponseWriter, r *http.Request) lumnet.Reply {
	authz := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		return lumnet.ErrorR(lumErrors.InvalidArgf("unauthorized"))
	}
	raw := strings.TrimSpace(authz[7:])
	claims, err := h.svc.Config().ParseAccess(raw)
	if err != nil {
		return lumnet.ErrorR(lumErrors.InvalidArgf("unauthorized"))
	}
	return lumnet.OKR(UserPublic{
		ID:              claims.Sub,
		PrimaryTenantID: nullIfEmpty(claims.TenantID),
	})
}

func clearRefreshCookie(w http.ResponseWriter, cfg Config) {
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.RefreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.RefreshCookieSecure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}
