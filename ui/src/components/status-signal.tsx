import { Check, LoaderCircle, Minus, Pause, X } from "lucide-react"

import type { PipelineRunStatus, StepStatus } from "@/lib/api/types"
import { cn, titleCase } from "@/lib/utils"

type Status = PipelineRunStatus | StepStatus | "idle" | "disabled" | "unknown"

const config: Record<Status, { label: string; classes: string; icon: typeof Check }> = {
  succeeded: { label: "Succeeded", classes: "bg-success text-white dark:text-background", icon: Check },
  running: { label: "Running", classes: "bg-info text-white dark:text-background", icon: LoaderCircle },
  in_progress: { label: "In progress", classes: "bg-info text-white dark:text-background", icon: LoaderCircle },
  failed: { label: "Failed", classes: "bg-destructive text-white", icon: X },
  cancelled: { label: "Cancelled", classes: "bg-muted-foreground text-background", icon: Minus },
  queued: { label: "Queued", classes: "bg-warning text-background", icon: Pause },
  skipped: { label: "Skipped", classes: "bg-muted-foreground/65 text-background", icon: Minus },
  idle: { label: "No runs", classes: "bg-muted-foreground/45 text-background", icon: Minus },
  disabled: { label: "Disabled", classes: "bg-muted-foreground/40 text-background", icon: Pause },
  unknown: { label: "Unknown", classes: "bg-muted-foreground/45 text-background", icon: Minus },
}

export function StatusSignal({ status, size = "md", label, className }: { status?: Status | string | null; size?: "sm" | "md" | "lg"; label?: string; className?: string }) {
  const state = status && status in config ? (status as Status) : "unknown"
  const item = config[state]
  const Icon = item.icon
  return (
    <span className={cn("inline-flex items-center gap-2", className)} aria-label={label ?? item.label}>
      <span className="relative inline-flex shrink-0">
        {(state === "running" || state === "in_progress") ? <span className="absolute inset-0 rounded-full bg-info/70 motion-safe:animate-[signal-pulse_1.7s_ease-out_infinite]" /> : null}
        <span className={cn("relative inline-flex items-center justify-center rounded-full shadow-[0_0_0_3px_color-mix(in_oklab,currentColor_8%,transparent)]", item.classes, size === "sm" && "size-5", size === "md" && "size-7", size === "lg" && "size-10")}>
          <Icon className={cn(size === "sm" && "size-2.5", size === "md" && "size-3.5", size === "lg" && "size-5", (state === "running" || state === "in_progress") && "animate-spin")} strokeWidth={2.5} />
        </span>
      </span>
      {label !== undefined ? <span className="font-medium">{label || titleCase(state)}</span> : null}
    </span>
  )
}

export function StatusText({ status }: { status?: Status | string | null }) {
  const state = status && status in config ? (status as Status) : "unknown"
  const dotClasses: Record<Status, string> = {
    succeeded: "bg-success",
    running: "bg-info",
    in_progress: "bg-info",
    failed: "bg-destructive",
    cancelled: "bg-muted-foreground",
    queued: "bg-warning",
    skipped: "bg-muted-foreground",
    idle: "bg-muted-foreground/60",
    disabled: "bg-muted-foreground/60",
    unknown: "bg-muted-foreground/60",
  }
  return (
    <span className="inline-flex items-center gap-2 text-xs font-medium">
      <span className={cn("size-1.5 rounded-full", dotClasses[state])} />
      {config[state].label}
    </span>
  )
}
