import { cn } from "@/lib/utils"
import shogunMarkUrl from "@/assets/shogun-mark.svg"

export function ShogunMark({ className }: { className?: string }) {
  return (
    <span
      aria-hidden="true"
      className={cn("inline-block size-7 shrink-0 bg-current [mask-position:center] [mask-repeat:no-repeat] [mask-size:contain]", className)}
      style={{ WebkitMaskImage: `url(${shogunMarkUrl})`, maskImage: `url(${shogunMarkUrl})` }}
    />
  )
}

export function Brand({ compact = false }: { compact?: boolean }) {
  return (
    <div className="flex items-center gap-2.5">
      <span className="flex h-9 w-11 shrink-0 items-center justify-center rounded-lg bg-black text-primary shadow-sm">
        <ShogunMark className="h-6 w-9" />
      </span>
      {!compact ? (
        <span className="flex min-w-0 flex-col leading-none">
          <span className="truncate text-[15px] font-semibold tracking-[-0.02em]">Shogun</span>
          <span className="mt-1 text-[9px] font-semibold tracking-[0.2em] text-muted-foreground uppercase">Continuous delivery</span>
        </span>
      ) : null}
    </div>
  )
}
