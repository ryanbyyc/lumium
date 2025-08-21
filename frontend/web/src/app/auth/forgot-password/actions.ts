"use server"

export async function forgotPassword(
  _prev: { error?: string; ok?: boolean } | null,
  formData: FormData,
) {
  const email = formData.get("email") as string

  try {
    // await api.forgotPassword({ email })
    return { ok: true }
  } catch {
    return { error: "Failed to send reset link" }
  }
}
