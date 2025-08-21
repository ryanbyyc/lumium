CREATE EXTENSION IF NOT EXISTS pgcrypto;  -- gen_random_uuid()

CREATE TYPE role_enum AS ENUM ('admin','member','viewer');

CREATE TABLE tenants (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug TEXT UNIQUE NOT NULL CHECK (slug ~ '^[a-z0-9-]{3,}$'),
  name TEXT NOT NULL,
  mfa_required BOOLEAN NOT NULL DEFAULT FALSE, -- tenant-level MFA override
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT NOT NULL,  -- uniqueness enforced via lower() index below
  email_verified_at TIMESTAMPTZ,
  password_hash TEXT NOT NULL, -- store algorithm+params+salt in one hash string (argon2/bcrypt)
  name TEXT,
  primary_tenant_id UUID REFERENCES tenants(id),
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX users_idx_lower_email ON users (LOWER(email));

CREATE TABLE users_tenants (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  role role_enum NOT NULL, -- 'admin' | 'member' | 'viewer'
  is_primary BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, tenant_id)
);
CREATE INDEX users_tenants_idx_tenant_id ON users_tenants (tenant_id);


CREATE TABLE auth_permissions (
  code TEXT PRIMARY KEY, -- e.g. 'users.read', 'users.write', 'jobs.read', ...
  description TEXT NOT NULL
);

CREATE TABLE auth_role_permissions (
  role role_enum NOT NULL,
  permission_code TEXT REFERENCES auth_permissions(code) ON DELETE CASCADE,
  tenant_scoped BOOLEAN NOT NULL DEFAULT TRUE,
  PRIMARY KEY (role, permission_code)
);

CREATE TABLE auth_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  tenant_id UUID REFERENCES tenants(id), -- session may be tenant-contextual
  refresh_token_hash TEXT NOT NULL, -- store hash only; rotate on refresh
  user_agent TEXT,
  ip INET,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ
);
CREATE INDEX auth_sessions_idx_user_id ON auth_sessions (user_id);
CREATE INDEX auth_sessions_idx_expires_at ON auth_sessions (expires_at);

CREATE TABLE auth_password_reset_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL, -- never store plaintext
  requested_ip INET,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at TIMESTAMPTZ NOT NULL,
  UNIQUE (user_id, token_hash)
);

CREATE TABLE auth_one_time_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  purpose TEXT NOT NULL CHECK (purpose IN ('email_verify','mfa','invite')),
  token_hash TEXT NOT NULL, -- token -> hash in DB
  meta JSONB NOT NULL DEFAULT '{}', -- e.g. { "challenge_id": "...", "factor": "email" }
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE auth_login_attempts (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  email TEXT, -- captured even if user_id not found
  success BOOLEAN NOT NULL,
  reason TEXT, -- 'invalid_password','mfa_required','ok','locked',...
  ip INET,
  user_agent TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX auth_login_attempts_idx_email ON auth_login_attempts (email);
CREATE INDEX auth_login_attempts_idx_created_at ON auth_login_attempts (created_at DESC);

CREATE TABLE auth_mfa_factors (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type TEXT NOT NULL CHECK (type IN ('email','sms','totp')),
  label TEXT, -- 'work phone', etc.
  secret TEXT, -- TOTP secret or E.164 phone; email lives in users.email
  last_verified_at TIMESTAMPTZ,
  is_primary BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE auth_mfa_challenges (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  factor_id UUID REFERENCES auth_mfa_factors(id) ON DELETE SET NULL,
  code_hash TEXT NOT NULL, -- never store code plaintext
  attempts INT NOT NULL DEFAULT 0,
  max_attempts INT NOT NULL DEFAULT 5,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at TIMESTAMPTZ NOT NULL,
  fulfilled_at TIMESTAMPTZ
);


-- ============================
-- ROLES
-- ============================

-- Sys role: lumiumbot (Docker's POSTGRES_USER)
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'lumiumbot') THEN
    ALTER ROLE lumiumbot BYPASSRLS;
  END IF;
END$$;

-- App consumer role: lumiumapp (used by the API/services)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'lumiumapp') THEN
    CREATE ROLE lumiumapp LOGIN PASSWORD 'lumium';
  END IF;
END$$;

GRANT USAGE ON SCHEMA public TO lumiumapp;

-- Give lumiumapp broad CRUD on ALL existing tables in public
-- RLS will still enforce row-level access where enabled
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO lumiumapp;

GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO lumiumapp;

GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO lumiumapp;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO lumiumapp;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO lumiumapp;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT EXECUTE ON FUNCTIONS TO lumiumapp;

ALTER TABLE users_tenants ENABLE ROW LEVEL SECURITY;

-- SELECT within current tenant
CREATE POLICY tenant_membership_select ON users_tenants
  FOR SELECT
  USING (tenant_id::TEXT = current_setting('app.tenant_id', TRUE));

-- INSERT restricted to current tenant
CREATE POLICY tenant_membership_insert ON users_tenants
  FOR INSERT
  WITH CHECK (tenant_id::TEXT = current_setting('app.tenant_id', TRUE));

-- UPDATE restricted to current tenant
CREATE POLICY tenant_membership_update ON users_tenants
  FOR UPDATE
  USING (tenant_id::TEXT = current_setting('app.tenant_id', TRUE))
  WITH CHECK (tenant_id::TEXT = current_setting('app.tenant_id', TRUE));

-- DELETE restricted to current tenant
CREATE POLICY tenant_membership_delete ON users_tenants
  FOR DELETE
  USING (tenant_id::TEXT = current_setting('app.tenant_id', TRUE));
