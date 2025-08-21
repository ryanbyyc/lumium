"use client"

import { Card } from "@heroui/react"
import { ReactNode } from "react"

export default function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <div className="relative min-h-dvh text-white">
      <div aria-hidden className="absolute inset-0 overflow-hidden">
        <div className="absolute inset-0 bg-[hsl(var(--surface-0))]" />
        <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(950px_500px_at_92%_-10%,hsl(var(--brand-1)/0.28),transparent_60%),radial-gradient(700px_420px_at_6%_100%,hsl(var(--brand-2)/0.22),transparent_60%)]" />
        <svg
          className="absolute inset-0 h-full w-full opacity-[0.06]"
          xmlns="http://www.w3.org/2000/svg"
        >
          <defs>
            <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
              <path d="M 40 0 L 0 0 0 40" fill="none" stroke="white" strokeWidth="0.5" />
            </pattern>
          </defs>
          <rect width="100%" height="100%" fill="url(#grid)" />
        </svg>
        <div className="pointer-events-none absolute inset-x-0 bottom-0 h-[32vh] bg-[radial-gradient(60%_80%_at_50%_100%,hsl(var(--vignette)/0.65),transparent_60%)]" />
      </div>

      <div className="relative mx-auto flex min-h-dvh w-full max-w-[1400px] items-center justify-end px-4">
        <div className="w-full md:max-w-md">
          <Card className="card-surface ring-brand shadow-2xl backdrop-blur-xl">{children}</Card>
        </div>
      </div>
    </div>
  )
}
