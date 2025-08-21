"use client"

import {
  Button,
  Card,
  CardBody,
  CardHeader,
  Chip,
  Divider,
  Link as HLink,
  Input,
  Tooltip,
} from "@heroui/react"
import { motion } from "framer-motion"
import Link from "next/link"

const EASE = [0.16, 1, 0.3, 1] as const
const EASE_IO = [0.4, 0, 0.2, 1] as const

const popIn = (delay = 0) => ({
  initial: { opacity: 0, y: 12, scale: 0.98 },
  whileInView: { opacity: 1, y: 0, scale: 1 },
  viewport: { once: true, amount: 0.2 },
  transition: { duration: 0.45, ease: EASE, delay },
})

export default function LandingPage() {
  return (
    <div
      className="relative min-h-screen overflow-hidden text-white"
      style={{ backgroundColor: "hsl(var(--surface-0))" }}
    >
      <Decor />
      <main className="relative z-10 pt-16">
        <header className="nav-blur fixed inset-x-0 top-0 z-50">
          <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4 sm:px-6">
            <Link href="/" className="font-semibold tracking-tight">
              <motion.span
                initial={{ filter: "drop-shadow(0 0 0 hsl(var(--glow-1)/0))", opacity: 0.96 }}
                animate={{
                  opacity: [0.96, 1, 0.96],
                  filter: [
                    "drop-shadow(0 0 0 hsl(var(--glow-1)/0))",
                    "drop-shadow(0 0 12px hsl(var(--glow-1)/0.85))",
                    "drop-shadow(0 0 0 hsl(var(--glow-1)/0))",
                  ],
                }}
                transition={{ duration: 4.5, repeat: Infinity, ease: EASE_IO }}
                className="text-brand-gradient"
              >
                Lumium
              </motion.span>
            </Link>
            <nav className="hidden gap-6 text-sm text-white/80 sm:flex">
              <HLink href="#features" className="hover:text-[hsl(var(--brand-1))]">
                Features
              </HLink>
              <HLink href="#pipeline" className="hover:text-[hsl(var(--brand-1))]">
                How it works
              </HLink>
              <HLink href="#faq" className="hover:text-[hsl(var(--brand-1))]">
                FAQ
              </HLink>
            </nav>
            <div className="flex items-center gap-2">
              <Button as={Link} href="/auth/login" size="sm" className="btn-secondary">
                Log in
              </Button>
              <Button as={Link} href="/protected" size="sm" className="btn-primary">
                Open app
              </Button>
            </div>
          </div>
        </header>

        <section className="relative mx-auto max-w-6xl px-4 sm:px-6">
          <div className="py-20 text-center sm:py-28 sm:text-left lg:py-32">
            <motion.div {...popIn(0)}>
              <Chip color="primary" className="mb-4">
                Self-hosted photo platform
              </Chip>
            </motion.div>

            <motion.h1
              className="text-brand-gradient max-w-3xl text-4xl leading-[2.08] font-extrabold sm:text-6xl sm:leading-[2.06] xl:leading-[2.04]"
              initial={{ scale: 1, backgroundPosition: "0% 50%" }}
              animate={{
                scale: [1, 1.02, 1],
                backgroundPosition: ["0% 50%", "100% 50%", "0% 50%"],
              }}
              transition={{ duration: 8, repeat: Infinity, ease: EASE_IO }}
              style={{
                backgroundSize: "220% 220%",
                filter: "drop-shadow(0 0 18px hsl(var(--brand-1)/0.25))",
              }}
            >
              Own your photos.
            </motion.h1>

            <motion.h2
              className="mt-2 max-w-3xl text-2xl font-semibold text-white/90 sm:text-3xl"
              {...popIn(0.05)}
            >
              Organize, deduplicate, and back them up securely.
            </motion.h2>

            <motion.p className="mt-4 max-w-2xl text-white/70" {...popIn(0.1)}>
              Lumium helps you tame sprawling libraries with fast imports, smart duplicates, and
              storage targets you control.
            </motion.p>

            <motion.div
              className="mt-8 flex flex-col gap-3 sm:flex-row sm:justify-start"
              {...popIn(0.15)}
            >
              <Button as={Link} href="/protected" size="lg" className="btn-primary">
                Open app
              </Button>
              <Button as={Link} href="/auth/login" variant="bordered" size="lg">
                Log in
              </Button>
            </motion.div>

            <motion.div className="mx-auto mt-10 max-w-lg sm:mx-0" {...popIn(0.2)}>
              <Card className="card-surface ring-brand">
                <CardBody className="flex flex-col gap-3 sm:flex-row sm:items-center">
                  <Input
                    aria-label="Your email"
                    placeholder="you@domain.com"
                    type="email"
                    variant="bordered"
                    className="flex-1"
                  />
                  <Button className="btn-primary">Get updates</Button>
                </CardBody>
              </Card>
            </motion.div>
          </div>
        </section>

        <section id="features" className="relative mx-auto max-w-6xl px-4 pb-20 sm:px-6">
          <SectionDecor />
          <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            <motion.div {...popIn(0)}>
              <Feature
                icon="stack"
                title="Fast, lossless imports"
                text="Ingest from local or remote sources with zero quality loss and automatic metadata preservation."
              />
            </motion.div>
            <motion.div {...popIn(0.05)}>
              <Feature
                icon="eye"
                title="Smart deduplication"
                text="Perceptual hashing spots near-duplicates and burst shots so you choose the best and archive the rest."
              />
            </motion.div>
            <motion.div {...popIn(0.1)}>
              <Feature
                icon="lock"
                title="Your storage, your rules"
                text="Target local disks, NAS, S3-compatible buckets — or mix them with policy-based tiers."
              />
            </motion.div>
          </div>
        </section>

        <section id="pipeline" className="relative mx-auto max-w-6xl px-4 pb-24 sm:px-6">
          <div
            className="absolute inset-0 -z-10"
            style={{
              background:
                "linear-gradient(to bottom, transparent, hsl(var(--surface-2)/0.3), hsl(var(--surface-2)/0.4))",
            }}
          />
          <div className="grid items-center gap-10 lg:grid-cols-2">
            <motion.div {...popIn(0)}>
              <h2 className="text-2xl font-semibold">A pipeline that respects your originals</h2>
              <ul className="mt-3 list-disc space-y-1 pl-5 text-white/75">
                <li>Non-destructive imports with sidecar metadata</li>
                <li>Duplicate detection with human-friendly review</li>
                <li>Versioned backups and automatic integrity checks</li>
              </ul>

              <Divider className="my-6 opacity-20" />

              <div className="flex flex-wrap items-center gap-3">
                <Tech logo="go">Go</Tech>
                <Tech logo="postgres">Postgres</Tech>
                <Tech logo="s3">S3-compatible</Tech>
                <Tech logo="nats">NATS</Tech>
                <Tech logo="next">Next.js</Tech>
              </div>
            </motion.div>

            <motion.div {...popIn(0.06)}>
              <Card className="card-surface ring-brand shadow-xl">
                <CardHeader className="flex items-center justify-between px-6 pt-6">
                  <span className="text-sm text-white/70">pipeline preview</span>
                  <div className="flex gap-1.5">
                    <Dot color="hsl(var(--brand-1))" />
                    <Dot color="hsl(var(--accent-amber))" />
                    <Dot color="hsl(var(--accent-purple))" />
                  </div>
                </CardHeader>
                <CardBody className="px-6 pb-6">
                  <div className="relative aspect-video overflow-hidden rounded-xl ring-1 ring-white/10">
                    <div
                      className="absolute inset-0"
                      style={{
                        background:
                          "linear-gradient(135deg, hsl(var(--brand-1)/0.25), hsl(var(--brand-2)/0.22))",
                      }}
                    />
                    <svg className="absolute inset-0 h-full w-full opacity-[0.12]" aria-hidden>
                      <defs>
                        <pattern
                          id="diag"
                          width="28"
                          height="28"
                          patternUnits="userSpaceOnUse"
                          patternTransform="rotate(45)"
                        >
                          <line
                            x1="0"
                            y1="0"
                            x2="0"
                            y2="28"
                            stroke="currentColor"
                            strokeWidth="1"
                          />
                        </pattern>
                      </defs>
                      <rect width="100%" height="100%" fill="url(#diag)" />
                    </svg>
                  </div>
                </CardBody>
              </Card>
            </motion.div>
          </div>
        </section>

        <section id="faq" className="relative mx-auto max-w-6xl px-4 pt-28 pb-28 sm:px-6">
          <div
            className="absolute inset-0 -z-10"
            style={{
              background:
                "radial-gradient(60% 60% at 50% 0%, hsl(var(--brand-1)/0.18), transparent 70%)",
            }}
          />
          <h2 className="mb-6 text-2xl font-semibold">FAQ</h2>
          <div className="grid gap-4 md:grid-cols-2">
            <motion.div {...popIn(0)}>
              <FaqQ q="Is my data locked in?">
                Never. Everything is stored in open formats you can move or re-index elsewhere.
              </FaqQ>
            </motion.div>
            <motion.div {...popIn(0.05)}>
              <FaqQ q="Will this handle RAW?">
                Yes - we keep originals intact and sidecar metadata separate.
              </FaqQ>
            </motion.div>
            <motion.div {...popIn(0.1)}>
              <FaqQ q="Can I bring my own auth?">
                Yep. GitHub + local are built-in; SSO adapters are pluggable.
              </FaqQ>
            </motion.div>
            <motion.div {...popIn(0.15)}>
              <FaqQ q="What about storage costs?">
                Point Lumium at storage you already own - local disks, NAS, or S3-compatible.
              </FaqQ>
            </motion.div>
          </div>
        </section>

        <footer className="border-t border-white/10">
          <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-6 px-4 py-10 text-sm text-white/60 sm:flex-row sm:px-6">
            <p>© {new Date().getFullYear()} Lumium. All rights reserved.</p>
            <div className="flex gap-6">
              <HLink href="#features" className="hover:text-[hsl(var(--brand-1))]">
                Features
              </HLink>
              <HLink href="#faq" className="hover:text-[hsl(var(--brand-1))]">
                FAQ
              </HLink>
              <HLink href="/auth/login" className="hover:text-[hsl(var(--brand-1))]">
                Log in
              </HLink>
            </div>
          </div>
        </footer>
      </main>
    </div>
  )
}

function Decor() {
  return (
    <div aria-hidden className="absolute inset-0 z-0 overflow-hidden">
      <div
        className="absolute inset-0"
        style={{
          background: `linear-gradient(135deg,
            hsl(var(--surface-0) / 0.92) 0%,
            hsl(var(--surface-0) / 0.88) 60%,
            hsl(var(--vignette) / 0.82) 100%
          )`,
        }}
      />

      {/* primary red blob */}
      <motion.div
        className="absolute -top-40 -left-44 h-[36rem] w-[36rem] rounded-full blur-[140px]"
        style={{
          background:
            "radial-gradient(50% 50% at 50% 50%, hsl(var(--brand-1)/0.55) 0%, hsl(var(--brand-1)/0) 70%)",
        }}
        initial={{ x: -80, y: -40, scale: 1 }}
        animate={{ x: 30, y: 20, scale: 1.06 }}
        transition={{ duration: 18, repeat: Infinity, repeatType: "mirror", ease: EASE_IO }}
      />

      {/* orange blob */}
      <motion.div
        className="absolute top-1/3 -right-48 h-[34rem] w-[34rem] rounded-full blur-[140px]"
        style={{
          background:
            "radial-gradient(50% 50% at 50% 50%, hsl(var(--brand-2)/0.45) 0%, hsl(var(--brand-2)/0) 70%)",
        }}
        initial={{ x: 20, y: 0, scale: 1 }}
        animate={{ x: -40, y: -30, scale: 1.08, rotate: 2 }}
        transition={{ duration: 22, repeat: Infinity, repeatType: "mirror", ease: EASE_IO }}
      />

      {/* amber underglow */}
      <motion.div
        className="absolute bottom-[-6rem] left-1/3 h-[28rem] w-[28rem] rounded-full blur-[120px]"
        style={{
          background:
            "radial-gradient(50% 50% at 50% 50%, hsl(var(--accent-amber)/0.18) 0%, hsl(var(--accent-amber)/0) 70%)",
        }}
        initial={{ x: -20, y: 0, scale: 0.95 }}
        animate={{ x: 30, y: -25, scale: 1.05 }}
        transition={{ duration: 20, repeat: Infinity, repeatType: "mirror", ease: EASE_IO }}
      />

      {/* grid */}
      <svg className="absolute inset-0 h-full w-full opacity-[0.06]" aria-hidden>
        <defs>
          <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
            <path d="M40 0H0V40" fill="none" stroke="currentColor" strokeWidth="0.75" />
          </pattern>
          <filter id="noise">
            <feTurbulence
              type="fractalNoise"
              baseFrequency="0.8"
              numOctaves="2"
              stitchTiles="stitch"
            />
            <feColorMatrix type="saturate" values="0" />
            <feComponentTransfer>
              <feFuncA type="linear" slope="0.03" />
            </feComponentTransfer>
          </filter>
        </defs>
        <rect width="100%" height="100%" fill="url(#grid)" />
        <rect width="100%" height="100%" fill="hsl(var(--accent-purple)/0.12)" />
        <rect width="100%" height="100%" filter="url(#noise)" />
      </svg>

      {/* breathing vignette */}
      <motion.div
        className="absolute inset-0 opacity-25"
        style={{
          background:
            "radial-gradient(70% 60% at 50% 30%, rgba(0,0,0,0) 0%, hsl(var(--vignette)/0.45) 70%)",
        }}
        initial={{ scale: 1 }}
        animate={{ scale: 1.03 }}
        transition={{ duration: 14, repeat: Infinity, repeatType: "mirror", ease: EASE_IO }}
      />
    </div>
  )
}

function SectionDecor() {
  return (
    <div aria-hidden className="pointer-events-none absolute inset-0 -z-10">
      <motion.div
        className="absolute inset-x-0 -top-10 h-56 blur-2xl"
        style={{ background: "linear-gradient(to bottom, hsl(var(--brand-1)/0.14), transparent)" }}
        initial={{ opacity: 0.6, y: -6 }}
        animate={{ opacity: 0.9, y: 0 }}
        transition={{ duration: 8, repeat: Infinity, repeatType: "mirror", ease: EASE_IO }}
      />
      <motion.div
        className="absolute inset-x-0 -bottom-10 h-56 blur-2xl"
        style={{ background: "linear-gradient(to top, hsl(var(--brand-2)/0.14), transparent)" }}
        initial={{ opacity: 0.6, y: 6 }}
        animate={{ opacity: 0.9, y: 0 }}
        transition={{
          duration: 8,
          repeat: Infinity,
          repeatType: "mirror",
          ease: EASE_IO,
          delay: 1,
        }}
      />
      <svg className="absolute inset-0 h-full w-full opacity-[0.06]" aria-hidden>
        <defs>
          <pattern id="grid2" width="32" height="32" patternUnits="userSpaceOnUse">
            <path d="M32 0H0V32" fill="none" stroke="currentColor" strokeWidth="0.65" />
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#grid2)" />
      </svg>
    </div>
  )
}

function Dot({ color }: { color: string }) {
  return <span className="inline-block h-2.5 w-2.5 rounded-full" style={{ background: color }} />
}

function Tech({
  logo,
  children,
}: {
  logo: "go" | "postgres" | "s3" | "nats" | "next"
  children: string
}) {
  const Icon = () => {
    switch (logo) {
      case "go":
        return (
          <svg viewBox="0 0 64 24" className="h-4 w-auto">
            <rect width="64" height="24" rx="4" fill="#00ADD8" />
          </svg>
        )
      case "postgres":
        return (
          <svg viewBox="0 0 24 24" className="h-4 w-4">
            <path fill="#336791" d="M12 2c5 0 9 3 9 8s-4 8-9 8-9-3-9-8 4-8 9-8Z" />
          </svg>
        )
      case "s3":
        return (
          <svg viewBox="0 0 24 24" className="h-4 w-4">
            <path fill="#FF9900" d="M4 6h16v12H4z" />
          </svg>
        )
      case "nats":
        return (
          <svg viewBox="0 0 24 24" className="h-4 w-4">
            <path fill="#27AAE1" d="M12 2 2 7v10l10 5 10-5V7z" />
          </svg>
        )
      case "next":
        return (
          <svg viewBox="0 0 24 24" className="h-4 w-4">
            <circle cx="12" cy="12" r="10" fill="#000" />
          </svg>
        )
      default:
        return null
    }
  }
  return (
    <Tooltip content={children}>
      <motion.div
        {...popIn(0)}
        className="ring-brand inline-flex items-center gap-2 rounded-lg border border-white/10 bg-white/5 px-2.5 py-1.5"
      >
        <Icon />
        <span className="text-sm text-white/80">{children}</span>
      </motion.div>
    </Tooltip>
  )
}

function Feature({
  icon,
  title,
  text,
}: {
  icon: "stack" | "eye" | "lock"
  title: string
  text: string
}) {
  const Icon = () => {
    switch (icon) {
      case "stack":
        return (
          <svg width="18" height="18" viewBox="0 0 24 24">
            <path fill="currentColor" d="M12 2 2 7l10 5 10-5z" />
            <path fill="currentColor" d="M2 17l10 5 10-5" />
            <path fill="currentColor" d="M2 12l10 5 10-5" />
          </svg>
        )
      case "eye":
        return (
          <svg width="18" height="18" viewBox="0 0 24 24">
            <path
              fill="currentColor"
              d="M3 12s3-6 9-6 9 6 9 6-3 6-9 6-9-6-9-6m9 3a3 3 0 1 0 0-6 3 3 0 0 0 0 6"
            />
          </svg>
        )
      case "lock":
        return (
          <svg width="18" height="18" viewBox="0 0 24 24">
            <path
              fill="currentColor"
              d="M12 1a3 3 0 0 1 3 3v1h3a2 2 0 0 1 2 2v12a4 4 0 0 1-4 4H8a4 4 0 0 1-4-4V7a2 2 0 0 1 2-2h3V4a3 3 0 0 1 3-3Z"
            />
          </svg>
        )
    }
  }
  return (
    <Card className="card-surface ring-brand shadow-xl">
      <CardBody className="p-6">
        <div
          className="mb-3 inline-flex h-10 w-10 items-center justify-center rounded-xl ring-1 ring-white/10"
          style={{
            background:
              "linear-gradient(135deg, hsl(var(--brand-1)/0.20), hsl(var(--brand-2)/0.20))",
          }}
        >
          <Icon />
        </div>
        <h3 className="mb-1 text-lg font-semibold">{title}</h3>
        <p className="text-sm leading-relaxed text-white/70">{text}</p>
      </CardBody>
    </Card>
  )
}

function FaqQ({ q, children }: { q: string; children: React.ReactNode }) {
  return (
    <Card className="card-surface">
      <CardHeader className="px-6 pt-6">
        <h3 className="font-medium">{q}</h3>
      </CardHeader>
      <CardBody className="px-6 pb-6 text-sm text-white/70">{children}</CardBody>
    </Card>
  )
}
