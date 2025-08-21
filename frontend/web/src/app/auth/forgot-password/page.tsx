"use client"

import { Button, CardBody, CardHeader, Input } from "@heroui/react"
import { useActionState } from "react"
import { forgotPassword } from "./actions"

export default function ForgotPasswordPage() {
  const [state, formAction, pending] = useActionState(forgotPassword, null)

  return (
    <>
      <CardHeader className="flex items-center justify-between px-6 py-6">
        <div>
          <h1 className="text-xl font-semibold">Forgot password</h1>
          <p className="mt-1 text-sm text-[hsl(var(--text-2))]">
            Enter your email to reset your password
          </p>
        </div>
        <div className="rounded-xl border border-white/10 bg-white/10 px-3 py-1 text-xs text-white/80">
          <a href="/" className="text-sm text-[hsl(var(--text-2))] hover:text-white">
            Lumium
          </a>
        </div>
      </CardHeader>
      <CardBody className="px-6 pt-2 pb-8">
        <form action={formAction} className="grid gap-5">
          <Input
            name="email"
            type="email"
            placeholder="Email"
            variant="bordered"
            classNames={{ inputWrapper: "bg-[hsl(var(--surface-2))] border-white/10" }}
            isRequired
          />

          {(state as any)?.message && (
            <p className="text-sm text-green-400">{(state as any).message}</p>
          )}
          {(state as any)?.error && <p className="text-sm text-red-400">{(state as any).error}</p>}

          <Button type="submit" isLoading={pending} className="btn-primary mt-1 h-11 rounded-md">
            Send reset link
          </Button>

          <p className="mt-2 text-center text-sm text-[hsl(var(--text-2))]">
            Remembered your password?{" "}
            <a className="underline hover:text-white" href="/auth/login">
              Log in
            </a>
          </p>
        </form>
      </CardBody>
    </>
  )
}
