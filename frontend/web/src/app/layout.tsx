import ClientProviders from "@/components/ClientProviders"
import type { ReactNode } from "react"
import "./globals.css"

export const metadata = { title: "Lumium" }

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" data-theme="crimson">
      <body className="bg-background text-foreground">
        <ClientProviders>{children}</ClientProviders>
      </body>
    </html>
  )
}
