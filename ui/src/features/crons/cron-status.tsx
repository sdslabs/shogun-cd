import { Clock3, Pause } from "lucide-react"

import { cn } from "@/lib/utils"

export function CronStatus({ enabled }: { enabled: boolean }) {
  const Icon = enabled ? Clock3 : Pause
  return <span
    title={enabled ? "Schedule enabled. Future occurrences will start automatically." : "Schedule paused. Future occurrences will not start."}
    className={cn(
      "inline-flex h-8 shrink-0 items-center gap-2 rounded-lg border px-3 text-xs font-semibold",
      enabled ? "border-emerald-600/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400" : "border-warning/25 bg-warning/10 text-amber-800 dark:text-warning",
    )}
  >
    <Icon className="size-3.5" strokeWidth={2} aria-hidden="true" />
    {enabled ? "Enabled" : "Paused"}
  </span>
}
