import type * as React from "react"
import { AlertTriangle, Inbox, RotateCw } from "lucide-react"

import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

export function PageHeader({ eyebrow, title, description, actions, className }: { eyebrow?: string; title: string; description?: string; actions?: React.ReactNode; className?: string }) {
  return (
    <header className={cn("section-reveal flex flex-col gap-5 border-b border-border/70 px-5 py-7 sm:px-7 lg:flex-row lg:items-end lg:justify-between lg:px-9", className)}>
      <div className="min-w-0">
        {eyebrow ? <p className="mb-2 text-[10px] font-semibold tracking-[0.18em] text-primary uppercase">{eyebrow}</p> : null}
        <h1 className="truncate text-2xl font-semibold tracking-[-0.035em] sm:text-[28px]">{title}</h1>
        {description ? <p className="mt-2 max-w-2xl text-sm leading-6 text-muted-foreground">{description}</p> : null}
      </div>
      {actions ? <div className="flex shrink-0 flex-wrap items-center gap-2">{actions}</div> : null}
    </header>
  )
}

export function PageBody({ className, ...props }: React.ComponentProps<"div">) {
  return <div className={cn("section-reveal-delayed px-5 py-6 sm:px-7 lg:px-9 lg:py-8", className)} {...props} />
}

export function EmptyState({ title, description, action, compact = false }: { title: string; description: string; action?: React.ReactNode; compact?: boolean }) {
  return (
    <div className={cn("flex flex-col items-center justify-center text-center", compact ? "min-h-48 px-5 py-8" : "min-h-[380px] px-6 py-12")}>
      <div className="mb-4 flex size-10 items-center justify-center rounded-xl border bg-muted/40 text-muted-foreground"><Inbox className="size-4" /></div>
      <h2 className="font-semibold tracking-[-0.015em]">{title}</h2>
      <p className="mt-1.5 max-w-sm text-sm leading-6 text-muted-foreground">{description}</p>
      {action ? <div className="mt-5">{action}</div> : null}
    </div>
  )
}

export function ErrorState({ title = "Couldn’t load this view", error, retry, compact = false }: { title?: string; error?: unknown; retry?: () => void; compact?: boolean }) {
  const message = error instanceof Error ? error.message : "An unexpected error occurred."
  return (
    <div className={cn("flex flex-col items-center justify-center text-center", compact ? "min-h-48 px-5 py-8" : "min-h-[380px] px-6 py-12")}>
      <div className="mb-4 flex size-10 items-center justify-center rounded-xl border border-destructive/20 bg-destructive/8 text-destructive"><AlertTriangle className="size-4" /></div>
      <h2 className="font-semibold tracking-[-0.015em]">{title}</h2>
      <p className="mt-1.5 max-w-md text-sm leading-6 text-muted-foreground">{message}</p>
      {retry ? <Button variant="outline" size="sm" onClick={retry} className="mt-5"><RotateCw />Try again</Button> : null}
    </div>
  )
}
