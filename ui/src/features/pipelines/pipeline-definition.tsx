import { Braces, ChevronDown } from "lucide-react"

import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible"
import type { PipelineStepSummary } from "@/lib/api/types"
import { cn, titleCase } from "@/lib/utils"

const fieldOrder = ["trigger_when", "target", "files", "commands", "changes"]

export function PipelineDefinition({ steps, open, selectedStep, onOpenChange }: { steps: PipelineStepSummary[]; open: boolean; selectedStep: number; onOpenChange: (open: boolean) => void }) {
  const selected = steps.find((step) => step.index === selectedStep) ?? steps[0]

  return (
    <Collapsible className="pipeline-definition border-t" open={open} onOpenChange={onOpenChange}>
      <CollapsibleTrigger asChild>
        <button type="button" className="flex w-full items-center gap-3 px-5 py-3.5 text-left outline-none transition-colors hover:bg-muted/25 focus-visible:bg-muted/25 lg:px-6" aria-controls="pipeline-configuration">
          <span className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground"><Braces className="size-3.5" /></span>
          <div className="min-w-0 flex-1">
            <h3 className="text-sm font-semibold tracking-[-0.015em]">Pipeline configuration</h3>
            <p className="mt-0.5 truncate text-[11px] text-muted-foreground">{selected ? `Step ${selected.index + 1} · ${titleCase(selected.type)}` : "Unresolved Git values"}</p>
          </div>
          <span className="hidden text-[10px] font-medium tracking-[.1em] text-muted-foreground uppercase sm:block">Read-only</span>
          <ChevronDown className={cn("size-4 text-muted-foreground transition-transform duration-200", open && "rotate-180")} />
        </button>
      </CollapsibleTrigger>
      {selected ? (
        <CollapsibleContent id="pipeline-configuration" className="pipeline-definition-content section-reveal overflow-hidden border-t bg-muted/[.07]">
          <StepDefinition step={selected} />
        </CollapsibleContent>
      ) : null}
    </Collapsible>
  )
}

function StepDefinition({ step }: { step: PipelineStepSummary }) {
  return (
    <article className="p-5 lg:px-6">
      <div className="mb-4 flex flex-wrap items-baseline gap-x-3 gap-y-1">
        <span className="font-mono text-[10px] font-semibold tracking-[.12em] text-primary uppercase">Step {step.index + 1}</span>
        <h3 className="text-sm font-semibold tracking-[-0.015em]">{titleCase(step.type)}</h3>
        <span className="text-[11px] text-muted-foreground">{step.target?.includes("{{") ? "Runtime target" : step.target || "Repository operation"}</span>
      </div>
      <DefinitionObject value={step.config} />
    </article>
  )
}

function DefinitionObject({ value }: { value: Record<string, unknown> }) {
  const entries = Object.entries(value).sort(([left], [right]) => {
    const leftIndex = fieldOrder.indexOf(left)
    const rightIndex = fieldOrder.indexOf(right)
    return (leftIndex === -1 ? fieldOrder.length : leftIndex) - (rightIndex === -1 ? fieldOrder.length : rightIndex) || left.localeCompare(right)
  })

  if (!entries.length) return <p className="text-xs text-muted-foreground">No configurable fields.</p>

  return (
    <dl className="divide-y overflow-hidden rounded-lg border bg-background/30">
      {entries.map(([field, fieldValue]) => (
        <div className="grid gap-2 px-3.5 py-3 sm:grid-cols-[120px_minmax(0,1fr)] sm:gap-4" key={field}>
          <dt className="text-[10px] font-semibold tracking-[.1em] text-muted-foreground uppercase">{field.replaceAll("_", " ")}</dt>
          <dd className="min-w-0"><DefinitionValue value={fieldValue} mutationChanges={field === "changes"} /></dd>
        </div>
      ))}
    </dl>
  )
}

function DefinitionValue({ value, mutationChanges = false }: { value: unknown; mutationChanges?: boolean }) {
  if (value == null) return <span className="text-xs text-muted-foreground">—</span>
  if (Array.isArray(value)) {
    if (!value.length) return <span className="font-mono text-xs text-muted-foreground">[]</span>
    return (
      <div className="divide-y overflow-hidden rounded-md border bg-background/25">
        {value.map((item, index) => (
          <div className="flex min-w-0 gap-3 px-3 py-2.5" key={index}>
            <span className="mt-0.5 w-4 shrink-0 text-right font-mono text-[9px] text-muted-foreground">{index + 1}</span>
            <div className="min-w-0 flex-1">{isRecord(item) ? <InlineRecord value={item} mutationChange={mutationChanges} /> : <code className="block whitespace-pre-wrap break-words text-xs leading-5">{String(item)}</code>}</div>
          </div>
        ))}
      </div>
    )
  }
  if (isRecord(value)) return <InlineRecord value={value} />
  return <code className="block whitespace-pre-wrap break-words text-xs leading-5">{String(value)}</code>
}

function InlineRecord({ value, mutationChange = false }: { value: Record<string, unknown>; mutationChange?: boolean }) {
  return (
    <div className="grid gap-x-5 gap-y-2 sm:grid-cols-2">
      {Object.entries(value).map(([field, fieldValue]) => (
        <div className={cn("min-w-0", mutationChange && field === "value" && "sm:col-span-2 sm:mt-0.5 sm:border-t sm:pt-2.5")} key={field}>
          <span className="block text-[9px] font-semibold tracking-[.08em] text-muted-foreground uppercase">{field.replaceAll("_", " ")}</span>
          <code className="mt-0.5 block break-words text-xs leading-5">{String(fieldValue)}</code>
        </div>
      ))}
    </div>
  )
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value)
}
