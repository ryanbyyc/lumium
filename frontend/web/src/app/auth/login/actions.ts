"use server"

import { signIn } from "@/lib/auth"
import { AuthError } from "next-auth"

export async function login(_prev: { error?: string } | null, formData: FormData) {
  const username = formData.get("username") as string
  const password = formData.get("password") as string

  try {
    // Stay on /login: no redirect here
    await signIn("credentials", { username, password, redirect: false })
    return { ok: true }
  } catch (e) {
    if (e instanceof AuthError && e.type === "CredentialsSignin") {
      return { error: "Invalid username or password" }
    }
    return { error: "Something went wrong" }
  }
}
