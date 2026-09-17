import type { PipelineStepSummary } from "@/lib/api/types"
import { titleCase } from "@/lib/utils"

const fieldOrder = ["trigger_when", "target", "files", "commands", "changes"]

export function PipelineDefinition({ steps }: { steps: PipelineStepSummary[] }) {
  return (
    <section>
      <div className="mb-4 flex items-end justify-between gap-4">
        <div>
          <p className="text-[10px] font-semibold tracking-[.16em] text-muted-foreground uppercase">Git definition</p>
          <h2 className="mt-1 text-lg font-semibold tracking-[-0.025em]">Step configuration</h2>
        </div>
        <span className="text-[11px] text-muted-foreground">Read-only · unresolved values</span>
      </div>
      <div className="divide-y overflow-hidden rounded-xl border bg-card">
        {steps.map((step) => <StepDefinition key={`${step.index}-${step.type}`} step={step} />)}
      </div>
    </section>
  )
}

function StepDefinition({ step }: { step: PipelineStepSummary }) {
  return (
    <article className="grid gap-5 p-5 lg:grid-cols-[180px_minmax(0,1fr)] lg:p-6">
      <div>
        <span className="font-mono text-[10px] font-semibold tracking-[.12em] text-primary uppercase">Step {step.index + 1}</span>
        <h3 className="mt-1.5 text-base font-semibold tracking-[-0.02em]">{titleCase(step.type)}</h3>
        <p className="mt-2 text-xs leading-5 text-muted-foreground">{step.target?.includes("{{") ? "Runtime target" : step.target || "Repository operation"}</p>
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
    <dl className="overflow-hidden rounded-lg border bg-background/35">
      {entries.map(([field, fieldValue]) => (
        <div className="grid border-b last:border-0 sm:grid-cols-[140px_minmax(0,1fr)]" key={field}>
          <dt className="bg-muted/30 px-3.5 py-3 text-[10px] font-semibold tracking-[.1em] text-muted-foreground uppercase">{field.replaceAll("_", " ")}</dt>
          <dd className="min-w-0 px-3.5 py-3"><DefinitionValue value={fieldValue} /></dd>
        </div>
      ))}
    </dl>
  )
}

function DefinitionValue({ value }: { value: unknown }) {
  if (value == null) return <span className="text-xs text-muted-foreground">—</span>
  if (Array.isArray(value)) {
    if (!value.length) return <span className="font-mono text-xs text-muted-foreground">[]</span>
    return (
      <div className="space-y-2">
        {value.map((item, index) => (
          <div className="flex min-w-0 gap-3" key={index}>
            <span className="mt-0.5 w-5 shrink-0 text-right font-mono text-[10px] text-muted-foreground">{index + 1}</span>
            <div className="min-w-0 flex-1">{isRecord(item) ? <DefinitionObject value={item} /> : <code className="block whitespace-pre-wrap break-words text-xs leading-5">{String(item)}</code>}</div>
          </div>
        ))}
      </div>
    )
  }
  if (isRecord(value)) return <DefinitionObject value={value} />
  return <code className="block whitespace-pre-wrap break-words text-xs leading-5">{String(value)}</code>
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value)
}
