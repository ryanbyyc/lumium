import { create } from "zustand"
import { AuthAPI, type AuthResult } from "../api/auth"

type State = {
  user?: AuthResult["user"]
  accessToken?: string
  loading: boolean
  error?: string
  setUser: (u?: AuthResult["user"]) => void
  login: (email: string, password: string, tenant_id?: string, mfa_code?: string) => Promise<void>
  logout: () => Promise<void>
  bootstrap: () => Promise<void>
}

export const useAuth = create<State>((set) => ({
  loading: false,
  setUser: (user) => set({ user }),
  login: async (email, password, tenant_id, mfa_code) => {
    set({ loading: true, error: undefined })
    try {
      const res = await AuthAPI.login({ email, password, tenant_id, mfa_code })
      if (res.mfa_required) {
        set({ loading: false })
        // UI should collect code -> call AuthAPI.mfaVerify
        return
      }
      set({ user: res.user, accessToken: res.access_token, loading: false })
    } catch (e: any) {
      set({ error: e?.message ?? "Login failed", loading: false })
      throw e
    }
  },
  logout: async () => {
    await AuthAPI.logout()
    set({ user: undefined, accessToken: undefined })
  },
  bootstrap: async () => {
    set({ loading: true })
    try {
      const me = await AuthAPI.me()
      set({ user: me, loading: false })
    } catch {
      set({ user: undefined, loading: false })
    }
  },
}))
