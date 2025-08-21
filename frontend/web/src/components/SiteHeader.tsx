"use client"

import { Button, Link, Navbar, NavbarBrand, NavbarContent, NavbarItem } from "@heroui/react"

export default function SiteHeader() {
  return (
    <Navbar maxWidth="xl" className="bg-transparent">
      <NavbarBrand>
        <Link href="/" className="font-semibold tracking-tight">
          Lumium
        </Link>
      </NavbarBrand>

      <NavbarContent justify="center" className="hidden gap-6 md:flex">
        <NavbarItem>
          <Link href="#features" className="opacity-80 hover:opacity-100">
            Features
          </Link>
        </NavbarItem>
        <NavbarItem>
          <Link href="#how" className="opacity-80 hover:opacity-100">
            How it works
          </Link>
        </NavbarItem>
        <NavbarItem>
          <Link href="#faq" className="opacity-80 hover:opacity-100">
            FAQ
          </Link>
        </NavbarItem>
      </NavbarContent>

      <NavbarContent justify="end" className="gap-3">
        <NavbarItem className="hidden sm:flex">
          <Button as={Link} href="/login" variant="bordered">
            Log in
          </Button>
        </NavbarItem>
        <NavbarItem>
          <Button as={Link} href="/protected" color="primary">
            Open app
          </Button>
        </NavbarItem>
      </NavbarContent>
    </Navbar>
  )
}
