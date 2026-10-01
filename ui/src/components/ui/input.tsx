import * as React from "react"

import { cn } from "@/lib/utils"

function Input({ className, type, ...props }: React.ComponentProps<"input">) {
  return (
    <input
      type={type}
      className={cn(
        "flex h-9 w-full min-w-0 rounded-md border border-input bg-background/75 px-3 py-1 text-sm shadow-sm outline-none transition-[border-color,box-shadow,background-color] placeholder:text-muted-foreground/75 focus:border-ring/70 focus:bg-background focus:ring-2 focus:ring-ring/20 disabled:pointer-events-none disabled:opacity-50",
        className,
      )}
      {...props}
    />
  )
}

export { Input }
