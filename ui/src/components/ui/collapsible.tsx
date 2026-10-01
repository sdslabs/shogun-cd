import * as React from "react"
import { Collapsible as CollapsiblePrimitive } from "radix-ui"

const Collapsible = CollapsiblePrimitive.Root
const CollapsibleTrigger = CollapsiblePrimitive.Trigger

function CollapsibleContent({ className, ...props }: React.ComponentProps<typeof CollapsiblePrimitive.Content>) {
  return <CollapsiblePrimitive.Content className={className} {...props} />
}

export { Collapsible, CollapsibleContent, CollapsibleTrigger }
