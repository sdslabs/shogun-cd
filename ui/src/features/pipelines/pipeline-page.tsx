import { useState } from "react"
import { ArrowLeft, Boxes, Braces, ChevronRight, FilePenLine, Files, GitBranch, Pause, Play, RefreshCw, Terminal, Timer } from "lucide-react"
import { Link, useNavigate, useParams } from "react-router-dom"

import { EmptyState, ErrorState, PageBody, PageHeader } from "@/components/page"
import { StatusSignal, StatusText } from "@/components/status-signal"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { usePipeline, usePipelineRuns } from "@/features/pipelines/hooks"
import { PipelineDefinition } from "@/features/pipelines/pipeline-definition"
import { RunPipelineDialog } from "@/features/pipelines/run-pipeline-dialog"
import { useAuth } from "@/features/auth/auth-provider"
import { formatDateTime, formatDuration, formatRelativeTime } from "@/lib/format"
import type { PipelineRun, PipelineStepSummary } from "@/lib/api/types"
import { cn, pluralize, titleCase } from "@/lib/utils"

export function PipelinePage() {
  const { pipeline: pipelineParam = "" } = useParams()
  const pipelineName = decodeURIComponent(pipelineParam)
  const navigate = useNavigate()
  const pipelineQuery = usePipeline(pipelineName)
  const runsQuery = usePipelineRuns(pipelineName)
  const { claims } = useAuth()
  const pipeline = pipelineQuery.pipeline
  const [configurationOpen, setConfigurationOpen] = useState(false)
  const [selectedDefinitionStep, setSelectedDefinitionStep] = useState(0)
  const [runDialogOpen, setRunDialogOpen] = useState(false)

  if (pipelineQuery.isLoading) return <PipelinePageSkeleton />
  if (pipelineQuery.isError) return <ErrorState error={pipelineQuery.error} retry={() => pipelineQuery.refetch()} />
  if (!pipeline) return <ErrorState title="Pipeline not found" error={new Error(`“${pipelineName}” is not present in the current Git index.`)} />

  const canRunFromUI = claims?.role === "admin" && pipeline.triggers.some((trigger) => trigger.type === "ui_trigger")

  return (
    <div>
      <PageHeader
        eyebrow="Pipeline"
        title={pipeline.name}
        description={`${pipeline.enabled ? "Enabled" : "Disabled"} · ${pluralize(pipeline.steps.length, "step")} · ${pluralize(pipeline.triggers.length, "trigger")}`}
        actions={
          <>
            <Button asChild variant="ghost" size="sm"><Link to="/pipelines"><ArrowLeft />All pipelines</Link></Button>
            <Button variant="outline" size="sm" onClick={() => runsQuery.refetch()} disabled={runsQuery.isFetching}><RefreshCw className={runsQuery.isFetching ? "animate-spin" : ""} />Refresh runs</Button>
            {canRunFromUI ? <Button size="sm" onClick={() => setRunDialogOpen(true)} disabled={!pipeline.enabled} title={pipeline.enabled ? undefined : "This pipeline is disabled"}>{pipeline.enabled ? <Play /> : <Pause />}Run pipeline</Button> : null}
          </>
        }
      />
      <PageBody className="space-y-7">
        <section className="overflow-hidden rounded-xl border bg-card">
          <div className="flex items-center justify-between gap-4 p-5 lg:px-6">
            <div>
              <p className="text-[10px] font-semibold tracking-[.16em] text-muted-foreground uppercase">Pipeline overview</p>
              <h2 className="mt-1.5 font-semibold tracking-[-0.02em]">Execution plan</h2>
            </div>
            <Badge variant={pipeline.enabled ? "success" : "outline"}>{pipeline.enabled ? "Enabled" : "Disabled"}</Badge>
          </div>
          <div className="grid border-t lg:grid-cols-[minmax(0,1fr)_18rem]">
            <div className="min-w-0 p-5 lg:p-6">
              <div className="mb-4">
                <p className="text-[10px] font-semibold tracking-[.14em] text-muted-foreground uppercase">Stage sequence</p>
                <p className="mt-1 text-xs text-muted-foreground">Select a stage to inspect its Git-defined configuration.</p>
              </div>
              <ExecutionPath
                steps={pipeline.steps}
                selectedStep={selectedDefinitionStep}
                configurationOpen={configurationOpen}
                onSelect={(step) => {
                  setSelectedDefinitionStep(step)
                  setConfigurationOpen(true)
                }}
              />
            </div>
            <div className="border-t p-5 lg:border-t-0 lg:border-l lg:p-6">
              <div className="mb-4 flex items-center gap-2 text-muted-foreground">
                <GitBranch className="size-3.5" />
                <p className="text-[10px] font-semibold tracking-[.14em] uppercase">Triggers</p>
              </div>
              <div className="space-y-3">
                {pipeline.triggers.map((trigger, index) => (
                  <div key={`${trigger.type}-${index}`}>
                    <Badge variant="outline">{titleCase(trigger.type)}</Badge>
                    {trigger.paths?.length ? <p className="mt-1.5 break-words font-mono text-[10px] leading-4 text-muted-foreground">{trigger.paths.join(" · ")}</p> : null}
                  </div>
                ))}
              </div>
            </div>
          </div>
          <PipelineDefinition
            steps={pipeline.steps}
            open={configurationOpen}
            selectedStep={selectedDefinitionStep}
            onOpenChange={setConfigurationOpen}
          />
          {canRunFromUI ? <RunPipelineDialog pipelineName={pipeline.name} open={runDialogOpen} onOpenChange={setRunDialogOpen} onStarted={(runID) => navigate(`/pipelines/${encodeURIComponent(pipeline.name)}/runs/${runID}`)} /> : null}
        </section>

        <section>
          <div className="mb-4 flex items-end justify-between"><div><p className="text-[10px] font-semibold tracking-[.16em] text-muted-foreground uppercase">History</p><h2 className="mt-1 text-lg font-semibold tracking-[-0.025em]">Past runs</h2></div><span className="text-xs text-muted-foreground">{pluralize(runsQuery.data?.length ?? 0, "execution")}</span></div>
          <div className="overflow-hidden rounded-xl border bg-card">
            {runsQuery.isLoading ? <RunTableSkeleton /> : runsQuery.isError ? <ErrorState error={runsQuery.error} retry={() => runsQuery.refetch()} compact /> : !runsQuery.data?.length ? <EmptyState title="No runs yet" description="This pipeline has not produced an execution record." compact /> : (
              <Table><TableHeader><TableRow className="hover:bg-transparent"><TableHead>Run</TableHead><TableHead>Result</TableHead><TableHead>Trigger</TableHead><TableHead>Started</TableHead><TableHead>Duration</TableHead><TableHead className="w-12"><span className="sr-only">Open</span></TableHead></TableRow></TableHeader><TableBody>{runsQuery.data.map((run) => <RunRow key={run.id} run={run} onOpen={() => navigate(`/pipelines/${encodeURIComponent(pipeline.name)}/runs/${run.id}`)} />)}</TableBody></Table>
            )}
          </div>
        </section>
      </PageBody>
    </div>
  )
}

const stepIcons = {
  mutate: FilePenLine,
  sync: Files,
  exec: Terminal,
  apply: Boxes,
}

function ExecutionPath({ steps, selectedStep, configurationOpen, onSelect }: { steps: PipelineStepSummary[]; selectedStep: number; configurationOpen: boolean; onSelect: (step: number) => void }) {
  return (
    <div className="overflow-x-auto pb-1">
      <div className="flex min-w-full items-center">
        {steps.map((step, position) => {
          const Icon = stepIcons[step.type as keyof typeof stepIcons] ?? Braces
          const selected = configurationOpen && step.index === selectedStep
          const target = step.target?.includes("{{") ? "Runtime target" : step.target || (step.type === "mutate" ? "Repository operation" : "No target")

          return (
            <div className="flex min-w-[10rem] flex-1 items-start" key={`${step.index}-${step.type}`}>
              <button
                type="button"
                aria-controls="pipeline-configuration"
                aria-expanded={selected}
                onClick={() => onSelect(step.index)}
                className={cn(
                  "group flex min-w-0 flex-1 items-center gap-3 rounded-lg p-2 text-left outline-none transition-colors hover:bg-muted/40 focus-visible:ring-2 focus-visible:ring-ring/50",
                  selected && "bg-primary/[.09]",
                )}
              >
                <span className={cn("flex size-9 shrink-0 items-center justify-center rounded-full border bg-background text-muted-foreground transition-colors group-hover:border-primary/35 group-hover:text-foreground", selected && "border-primary bg-primary text-primary-foreground")}>
                  <Icon className="size-4" />
                </span>
                <span className="min-w-0">
                  <span className="block font-mono text-[9px] font-semibold tracking-[.12em] text-muted-foreground uppercase">Step {String(step.index + 1).padStart(2, "0")}</span>
                  <span className="mt-0.5 block text-xs font-semibold">{titleCase(step.type)}</span>
                  <span className="mt-0.5 block truncate text-[9px] text-muted-foreground">{target}</span>
                </span>
              </button>
              {position < steps.length - 1 ? (
                <span
                  className={cn(
                    "relative mx-1 mt-[25px] h-px w-5 shrink-0 bg-border text-border transition-colors xl:w-8",
                    selected && "bg-primary/45 text-primary/45",
                  )}
                  aria-hidden="true"
                >
                  <span className="absolute -top-[2.5px] -right-px size-1.5 rotate-45 border-t border-r border-current" />
                </span>
              ) : null}
            </div>
          )
        })}
      </div>
    </div>
  )
}

function RunRow({ run, onOpen }: { run: PipelineRun; onOpen: () => void }) {
  return <TableRow role="link" tabIndex={0} onClick={onOpen} onKeyDown={(event) => { if (event.key === "Enter" || event.key === " ") onOpen() }} className="group cursor-pointer"><TableCell><div className="flex items-center gap-3"><StatusSignal status={run.status} size="sm" /><span className="font-mono text-xs font-semibold">#{run.id}</span></div></TableCell><TableCell><StatusText status={run.status} /></TableCell><TableCell><Badge variant="outline">{titleCase(run.trigger_kind)}</Badge></TableCell><TableCell><span className="text-xs font-medium">{formatRelativeTime(run.started_at)}</span><p className="mt-1 text-[10px] text-muted-foreground" title={formatDateTime(run.started_at)}>{formatDateTime(run.started_at)}</p></TableCell><TableCell><span className="flex items-center gap-1.5 text-xs"><Timer className="size-3 text-muted-foreground" />{formatDuration(run.started_at, run.finished_at)}</span></TableCell><TableCell><ChevronRight className="size-4 text-muted-foreground transition-transform group-hover:translate-x-0.5" /></TableCell></TableRow>
}

function PipelinePageSkeleton() { return <div><div className="border-b px-9 py-8"><Skeleton className="h-3 w-20" /><Skeleton className="mt-4 h-8 w-64" /><Skeleton className="mt-3 h-4 w-44" /></div><PageBody><Skeleton className="h-64 w-full rounded-xl" /><Skeleton className="mt-8 h-72 w-full rounded-xl" /></PageBody></div> }
function RunTableSkeleton() { return <div>{Array.from({ length: 5 }).map((_, index) => <div className="flex h-16 items-center gap-8 border-b px-4 last:border-0" key={index}><Skeleton className="size-5 rounded-full" /><Skeleton className="h-3 w-16" /><Skeleton className="h-5 w-24" /><Skeleton className="ml-auto h-3 w-28" /><Skeleton className="h-3 w-14" /></div>)}</div> }
