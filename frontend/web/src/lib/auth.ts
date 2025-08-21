import NextAuth from "next-auth"
import Credentials from "next-auth/providers/credentials"

type AccessTokenShape = {
  accessToken?: string
  refreshToken?: string
  accessTokenExpires?: number // epoch ms
  role?: string
}

async function refreshAccessToken(token: any): Promise<AccessTokenShape> {
  // TODO: replace withGo API refresh endpoint
  // I.e. :
  // const res = await fetch(process.env.API_URL + '/auth/refresh', {
  //   method: 'POST',
  //   headers: { 'Content-Type': 'application/json' },
  //   body: JSON.stringify({ refreshToken: token.refreshToken }),
  //   cache: 'no-store',
  // })
  // if (!res.ok) throw new Error('Refresh failed')
  // const data = await res.json()
  // return {
  //   accessToken: data.accessToken,
  //   refreshToken: data.refreshToken ?? token.refreshToken,
  //   accessTokenExpires: Date.now() + data.expiresIn * 1000,
  //   role: token.role,
  // }

  // Temporary fallback so frontend dev is smooth
  return {
    accessToken: token.accessToken,
    refreshToken: token.refreshToken,
    accessTokenExpires: Date.now() + 15 * 60 * 1000,
    role: token.role,
  }
}

export const { auth, handlers, signIn, signOut } = NextAuth({
  providers: [
    Credentials({
      credentials: {
        username: { label: "Username", type: "text" },
        password: { label: "Password", type: "password" },
      },
      async authorize(creds) {
        console.log("[authorize] creds:", creds)
        if (!creds?.username || !creds?.password) return null

        if (creds.username === "admin" && creds.password === "admin") {
          return {
            id: "1",
            name: "Admin",
            email: "admin@example.com",
            role: "admin",
            accessToken: "dev-admin-token",
            refreshToken: "dev-admin-refresh",
            expiresIn: 15 * 60,
          } as any
        }
        if (creds.username === "user" && creds.password === "user") {
          return {
            id: "2",
            name: "User",
            email: "user@example.com",
            role: "user",
            accessToken: "dev-user-token",
            refreshToken: "dev-user-refresh",
            expiresIn: 15 * 60,
          } as any
        }
        return null
      },
    }),
  ],
  session: { strategy: "jwt" },
  callbacks: {
    async jwt({ token, user, trigger }) {
      if (user) {
        const u = user as any
        token.role = u.role
        token.accessToken = u.accessToken ?? token.accessToken
        token.refreshToken = u.refreshToken ?? token.refreshToken
        if (u.expiresIn) token.accessTokenExpires = Date.now() + u.expiresIn * 1000
      }

      if (token.accessToken && token.accessTokenExpires && Date.now() < +token.accessTokenExpires) {
        return token
      }

      try {
        const refreshed = await refreshAccessToken(token)
        return {
          ...token,
          accessToken: refreshed.accessToken,
          refreshToken: refreshed.refreshToken ?? token.refreshToken,
          accessTokenExpires: refreshed.accessTokenExpires,
          role: refreshed.role ?? token.role,
        }
      } catch {
        return { ...token, accessToken: null, refreshToken: null, accessTokenExpires: 0 }
      }
    },
    async session({ session, token }) {
      ;(session as any).role = token.role
      ;(session as any).accessToken = token.accessToken
      return session
    },
  },
  pages: {},
})
