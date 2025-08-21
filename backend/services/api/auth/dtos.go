package auth

// SignupDTO is the http data transfer object for registering a new user
// swagger:model
type SignupDTO struct {
	Email      string `json:"email"        validate:"required,email"`
	Password   string `json:"password"     validate:"required,min=8,max=128"`
	Name       string `json:"name,omitempty"        validate:"omitempty,max=120"`
	TenantSlug string `json:"tenant_slug,omitempty" validate:"omitempty,min=3,max=60"`
}

// LoginDTO is the http data transfer object for logging in
// swagger:model
type LoginDTO struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required"`
	TenantID       string `json:"tenant_id,omitempty" validate:"omitempty,uuid4" format:"uuid"`
	MFAChallengeID string `json:"mfa_challenge_id,omitempty" validate:"omitempty,uuid4" format:"uuid"`
	MFACode        string `json:"mfa_code,omitempty" validate:"omitempty,len=6,numeric"`
}

// UserPublic defines the data transfer object for users
// swagger:model
type UserPublic struct {
	ID              string  `json:"id"`
	Email           string  `json:"email,omitempty"`
	Name            string  `json:"name,omitempty"`
	PrimaryTenantID *string `json:"primary_tenant_id,omitempty"`
}

// MFARequiredWire defines the wire response for MFA
// swagger:model
type MFARequiredWire struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details MFARequiredDTO `json:"details"`
}

// MFARequiredDTO defines the data transfer object for MFA
// swagger:model
type MFARequiredDTO struct {
	ChallengeID string   `json:"challenge_id"`
	Factors     []string `json:"factors"`
}

// ForgotDTO defines the data transfer object for a user's forgotten password
// swagger:model
type ForgotDTO struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetDTO defines the data transfer object for a user's reset password flow
// swagger:model
type ResetDTO struct {
	Token    string `json:"token"    validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

// ResultWire defines the wire response for authentication
// swagger:model
type ResultWire struct {
	User        UserPublic `json:"user"`
	AccessToken string     `json:"access_token"`
	ExpiresIn   int        `json:"expires_in"`
}

// Result defines the standard response for authentication
// swagger:model
type Result struct {
	User        UserPublic `json:"user"`
	AccessToken string     `json:"access_token"`
	ExpiresIn   int        `json:"expires_in"`
	MFARequired *struct {
		ChallengeID string   `json:"challenge_id"`
		Factors     []string `json:"factors"`
	} `json:"mfa_required,omitempty"`
}

type (
	// MFAChallengeDTO defines the standard shape for MFA challenges
	// swagger:model
	MFAChallengeDTO mfaChallengeShape

	// MFAVerifyDTO defines the standard shape for MFA verifications
	// swagger:model
	MFAVerifyDTO mfaVerifyShape
)

// AcceptedWire is a small acknowledgement envelope for 202 responses.
// swagger:model
type AcceptedWire struct {
	Code    string `json:"code" example:"accepted"`
	Message string `json:"message" example:"If an account exists, you'll receive an email with instructions."`
}
