"use client"

import { Button, CardBody, CardHeader, Divider, Input } from "@heroui/react"
import { useRouter } from "next/navigation"
import { useActionState, useEffect } from "react"
import { signup } from "./actions"

export default function SignupPage() {
  const router = useRouter()
  const [state, formAction, pending] = useActionState(signup, null)

  useEffect(() => {
    if ((state as any)?.ok) router.replace("/dashboard")
  }, [state, router])

  return (
    <>
      <CardHeader className="flex items-center justify-between px-6 py-6">
        <div>
          <h1 className="text-xl font-semibold">Register</h1>
          <p className="mt-1 text-sm text-[hsl(var(--text-2))]">Create your Lumium account</p>
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
            name="email"
            type="email"
            placeholder="Email"
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

          {(state as any)?.error && <p className="text-sm text-red-400">{(state as any).error}</p>}

          <Button type="submit" isLoading={pending} className="btn-primary mt-1 h-11 rounded-md">
            Sign up
          </Button>

          <Divider className="my-3 opacity-40" />

          <p className="mt-2 text-center text-sm text-[hsl(var(--text-2))]">
            Already have an account?{" "}
            <a className="underline hover:text-white" href="/auth/login">
              Log in
            </a>
          </p>
        </form>
      </CardBody>
    </>
  )
}
