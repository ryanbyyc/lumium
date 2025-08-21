import { NextResponse } from "next/server"
import { auth } from "./lib/auth"

const PUBLIC_PATHS = ["/login", "/signup", "/", "/_next", "/api", "/public", "/favicon.ico"]

export default auth((req) => {
  const { nextUrl } = req
  const isPublic = PUBLIC_PATHS.some((p) => nextUrl.pathname.startsWith(p))
  const user = (req as any).auth?.user
  const role = (req as any).auth?.user?.role

  // Redirect authed users away from auth pages
  if (user && (nextUrl.pathname === "/login" || nextUrl.pathname === "/signup")) {
    return NextResponse.redirect(new URL("/dashboard", nextUrl))
  }

  // Protect /protected and /admin
  if (
    !isPublic &&
    (nextUrl.pathname.startsWith("/protected") || nextUrl.pathname.startsWith("/admin"))
  ) {
    if (!user) {
      const url = new URL("/login", nextUrl)
      url.searchParams.set("callbackUrl", nextUrl.pathname)
      return NextResponse.redirect(url)
    }
    // Simple role check for /admin
    if (nextUrl.pathname.startsWith("/admin") && role !== "admin") {
      return NextResponse.redirect(new URL("/dashboard", nextUrl))
    }
  }

  return NextResponse.next()
})

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
}
