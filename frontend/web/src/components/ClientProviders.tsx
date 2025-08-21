"use client"
import type { ReactNode } from "react"
import { HeroUIProvider } from "@heroui/react"
export default function ClientProviders({ children }: { children: ReactNode }) {
  return <HeroUIProvider>{children}</HeroUIProvider>
}