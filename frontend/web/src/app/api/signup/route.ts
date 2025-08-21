import { NextResponse } from "next/server"

// @TODO: Replace this with a proxy to the Go API
export async function POST(req: Request) {
  const { username, email, password } = await req.json()

  if (!username || !email || !password) {
    return NextResponse.json({ error: "Missing fields" }, { status: 400 })
  }

  // TODO: call Go API /auth/signup
  // Example:
  // const res = await fetch(process.env.API_URL + '/auth/signup', { ... })
  // return NextResponse.json(await res.json(), { status: res.status })

  // Dev stub success
  return NextResponse.json({ ok: true }, { status: 201 })
}
