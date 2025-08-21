"use client"

export default function SiteFooter() {
  return (
    <footer className="border-t border-white/10">
      <div className="mx-auto max-w-6xl px-6 py-10 text-sm text-white/60">
        <div className="flex flex-col gap-6 sm:flex-row sm:items-center sm:justify-between">
          <p>© {new Date().getFullYear()} Lumium. All rights reserved.</p>
          <nav className="flex gap-6">
            <a href="#features" className="hover:text-white">
              Features
            </a>
            <a href="#faq" className="hover:text-white">
              FAQ
            </a>
            <a href="/login" className="hover:text-white">
              Log in
            </a>
          </nav>
        </div>
      </div>
    </footer>
  )
}
