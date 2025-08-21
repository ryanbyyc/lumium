package auth

// Service-layer contracts (not HTTP DTOs, not DB entities)

// LoginInput is the service contract for logging in
// swagger:model
type LoginInput struct {
	Email, Password, TenantID, MFAChallengeID, MFACode string
	UserAgent, IP                                      string
}

// LoginResult is the service contract for logging in results
// swagger:model
type LoginResult struct {
	UserID, TenantID string
	Access           string
	ExpiresIn        int
	RefreshRaw       string
}

// MFARequired is the service contract for MFA
// swagger:model
type MFARequired struct {
	ChallengeID string
	Factors     []string
}

// SignupInput is the service contract for Sign up
// swagger:model
type SignupInput struct {
	Email, Password, Name, TenantSlug string
	UserAgent, IP                     string
}

// SignupResult is the service contract for Sign up response
// swagger:model
type SignupResult struct {
	UserID, TenantID string
	Access           string
	ExpiresIn        int
	RefreshRaw       string
}

// RefreshInput is the service contract for refreshing a JWT
// swagger:model
type RefreshInput struct {
	RefreshOpaque string
	UserAgent     string
	IP            string
}

// RefreshResult is the service contract response for refreshing a JWT
// swagger:model
type RefreshResult struct {
	Access     string
	ExpiresIn  int
	RefreshRaw string
}

// MFAChallengeResult is the service contract response for MFA challenges
// swagger:model
type MFAChallengeResult struct {
	ChallengeID string
	Factors     []string
}

// ForgotInput is the service contract response for forgotten passwords
// swagger:model
type ForgotInput struct {
	Email string
	IP    string
}

// ResetInput is the service contract response for resetting a password
// swagger:model
type ResetInput struct {
	Token    string
	Password string
}

// AccessClaims is the stable, JWT-agnostic view other packages can depend on
// swagger:model
type AccessClaims struct {
	Sub      string   `json:"sub"`
	TenantID string   `json:"tenant_id,omitempty"`
	Roles    []string `json:"roles,omitempty"`
}

type mfaChallengeShape struct {
	UserID string `json:"user_id" validate:"required,uuid4"`
}
type mfaVerifyShape struct {
	ChallengeID string `json:"challenge_id" validate:"required,uuid4"`
	Code        string `json:"code"         validate:"required,len=6,numeric"`
}

// MFAChallengeInput is alias for MFA challenge shapes
// swagger:model
type MFAChallengeInput mfaChallengeShape

// MFAVerifyInput is alias for MFA verification shapes
// swagger:model
type MFAVerifyInput mfaVerifyShape

// MFALockedResponse documents the 423 response when extra verification is required.
// swagger:model
type MFALockedResponse struct {
	Code    string `json:"code" example:"mfa_required"`
	Message string `json:"message" example:"Additional verification required"`
	Details struct {
		ChallengeID string   `json:"challenge_id" example:"ch_01JABCXYZ"`
		Factors     []string `json:"factors" example:"[\"totp\",\"email\"]"`
	} `json:"details"`
}

// ErrorWire documents simple auth errors
// swagger:model
type ErrorWire struct {
	// Machine-readable code
	Code string `json:"code" example:"invalid_credentials"`
	// Human-readable message
	Message string `json:"message" example:"Invalid email or password."`
}

// RefreshWire is the documented shape returned by /auth/refresh
// swagger:model
type RefreshWire struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}
