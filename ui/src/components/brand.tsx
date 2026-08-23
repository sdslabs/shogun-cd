import { cn } from "@/lib/utils"

export function ShogunMark({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 32 32"
      aria-hidden="true"
      className={cn("size-7", className)}
      fill="none"
    >
      <path
        d="M16 2.75 27.47 9.4v13.2L16 29.25 4.53 22.6V9.4L16 2.75Z"
        fill="currentColor"
        opacity=".14"
      />
      <path
        d="M16 3.75 26.6 9.9v12.2L16 28.25 5.4 22.1V9.9L16 3.75Z"
        stroke="currentColor"
        strokeWidth="1.5"
      />
      <path d="M9.2 11.4h13.6M11 15.8h10M13.15 20.4h5.7" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
      <path d="M16 8.7v13.6" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" opacity=".7" />
    </svg>
  )
}

export function Brand({ compact = false }: { compact?: boolean }) {
  return (
    <div className="flex items-center gap-2.5">
      <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-sm">
        <ShogunMark className="size-6" />
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
