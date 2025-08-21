import { api } from "./client"

export type SignupInput = {
  email: string
  password: string
  name?: string
  tenant_slug?: string // optional invite-less tenant selection
}

export type LoginInput = {
  email: string
  password: string
  mfa_code?: string
  tenant_id?: string
}

export type AuthResult = {
  user: { id: string; email: string; name?: string; primary_tenant_id?: string }
  access_token: string // short-lived
  expires_in: number // seconds
  mfa_required?: { challenge_id: string; factors: string[] }
}

export const AuthAPI = {
  signup: (input: SignupInput) => api.post<AuthResult>("/auth/signup", input),
  login: (input: LoginInput) => api.post<AuthResult>("/auth/login", input),
  mfaVerify: (challenge_id: string, code: string) =>
    api.post<AuthResult>("/auth/mfa/verify", { challenge_id, code }),
  forgot: (email: string) => api.post<void>("/auth/forgot", { email }),
  reset: (token: string, password: string) => api.post<void>("/auth/reset", { token, password }),
  me: () => api.get<AuthResult["user"]>("/auth/me"),
  logout: () => api.post<void>("/auth/logout"),
}
