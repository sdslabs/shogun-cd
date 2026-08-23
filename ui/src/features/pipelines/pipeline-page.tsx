import { ArrowLeft, ChevronRight, GitBranch, RefreshCw, Route, Timer, Workflow } from "lucide-react"
import { Link, useNavigate, useParams } from "react-router-dom"

import { EmptyState, ErrorState, PageBody, PageHeader } from "@/components/page"
import { StepChain } from "@/components/step-chain"
import { StatusSignal, StatusText } from "@/components/status-signal"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { usePipeline, usePipelineRuns } from "@/features/pipelines/hooks"
import { formatDateTime, formatDuration, formatRelativeTime } from "@/lib/format"
import type { PipelineRun } from "@/lib/api/types"
import { pluralize, titleCase } from "@/lib/utils"

export function PipelinePage() {
  const { pipeline: pipelineParam = "" } = useParams()
  const pipelineName = decodeURIComponent(pipelineParam)
  const navigate = useNavigate()
  const pipelineQuery = usePipeline(pipelineName)
  const runsQuery = usePipelineRuns(pipelineName)
  const pipeline = pipelineQuery.pipeline

  if (pipelineQuery.isLoading) return <PipelinePageSkeleton />
  if (pipelineQuery.isError) return <ErrorState error={pipelineQuery.error} retry={() => pipelineQuery.refetch()} />
  if (!pipeline) return <ErrorState title="Pipeline not found" error={new Error(`“${pipelineName}” is not present in the current Git index.`)} />

  return (
    <div>
      <PageHeader
        eyebrow="Pipeline"
        title={pipeline.name}
        description={`${pipeline.enabled ? "Enabled" : "Disabled"} · ${pluralize(pipeline.steps.length, "step")} · ${pluralize(pipeline.triggers.length, "trigger")}`}
        actions={<><Button asChild variant="ghost" size="sm"><Link to="/pipelines"><ArrowLeft />All pipelines</Link></Button><Button variant="outline" size="sm" onClick={() => runsQuery.refetch()} disabled={runsQuery.isFetching}><RefreshCw className={runsQuery.isFetching ? "animate-spin" : ""} />Refresh runs</Button></>}
      />
      <PageBody className="space-y-7">
        <section className="grid overflow-hidden rounded-xl border bg-card lg:grid-cols-[minmax(0,1.55fr)_minmax(260px,.75fr)]">
          <div className="min-w-0 border-b p-5 lg:border-r lg:border-b-0 lg:p-6">
            <div className="mb-5 flex items-center justify-between"><div><p className="text-[10px] font-semibold tracking-[.16em] text-muted-foreground uppercase">Execution path</p><h2 className="mt-1.5 font-semibold tracking-[-0.02em]">Pipeline structure</h2></div><Badge variant={pipeline.enabled ? "success" : "outline"}>{pipeline.enabled ? "Enabled" : "Disabled"}</Badge></div>
            <StepChain steps={pipeline.steps} />
          </div>
          <div className="grid content-start divide-y">
            <DetailBlock icon={GitBranch} label="Triggers"><div className="space-y-2">{pipeline.triggers.map((trigger, index) => <div key={`${trigger.type}-${index}`}><Badge variant="outline">{titleCase(trigger.type)}</Badge>{trigger.paths?.length ? <p className="mt-1.5 break-words font-mono text-[10px] leading-4 text-muted-foreground">{trigger.paths.join(" · ")}</p> : null}</div>)}</div></DetailBlock>
            <DetailBlock icon={Route} label="Target behavior"><p className="text-xs leading-5 text-muted-foreground">Targets are resolved per step. Runtime webhook values are shown only during execution.</p></DetailBlock>
            <DetailBlock icon={Workflow} label="Source of truth"><p className="text-xs text-muted-foreground">Read-only · managed in Git</p></DetailBlock>
          </div>
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

function DetailBlock({ icon: Icon, label, children }: { icon: typeof GitBranch; label: string; children: React.ReactNode }) {
  return <div className="flex gap-3 p-5"><span className="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground"><Icon className="size-3.5" /></span><div className="min-w-0"><p className="mb-2 text-[10px] font-semibold tracking-[.12em] text-muted-foreground uppercase">{label}</p>{children}</div></div>
}

function RunRow({ run, onOpen }: { run: PipelineRun; onOpen: () => void }) {
  return <TableRow role="link" tabIndex={0} onClick={onOpen} onKeyDown={(event) => { if (event.key === "Enter" || event.key === " ") onOpen() }} className="group cursor-pointer"><TableCell><div className="flex items-center gap-3"><StatusSignal status={run.status} size="sm" /><span className="font-mono text-xs font-semibold">#{run.id}</span></div></TableCell><TableCell><StatusText status={run.status} /></TableCell><TableCell><Badge variant="outline">{titleCase(run.trigger_kind)}</Badge></TableCell><TableCell><span className="text-xs font-medium">{formatRelativeTime(run.started_at)}</span><p className="mt-1 text-[10px] text-muted-foreground" title={formatDateTime(run.started_at)}>{formatDateTime(run.started_at)}</p></TableCell><TableCell><span className="flex items-center gap-1.5 text-xs"><Timer className="size-3 text-muted-foreground" />{formatDuration(run.started_at, run.finished_at)}</span></TableCell><TableCell><ChevronRight className="size-4 text-muted-foreground transition-transform group-hover:translate-x-0.5" /></TableCell></TableRow>
}

function PipelinePageSkeleton() { return <div><div className="border-b px-9 py-8"><Skeleton className="h-3 w-20" /><Skeleton className="mt-4 h-8 w-64" /><Skeleton className="mt-3 h-4 w-44" /></div><PageBody><Skeleton className="h-64 w-full rounded-xl" /><Skeleton className="mt-8 h-72 w-full rounded-xl" /></PageBody></div> }
function RunTableSkeleton() { return <div>{Array.from({ length: 5 }).map((_, index) => <div className="flex h-16 items-center gap-8 border-b px-4 last:border-0" key={index}><Skeleton className="size-5 rounded-full" /><Skeleton className="h-3 w-16" /><Skeleton className="h-5 w-24" /><Skeleton className="ml-auto h-3 w-28" /><Skeleton className="h-3 w-14" /></div>)}</div> }
