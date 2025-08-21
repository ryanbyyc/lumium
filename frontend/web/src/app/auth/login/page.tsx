"use client"

import { Button, CardBody, CardHeader, Checkbox, Divider, Input } from "@heroui/react"
import { useRouter, useSearchParams } from "next/navigation"
import { useActionState, useEffect } from "react"
import { login } from "./actions"

export default function LoginPage() {
  const router = useRouter()
  const params = useSearchParams()
  const callbackUrl = params.get("callbackUrl") ?? "/dashboard"
  const [state, formAction, pending] = useActionState(login, null)

  useEffect(() => {
    if ((state as any)?.ok) router.replace(callbackUrl)
  }, [state, callbackUrl, router])

  return (
    <>
      <CardHeader className="flex items-center justify-between px-6 py-6">
        <div>
          <h1 className="text-xl font-semibold">Sign in</h1>
          <p className="mt-1 text-sm text-[hsl(var(--text-2))]">Welcome back</p>
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
            name="username"
            type="text"
            placeholder="Username"
            variant="bordered"
            classNames={{ inputWrapper: "bg-[hsl(var(--surface-2))] border-white/10" }}
            isRequired
          />
          <Input
            name="password"
            type="password"
            placeholder="Password"
            variant="bordered"
            classNames={{ inputWrapper: "bg-[hsl(var(--surface-2))] border-white/10" }}
            isRequired
          />

          <div className="mt-1 flex items-center justify-between">
            <Checkbox name="remember" radius="sm">
              Remember me
            </Checkbox>
            <a
              href="/auth/forgot-password"
              className="text-sm text-[hsl(var(--text-2))] underline hover:text-white"
            >
              Forgot password?
            </a>
          </div>

          <input type="hidden" name="callbackUrl" value={callbackUrl} />

          {(state as any)?.error && <p className="text-sm text-red-400">{(state as any).error}</p>}

          <Button type="submit" isLoading={pending} className="btn-primary mt-1 h-11 rounded-md">
            Log in
          </Button>

          <Divider className="my-3 opacity-40" />

          <div className="grid gap-2 sm:grid-cols-2">
            <Button variant="flat" className="w-full" isDisabled>
              Continue with Google
            </Button>
            <Button variant="flat" className="w-full" isDisabled>
              Continue with GitHub
            </Button>
          </div>

          <p className="mt-2 text-center text-sm text-[hsl(var(--text-2))]">
            Need to create an account?{" "}
            <a className="underline hover:text-white" href="/auth/signup">
              Sign Up
            </a>
          </p>
        </form>
      </CardBody>
    </>
  )
}
