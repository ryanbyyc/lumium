package auth

import (
	"net/http"
	"strings"

	"lumium/lib/lumnet"
)

// Register is the handler endpoint for creating a new user
//
// @Summary     Register
// @Description Create a new account. Success returns an access token in the body and sets a refresh-token cookie
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input  body  SignupDTO  true  "new account"
// @Success     200    {object}  ResultWire  "OK"
// @Header      200    {string}  Set-Cookie  "HttpOnly refresh token cookie (name & attributes per server config)"
// @Failure     400    {string}  string      "bad request / validation error"
// @Failure     409    {string}  string      "email already in use or tenant conflict"
// @Router      /auth/register [post]
func (h *Auth) Register(w http.ResponseWriter, r *http.Request) lumnet.Reply {
	in, err := lumnet.ParseJSON[SignupDTO](r)
	if err != nil {
		return lumnet.ErrorR(err)
	}

	res, err := h.svc.Signup(r.Context(), SignupInput{
		Email:      strings.ToLower(strings.TrimSpace(in.Email)),
		Password:   in.Password,
		Name:       strings.TrimSpace(in.Name),
		TenantSlug: strings.ToLower(strings.TrimSpace(in.TenantSlug)),
		UserAgent:  r.UserAgent(),
		IP:         clientIP(r),
	})
	if err != nil {
		return lumnet.ErrorR(err)
	}

	setRefreshCookie(w, h.svc.Config(), res.RefreshRaw)

	return lumnet.OKR(ResultWire{
		User: UserPublic{
			ID:              res.UserID,
			Email:           in.Email,
			Name:            in.Name,
			PrimaryTenantID: nullIfEmpty(res.TenantID),
		},
		AccessToken: res.Access,
		ExpiresIn:   res.ExpiresIn,
	})
}
