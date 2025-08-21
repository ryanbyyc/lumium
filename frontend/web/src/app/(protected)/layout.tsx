import { Navbar } from "@/components/Navbar"
import type { ReactNode } from "react"

export default function ProtectedLayout({ children }: { children: ReactNode }) {
  return (
    <div className="bg-background text-foreground min-h-screen">
      <Navbar variant="protected" />
      <div className="mx-auto max-w-6xl p-6">{children}</div>
    </div>
  )
}
