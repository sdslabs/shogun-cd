import * as React from "react"
import { ArrowLeft, ChevronRight, Clock3, GitBranch, Radio, Timer } from "lucide-react"
import { Link, NavLink, useParams } from "react-router-dom"

import { ErrorState, PageBody } from "@/components/page"
import { StepChain } from "@/components/step-chain"
import { StatusSignal, StatusText } from "@/components/status-signal"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { LogViewer } from "@/features/pipelines/log-viewer"
import { usePipeline, usePipelineRun, usePipelineRuns } from "@/features/pipelines/hooks"
import { formatDateTime, formatDuration, formatRelativeTime } from "@/lib/format"
import { cn, titleCase } from "@/lib/utils"

export function RunPage() {
  const { pipeline: pipelineParam = "", runId: runIdParam = "" } = useParams()
  const pipelineName = decodeURIComponent(pipelineParam)
  const runId = Number(runIdParam)
  const pipelineQuery = usePipeline(pipelineName)
  const runQuery = usePipelineRun(pipelineName, runId)
  const runsQuery = usePipelineRuns(pipelineName)
  const [selection, setSelection] = React.useState<{ runId: number; step: number | null }>({ runId, step: null })
  const selectedStep = selection.runId === runId ? selection.step : null
  const setSelectedStep = (step: number | null) => setSelection({ runId, step })
  const run = runQuery.data

  if (!Number.isInteger(runId) || runId <= 0) return <ErrorState title="Invalid run number" error={new Error("Run IDs must be positive numbers.")} />
  if (runQuery.isLoading || pipelineQuery.isLoading) return <RunPageSkeleton />
  if (runQuery.isError) return <ErrorState error={runQuery.error} retry={() => runQuery.refetch()} />
  if (!run || !pipelineQuery.pipeline) return <ErrorState title="Run not found" error={new Error(`Run #${runId} is not available for ${pipelineName}.`)} />

  const pipeline = pipelineQuery.pipeline
  return (
    <div>
      <div className="border-b border-border/70 px-5 py-5 sm:px-7 lg:px-9">
        <Button asChild variant="ghost" size="sm" className="-ml-2 mb-4"><Link to={`/pipelines/${encodeURIComponent(pipeline.name)}`}><ArrowLeft />{pipeline.name}</Link></Button>
        <div className="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
          <div className="flex items-center gap-4"><StatusSignal status={run.status} size="lg" /><div><p className="text-[10px] font-semibold tracking-[.18em] text-primary uppercase">Execution</p><h1 className="mt-1 text-2xl font-semibold tracking-[-0.04em]">{pipeline.name} <span className="font-mono text-muted-foreground">#{run.id}</span></h1></div></div>
          <div className="flex flex-wrap items-center gap-x-6 gap-y-2 text-xs"><Meta icon={GitBranch} label={titleCase(run.trigger_kind)} /><Meta icon={Clock3} label={formatDateTime(run.started_at)} /><Meta icon={Timer} label={formatDuration(run.started_at, run.finished_at)} />{run.status === "running" ? <Badge variant="info"><Radio className="size-2.5" />Polling live state</Badge> : null}</div>
        </div>
      </div>

      <PageBody className="grid items-start gap-6 xl:grid-cols-[230px_minmax(0,1fr)]">
        <aside className="overflow-hidden rounded-xl border bg-card xl:sticky xl:top-[72px]">
          <div className="flex items-center justify-between border-b px-4 py-3"><div><p className="text-[10px] font-semibold tracking-[.14em] text-muted-foreground uppercase">Run switcher</p><p className="mt-1 text-xs font-medium">{runsQuery.data?.length ?? 0} executions</p></div><Button asChild variant="ghost" size="icon-sm"><Link to={`/pipelines/${encodeURIComponent(pipeline.name)}`} aria-label="View all runs"><ChevronRight /></Link></Button></div>
          <div className="max-h-[calc(100vh-230px)] overflow-y-auto p-2">
            {runsQuery.isLoading ? Array.from({ length: 5 }).map((_, index) => <Skeleton className="mb-2 h-14 w-full" key={index} />) : runsQuery.data?.map((item) => (
              <NavLink key={item.id} to={`/pipelines/${encodeURIComponent(pipeline.name)}/runs/${item.id}`} className={({ isActive }) => cn("mb-1 flex items-center gap-3 rounded-lg px-3 py-2.5 text-xs transition-colors hover:bg-muted", isActive && "bg-accent text-accent-foreground")}>
                <StatusSignal status={item.status} size="sm" /><span className="min-w-0 flex-1"><span className="block font-mono font-semibold">Run #{item.id}</span><span className="mt-1 block truncate text-[10px] text-muted-foreground">{formatRelativeTime(item.started_at)} · {formatDuration(item.started_at, item.finished_at)}</span></span>
              </NavLink>
            ))}
          </div>
        </aside>

        <div className="min-w-0 space-y-5">
          <section className="rounded-xl border bg-card p-5 sm:p-6">
            <div className="mb-4 flex items-center justify-between"><div><p className="text-[10px] font-semibold tracking-[.15em] text-muted-foreground uppercase">Stage timeline</p><h2 className="mt-1 font-semibold tracking-[-0.02em]">Select a stage to isolate its output</h2></div><StatusText status={run.status} /></div>
            <StepChain steps={pipeline.steps} runSteps={run.steps} selectedStep={selectedStep} onSelect={setSelectedStep} />
          </section>
          <LogViewer steps={run.steps} selectedStep={selectedStep} />
        </div>
      </PageBody>
    </div>
  )
}

function Meta({ icon: Icon, label }: { icon: typeof GitBranch; label: string }) { return <span className="flex items-center gap-1.5 text-muted-foreground"><Icon className="size-3.5" /><span className="text-foreground">{label}</span></span> }
function RunPageSkeleton() { return <div><div className="border-b px-9 py-7"><Skeleton className="h-4 w-24" /><Skeleton className="mt-5 h-9 w-72" /></div><PageBody className="grid gap-6 xl:grid-cols-[230px_1fr]"><Skeleton className="h-[500px] rounded-xl" /><div><Skeleton className="h-40 rounded-xl" /><Skeleton className="mt-5 h-[500px] rounded-xl" /></div></PageBody></div> }
