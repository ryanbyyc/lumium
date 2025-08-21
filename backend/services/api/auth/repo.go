package auth

import (
	"context"
	"time"

	"lumium/lib/store"
)

// Repo is the auth data‐access interface. Implementations read/write authentication state
// (users, memberships, sessions, MFA, and password resets) via a store.Queryer.
type Repo interface {
	// GetUserByEmail returns (userID, passwordHash, isActive) for a normalized email.
	GetUserByEmail(ctx context.Context, q store.Queryer, email string) (
		userID, pwHash string, isActive bool, err error,
	)

	// GetPrimaryTenantID returns the user's primary tenant id (or "" if none).
	GetPrimaryTenantID(ctx context.Context, q store.Queryer, userID string) (
		tenantID string, err error)

	// GetRolesForUserTenant returns the role list for a user within a tenant. If tenantID is empty,
	// it returns roles across all memberships.
	GetRolesForUserTenant(
		ctx context.Context,
		q store.Queryer,
		uID string,
		tenantID string,
	) ([]string, error)

	// TenantRequiresMFA reports whether a tenant enforces MFA.
	TenantRequiresMFA(ctx context.Context, q store.Queryer, tenantID string) (bool, error)

	// UserHasMFAFactor reports whether the user has at least one MFA factor enrolled.
	UserHasMFAFactor(ctx context.Context, q store.Queryer, userID string) (bool, error)

	// CreateMFAChallenge creates a one-time MFA challenge with a hashed code and TTL.
	CreateMFAChallenge(
		ctx context.Context,
		q store.Queryer,
		userID string,
		ttl time.Duration,
		codeHash string,
	) (challengeID string, err error)

	// VerifyAndConsumeMFA atomically checks an MFA code hash, increments attempts on failure, and
	// marks the challenge fulfilled on success. Returns (ok, userID).
	VerifyAndConsumeMFA(
		ctx context.Context,
		q store.Queryer,
		challengeID string,
		codeHash string,
	) (ok bool, userID string, err error)

	// InsertSession writes a refresh session (hashed token) with UA/IP and expiry.
	InsertSession(
		ctx context.Context,
		q store.Queryer,
		userID string,
		tenantID string,
		refreshHash string,
		ua string,
		ip string,
		ttl time.Duration,
	) error

	// InsertLoginAttempt records a login attempt for auditing and lockout logic.
	InsertLoginAttempt(
		ctx context.Context,
		q store.Queryer,
		userID *string,
		email string,
		success bool,
		reason string,
		ip string,
		ua string,
	) error

	// CreateUser inserts a new user and returns its ID.
	CreateUser(ctx context.Context, q store.Queryer, email, pwHash, name string) (string, error)

	// EnsureTenantBySlug ensures a tenant exists for slug and returns its ID.
	EnsureTenantBySlug(ctx context.Context, q store.Queryer, slug string) (string, error)

	// UpsertUserTenantAdmin adds (or keeps) the user as admin in the tenant.
	UpsertUserTenantAdmin(ctx context.Context, q store.Queryer, userID, tenantID string) error

	// SetPrimaryTenantIfNull sets the user's primary tenant if it is currently NULL.
	SetPrimaryTenantIfNull(ctx context.Context, q store.Queryer, userID, tenantID string) error

	// GetActiveSessionByHash looks up an unrevoked, unexpired session by token hash.
	GetActiveSessionByHash(
		ctx context.Context,
		q store.Queryer,
		hash string,
	) (userID, tenantID string, err error)

	// RevokeSessionByHash marks a session revoked by token hash (idempotent).
	RevokeSessionByHash(ctx context.Context, q store.Queryer, hash string) error

	// GetUserIDByEmail returns a user ID for a normalized email.
	GetUserIDByEmail(ctx context.Context, q store.Queryer, email string) (string, error)

	// InsertPasswordResetToken stores a hashed reset token with IP and expiry.
	InsertPasswordResetToken(
		ctx context.Context,
		q store.Queryer,
		userID string,
		tokenHash string,
		requestedIP string,
		ttl time.Duration,
	) error

	// LookupResetUserID returns the user ID for a valid, unused reset token hash.
	LookupResetUserID(ctx context.Context, q store.Queryer, tokenHash string) (string, error)

	// UpdateUserPasswordHash replaces the user's password hash and bumps updated_at.
	UpdateUserPasswordHash(ctx context.Context, q store.Queryer, userID, newHash string) error

	// MarkPasswordResetUsed marks a reset token consumed.
	MarkPasswordResetUsed(ctx context.Context, q store.Queryer, tokenHash string) error

	// RevokeAllSessionsForUser revokes all active sessions for a user (post-reset).
	RevokeAllSessionsForUser(ctx context.Context, q store.Queryer, userID string) error
}

// GetUserByEmail returns (id, passwordHash, isActive) for the provided email.
func (r *repo) GetUserByEmail(
	ctx context.Context,
	q store.Queryer,
	email string,
) (string, string, bool, error) {
	var id, ph string
	var active bool
	err := q.QueryRow(
		ctx,
		`SELECT id::text, password_hash, is_active FROM users WHERE email = LOWER($1)`,
		email,
	).Scan(&id, &ph, &active)
	return id, ph, active, err
}

// GetPrimaryTenantID returns the primary tenant ID for a user (or "").
func (r *repo) GetPrimaryTenantID(
	ctx context.Context,
	q store.Queryer,
	userID string,
) (string, error) {
	var tid string
	err := q.QueryRow(
		ctx,
		`SELECT COALESCE(primary_tenant_id::text,'') FROM users WHERE id=$1`,
		userID,
	).Scan(&tid)
	return tid, err
}

// GetRolesForUserTenant returns the set of roles the user has in the given tenant (or across all
// if tenantID is empty).
func (r *repo) GetRolesForUserTenant(
	ctx context.Context,
	q store.Queryer,
	userID string,
	tenantID string,
) ([]string, error) {
	rows, err := q.Query(
		ctx,
		`SELECT role::text FROM users_tenants WHERE user_id=$1 AND ($2='' OR tenant_id::text=$2)`,
		userID,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var rr string
		_ = rows.Scan(&rr)
		roles = append(roles, rr)
	}
	return roles, nil
}

// TenantRequiresMFA reports whether MFA is enforced for the tenant.
func (r *repo) TenantRequiresMFA(
	ctx context.Context,
	q store.Queryer,
	tenantID string,
) (bool, error) {
	var f bool
	err := q.QueryRow(
		ctx,
		`SELECT COALESCE(mfa_required,false) FROM tenants WHERE id::text=$1`,
		tenantID,
	).Scan(&f)
	return f, err
}

// UserHasMFAFactor reports whether the user has any MFA factor enrolled.
func (r *repo) UserHasMFAFactor(
	ctx context.Context,
	q store.Queryer,
	userID string,
) (bool, error) {
	var f bool
	err := q.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM auth_mfa_factors WHERE user_id=$1)`,
		userID,
	).Scan(&f)
	return f, err
}

// CreateMFAChallenge inserts an MFA challenge with the code hash and returns its ID.
func (r *repo) CreateMFAChallenge(
	ctx context.Context,
	q store.Queryer,
	userID string,
	ttl time.Duration,
	codeHash string,
) (string, error) {
	var id string
	err := q.QueryRow(
		ctx,
		`INSERT INTO auth_mfa_challenges (user_id, factor_id, code_hash, max_attempts, expires_at)
		 VALUES ($1, NULL, $2, 5, NOW() + $3::interval)
		 RETURNING id::text`,
		userID,
		codeHash,
		ttl.String(),
	).Scan(&id)
	return id, err
}

// VerifyAndConsumeMFA atomically checks a challenge code hash, increments attempts on mismatch,
// fulfills on match, and returns (ok, userID).
func (r *repo) VerifyAndConsumeMFA(
	ctx context.Context,
	q store.Queryer,
	challengeID string,
	codeHash string,
) (bool, string, error) {
	var userID string
	var ok bool
	err := q.QueryRow(
		ctx,
		`UPDATE auth_mfa_challenges
		   SET attempts    = CASE WHEN $2 = code_hash THEN attempts ELSE attempts + 1 END,
		       fulfilled_at = CASE WHEN $2 = code_hash THEN NOW() ELSE fulfilled_at END
		 WHERE id=$1 AND fulfilled_at IS NULL AND expires_at > NOW() AND attempts < max_attempts
		 RETURNING user_id::text, $2 = code_hash`,
		challengeID,
		codeHash,
	).Scan(&userID, &ok)
	return ok, userID, err
}

// InsertSession inserts a refresh session (hashed token) with UA/IP and expiry.
func (r *repo) InsertSession(
	ctx context.Context,
	q store.Queryer,
	userID string,
	tenantID string, // empty string => NULL
	refreshHash string,
	userAgent string,
	ip string,
	ttl time.Duration,
) error {
	_, err := q.Exec(ctx, `
		INSERT INTO auth_sessions (
			user_id, tenant_id, refresh_token_hash, user_agent, ip, expires_at
		) VALUES (
			$1,
			NULLIF($2, '')::uuid,                     -- cast AFTER NULLIF
			$3,
			$4,
			$5,
			NOW() + ($6::bigint * interval '1 second') -- build interval from seconds
		)
	`,
		userID,
		tenantID, // "" -> NULL
		refreshHash,
		userAgent,
		ip,
		int64(ttl/time.Second), // pass seconds, not "720h0m0s"
	)
	return err
}

// InsertLoginAttempt records a login attempt for auditing and lockout logic.
func (r *repo) InsertLoginAttempt(
	ctx context.Context,
	q store.Queryer,
	userID *string,
	email string,
	success bool,
	reason string,
	ip string,
	ua string,
) error {
	var uid any
	if userID == nil {
		uid = nil
	} else {
		uid = *userID
	}
	_, err := q.Exec(
		ctx,
		`INSERT INTO auth_login_attempts (user_id, email, success, reason, ip, user_agent)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		uid,
		email,
		success,
		reason,
		ip,
		ua,
	)
	return err
}

// CreateUser inserts a new user and returns its ID.
func (r *repo) CreateUser(
	ctx context.Context,
	q store.Queryer,
	email string,
	pwHash string,
	name string,
) (string, error) {
	var id string
	err := q.QueryRow(
		ctx,
		`INSERT INTO users (email, password_hash, name)
		 VALUES (LOWER($1), $2, NULLIF($3,''))
		 RETURNING id::text`,
		email,
		pwHash,
		name,
	).Scan(&id)
	return id, err
}

// EnsureTenantBySlug ensures a tenant exists for slug and returns its ID.
func (r *repo) EnsureTenantBySlug(
	ctx context.Context,
	q store.Queryer,
	slug string,
) (string, error) {
	var id string
	err := q.QueryRow(
		ctx,
		`INSERT INTO tenants (slug, name)
		   VALUES ($1, $1)
		 ON CONFLICT (slug) DO UPDATE SET slug = EXCLUDED.slug
		 RETURNING id::text`,
		slug,
	).Scan(&id)
	return id, err
}

// UpsertUserTenantAdmin adds (or keeps) the user as admin in the tenant.
func (r *repo) UpsertUserTenantAdmin(
	ctx context.Context,
	q store.Queryer,
	userID string,
	tenantID string,
) error {
	_, err := q.Exec(
		ctx,
		`INSERT INTO users_tenants (user_id, tenant_id, role, is_primary)
		 VALUES ($1, $2, 'admin', true)
		 ON CONFLICT (user_id, tenant_id) DO NOTHING`,
		userID,
		tenantID,
	)
	return err
}

// SetPrimaryTenantIfNull sets the user's primary tenant if it is currently NULL.
func (r *repo) SetPrimaryTenantIfNull(
	ctx context.Context,
	q store.Queryer,
	userID string,
	tenantID string,
) error {
	_, err := q.Exec(
		ctx,
		`UPDATE users SET primary_tenant_id=$2 WHERE id=$1 AND primary_tenant_id IS NULL`,
		userID,
		tenantID,
	)
	return err
}

// GetActiveSessionByHash returns (userID, tenantID) for an active session by token hash.
func (r *repo) GetActiveSessionByHash(
	ctx context.Context,
	q store.Queryer,
	hash string,
) (string, string, error) {
	var uid, tid string
	err := q.QueryRow(
		ctx,
		`SELECT user_id::text, COALESCE(tenant_id::text,'')
		   FROM auth_sessions
		  WHERE refresh_token_hash=$1
		    AND revoked_at IS NULL
		    AND expires_at > NOW()
		  LIMIT 1`,
		hash,
	).Scan(&uid, &tid)
	return uid, tid, err
}

// RevokeSessionByHash marks a session revoked by its token hash (idempotent).
func (r *repo) RevokeSessionByHash(
	ctx context.Context,
	q store.Queryer,
	hash string,
) error {
	_, err := q.Exec(
		ctx,
		`UPDATE auth_sessions SET revoked_at = NOW()
		   WHERE refresh_token_hash=$1 AND revoked_at IS NULL`,
		hash,
	)
	return err
}

// GetUserIDByEmail returns the user ID for a normalized email.
func (r *repo) GetUserIDByEmail(
	ctx context.Context,
	q store.Queryer,
	email string,
) (string, error) {
	var id string
	err := q.QueryRow(
		ctx,
		`SELECT id::text FROM users WHERE email = LOWER($1)`,
		email,
	).Scan(&id)
	return id, err
}

// InsertPasswordResetToken inserts a hashed reset token with IP and expiry.
func (r *repo) InsertPasswordResetToken(
	ctx context.Context,
	q store.Queryer,
	userID string,
	tokenHash string,
	requestedIP string,
	ttl time.Duration,
) error {
	_, err := q.Exec(
		ctx,
		`INSERT INTO auth_password_reset_tokens (user_id, token_hash, requested_ip, expires_at)
		 VALUES ($1, $2, $3, NOW() + $4::interval)`,
		userID,
		tokenHash,
		requestedIP,
		ttl.String(),
	)
	return err
}

// LookupResetUserID returns the user ID for a valid, unused reset token hash.
func (r *repo) LookupResetUserID(
	ctx context.Context,
	q store.Queryer,
	tokenHash string,
) (string, error) {
	var uid string
	err := q.QueryRow(
		ctx,
		`SELECT user_id::text
		   FROM auth_password_reset_tokens
		  WHERE token_hash=$1 AND used_at IS NULL AND expires_at > NOW()`,
		tokenHash,
	).Scan(&uid)
	return uid, err
}

// UpdateUserPasswordHash replaces the user's password hash and bumps updated_at.
func (r *repo) UpdateUserPasswordHash(
	ctx context.Context,
	q store.Queryer,
	userID string,
	newHash string,
) error {
	_, err := q.Exec(
		ctx,
		`UPDATE users SET password_hash=$2, updated_at=NOW() WHERE id=$1`,
		userID,
		newHash,
	)
	return err
}

// MarkPasswordResetUsed marks a reset token as consumed.
func (r *repo) MarkPasswordResetUsed(
	ctx context.Context,
	q store.Queryer,
	tokenHash string,
) error {
	_, err := q.Exec(
		ctx,
		`UPDATE auth_password_reset_tokens SET used_at=NOW()
		   WHERE token_hash=$1 AND used_at IS NULL`,
		tokenHash,
	)
	return err
}

// RevokeAllSessionsForUser revokes all active sessions for the user.
func (r *repo) RevokeAllSessionsForUser(
	ctx context.Context,
	q store.Queryer,
	userID string,
) error {
	_, err := q.Exec(
		ctx,
		`UPDATE auth_sessions SET revoked_at=NOW()
		   WHERE user_id=$1 AND revoked_at IS NULL`,
		userID,
	)
	return err
}
