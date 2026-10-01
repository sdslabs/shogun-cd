import * as React from "react"

import { cn } from "@/lib/utils"

function Textarea({ className, ...props }: React.ComponentProps<"textarea">) {
  return (
    <textarea
      className={cn(
        "flex min-h-24 w-full rounded-md border border-input bg-background/75 px-3 py-2 text-sm shadow-sm outline-none transition-[border-color,box-shadow] placeholder:text-muted-foreground/75 focus:border-ring/70 focus:ring-2 focus:ring-ring/20 disabled:opacity-50",
        className,
      )}
      {...props}
    />
  )
}

export { Textarea }
