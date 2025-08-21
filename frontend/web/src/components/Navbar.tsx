"use client"

import { signOut } from "@/lib/auth"
import {
  Button,
  Navbar as HNavbar,
  Link,
  NavbarBrand,
  NavbarContent,
  NavbarItem,
} from "@heroui/react"
import { usePathname } from "next/navigation"

export function Navbar({ variant }: { variant: "public" | "protected" | "admin" }) {
  const pathname = usePathname()

  return (
    <HNavbar maxWidth="xl">
      <NavbarBrand>
        <Link href="/" className="font-semibold">
          Lumium
        </Link>
      </NavbarBrand>

      <NavbarContent justify="center" className="gap-6">
        <NavbarItem isActive={pathname.startsWith("/protected")}>
          <Link href="/protected">Dashboard</Link>
        </NavbarItem>
        <NavbarItem isActive={pathname.startsWith("/admin")}>
          <Link href="/admin">Admin</Link>
        </NavbarItem>
      </NavbarContent>

      <NavbarContent justify="end" className="gap-3">
        {variant === "public" ? (
          <NavbarItem>
            <Button as={Link} href="/login" color="primary">
              Sign in
            </Button>
          </NavbarItem>
        ) : (
          <>
            <NavbarItem className="hidden opacity-70 sm:flex">{variant}</NavbarItem>
            <NavbarItem>
              <Button
                variant="flat"
                onPress={async () => {
                  await signOut({ redirectTo: "/" })
                }}
              >
                Sign out
              </Button>
            </NavbarItem>
          </>
        )}
      </NavbarContent>
    </HNavbar>
  )
}
