"use server"

export async function signup(_prev: { error?: string } | null, formData: FormData) {
  const username = formData.get("username") as string
  const email = formData.get("email") as string
  const password = formData.get("password") as string

  try {
    // await api.signup({ username, email, password })
    return { ok: true }
  } catch {
    return { error: "Signup failed" }
  }
}
