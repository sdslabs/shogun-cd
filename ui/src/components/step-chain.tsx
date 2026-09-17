import { GitCommitHorizontal } from "lucide-react"

import { StatusSignal } from "@/components/status-signal"
import type { PipelineRunStep, PipelineStepSummary } from "@/lib/api/types"
import { cn, titleCase } from "@/lib/utils"

interface StepChainProps {
  steps: PipelineStepSummary[]
  runSteps?: PipelineRunStep[]
  selectedStep?: number | null
  onSelect?: (index: number | null) => void
  compact?: boolean
}

export function StepChain({ steps, runSteps, selectedStep, onSelect, compact = false }: StepChainProps) {
  const executedByIndex = new Map(runSteps?.map((step) => [step.step_index, step]))
  const interactive = Boolean(onSelect)

  if (compact) {
    return (
      <div className="flex items-center" aria-label={`${steps.length} pipeline steps`}>
        {steps.slice(0, 6).map((step, index) => {
          const runStep = executedByIndex.get(step.index)
          return (
            <div className="flex items-center" key={`${step.index}-${step.type}`}>
              {index > 0 ? <span className="h-px w-2 bg-border" /> : null}
              <span className={cn("size-2 rounded-full border bg-background", runStep?.status === "succeeded" && "border-success bg-success", runStep?.status === "failed" && "border-destructive bg-destructive", runStep?.status === "in_progress" && "border-info bg-info", !runStep && "border-muted-foreground/35")} title={`${index + 1}. ${titleCase(step.type)}`} />
            </div>
          )
        })}
        {steps.length > 6 ? <span className="ml-2 text-[10px] text-muted-foreground">+{steps.length - 6}</span> : null}
      </div>
    )
  }

  return (
    <div className="overflow-x-auto pb-2">
      <div className="flex min-w-max items-start py-2">
        {interactive ? <button type="button" onClick={() => onSelect?.(null)} className={cn("group flex w-24 flex-col items-center text-center outline-none", selectedStep == null && "text-primary")}>
          <span className={cn("flex size-9 items-center justify-center rounded-full border-2 bg-background transition-colors group-hover:border-primary/50", selectedStep == null ? "border-primary text-primary shadow-[0_0_0_4px_color-mix(in_oklab,var(--primary)_9%,transparent)]" : "border-border text-muted-foreground")}><GitCommitHorizontal className="size-4" /></span>
          <span className="mt-2 text-xs font-medium">All logs</span>
        </button> : null}
        {steps.map((step, index) => {
          const runStep = executedByIndex.get(step.index)
          const isSelected = selectedStep === step.index
          return (
            <div className="flex items-start" key={`${step.index}-${step.type}`}>
              {(interactive || index > 0) ? <span className={cn("mt-[17px] h-0.5 w-10 bg-border", runStep?.status === "succeeded" && "bg-success/45", runStep?.status === "failed" && "bg-destructive/40")} /> : null}
              <button type="button" disabled={!interactive} onClick={() => onSelect?.(step.index)} className={cn("group flex w-28 flex-col items-center text-center outline-none disabled:cursor-default", isSelected && "text-primary")}>
                <StatusSignal status={runStep?.status ?? "idle"} size="md" className={cn("rounded-full transition-shadow", isSelected && "shadow-[0_0_0_4px_color-mix(in_oklab,var(--primary)_12%,transparent)]")} />
                <span className="mt-2 text-xs font-semibold">{index + 1}. {titleCase(step.type)}</span>
                <span className="mt-0.5 max-w-24 truncate text-[10px] text-muted-foreground">{step.target?.includes("{{") ? "runtime target" : step.target || step.trigger_when || "pipeline"}</span>
              </button>
            </div>
          )
        })}
      </div>
    </div>
  )
}
