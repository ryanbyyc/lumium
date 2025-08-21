"use client"

import { Card, CardBody } from "@heroui/react"
import type { ReactNode } from "react"

export function FeatureCard({
  icon,
  title,
  children,
}: {
  icon: ReactNode
  title: string
  children: ReactNode
}) {
  return (
    <Card className="border border-white/10 bg-white/5 shadow-xl">
      <CardBody className="p-6">
        <div className="mb-3 inline-flex h-10 w-10 items-center justify-center rounded-xl bg-white/10">
          {icon}
        </div>
        <h3 className="mb-1 text-lg font-semibold">{title}</h3>
        <p className="text-sm leading-relaxed text-white/70">{children}</p>
      </CardBody>
    </Card>
  )
}
