import { useDeferredValue, useMemo, useState } from "react"
import { ArrowDownAZ, ChevronRight, Filter, RefreshCw, Search } from "lucide-react"
import { useNavigate } from "react-router-dom"

import { EmptyState, ErrorState, PageBody, PageHeader } from "@/components/page"
import { StatusSignal, StatusText } from "@/components/status-signal"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { usePipelines } from "@/features/pipelines/hooks"
import { formatDuration, formatRelativeTime } from "@/lib/format"
import type { PipelineSummary } from "@/lib/api/types"
import { pluralize, titleCase } from "@/lib/utils"

type EnabledFilter = "all" | "enabled" | "disabled"

export function PipelinesPage() {
  const navigate = useNavigate()
  const pipelinesQuery = usePipelines()
  const [search, setSearch] = useState("")
  const [enabledFilter, setEnabledFilter] = useState<EnabledFilter>("all")
  const deferredSearch = useDeferredValue(search.trim().toLowerCase())

  const pipelines = useMemo(() => {
    return [...(pipelinesQuery.data ?? [])]
      .filter((pipeline) => !deferredSearch || pipeline.name.toLowerCase().includes(deferredSearch) || pipeline.triggers.some((trigger) => trigger.type.toLowerCase().includes(deferredSearch)))
      .filter((pipeline) => enabledFilter === "all" || pipeline.enabled === (enabledFilter === "enabled"))
      .sort((a, b) => a.name.localeCompare(b.name))
  }, [deferredSearch, enabledFilter, pipelinesQuery.data])

  const counts = useMemo(() => ({ total: pipelinesQuery.data?.length ?? 0, enabled: pipelinesQuery.data?.filter((pipeline) => pipeline.enabled).length ?? 0 }), [pipelinesQuery.data])

  return (
    <div>
      <PageHeader
        eyebrow="Delivery workspace"
        title="Pipelines"
        description="Every delivery path indexed from your Git repository, ordered for predictable scanning."
        actions={<Button variant="outline" size="sm" onClick={() => pipelinesQuery.refetch()} disabled={pipelinesQuery.isFetching}><RefreshCw className={pipelinesQuery.isFetching ? "animate-spin" : ""} />Refresh</Button>}
      />
      <PageBody>
        <div className="mb-5 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div className="flex items-center gap-5 text-xs text-muted-foreground">
            <span><strong className="mr-1.5 text-base font-semibold text-foreground">{counts.total}</strong>indexed</span>
            <span><strong className="mr-1.5 text-base font-semibold text-foreground">{counts.enabled}</strong>enabled</span>
            <span><strong className="mr-1.5 text-base font-semibold text-foreground">{counts.total - counts.enabled}</strong>disabled</span>
          </div>
          <div className="flex flex-col gap-2 sm:flex-row">
            <div className="relative sm:w-64"><Search className="absolute top-1/2 left-3 size-3.5 -translate-y-1/2 text-muted-foreground" /><Input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Find a pipeline…" className="pl-9" /></div>
            <Select value={enabledFilter} onValueChange={(value) => setEnabledFilter(value as EnabledFilter)}><SelectTrigger className="sm:w-44"><Filter className="size-3.5 shrink-0 text-muted-foreground" /><SelectValue /></SelectTrigger><SelectContent><SelectItem value="all">All pipelines</SelectItem><SelectItem value="enabled">Enabled</SelectItem><SelectItem value="disabled">Disabled</SelectItem></SelectContent></Select>
          </div>
        </div>

        <section className="overflow-hidden rounded-xl border bg-card shadow-[0_1px_0_color-mix(in_oklab,var(--foreground)_3%,transparent)]">
          {pipelinesQuery.isLoading ? <PipelineTableSkeleton /> : pipelinesQuery.isError ? <ErrorState error={pipelinesQuery.error} retry={() => pipelinesQuery.refetch()} compact /> : pipelines.length === 0 ? <EmptyState title={pipelinesQuery.data?.length ? "No pipelines match" : "No pipelines indexed"} description={pipelinesQuery.data?.length ? "Adjust your search or filter to see more pipelines." : "Add a Pipeline manifest to the configured Git repository and wait for Shogun to index it."} compact /> : (
            <Table>
              <TableHeader><TableRow className="hover:bg-transparent"><TableHead className="w-[38%]">Pipeline</TableHead><TableHead>Latest signal</TableHead><TableHead>Last activity</TableHead><TableHead>Definition</TableHead><TableHead className="w-12"><span className="sr-only">Open</span></TableHead></TableRow></TableHeader>
              <TableBody>{pipelines.map((pipeline) => <PipelineRow pipeline={pipeline} key={pipeline.name} onOpen={() => navigate(`/pipelines/${encodeURIComponent(pipeline.name)}`)} />)}</TableBody>
            </Table>
          )}
        </section>
        <p className="mt-3 flex items-center gap-1.5 text-[11px] text-muted-foreground"><ArrowDownAZ className="size-3" />Sorted alphabetically · definitions are managed in Git</p>
      </PageBody>
    </div>
  )
}

function PipelineRow({ pipeline, onOpen }: { pipeline: PipelineSummary; onOpen: () => void }) {
  const latest = pipeline.last_run
  const status = !pipeline.enabled ? "disabled" : latest?.status ?? "idle"
  return (
    <TableRow tabIndex={0} role="link" onClick={onOpen} onKeyDown={(event) => { if (event.key === "Enter" || event.key === " ") onOpen() }} className={!pipeline.enabled ? "bg-muted/15 text-muted-foreground" : "group cursor-pointer"}>
      <TableCell className="py-3.5">
        <div className="flex items-center gap-3.5"><StatusSignal status={status} /><div className="min-w-0"><div className="flex items-center gap-2"><span className="truncate font-semibold tracking-[-0.012em] text-foreground">{pipeline.name}</span>{!pipeline.enabled ? <Badge variant="outline">Disabled</Badge> : null}</div><p className="mt-1 text-[11px] text-muted-foreground">{pluralize(pipeline.steps.length, "step")} · {pluralize(pipeline.triggers.length, "trigger")}</p></div></div>
      </TableCell>
      <TableCell>{latest ? <div><StatusText status={latest.status} /><p className="mt-1.5 text-[11px] text-muted-foreground">Run #{latest.id} · {titleCase(latest.trigger_kind)}</p></div> : <span className="text-xs text-muted-foreground">No executions yet</span>}</TableCell>
      <TableCell>{latest ? <div><span className="text-xs font-medium text-foreground">{formatRelativeTime(latest.started_at)}</span><p className="mt-1.5 text-[11px] text-muted-foreground">{formatDuration(latest.started_at, latest.finished_at)}</p></div> : <span className="text-xs text-muted-foreground">—</span>}</TableCell>
      <TableCell><div className="flex max-w-64 flex-wrap gap-1.5">{pipeline.steps.slice(0, 4).map((step) => <Badge variant="secondary" key={`${step.index}-${step.type}`}>{titleCase(step.type)}</Badge>)}{pipeline.steps.length > 4 ? <Badge variant="outline">+{pipeline.steps.length - 4}</Badge> : null}</div></TableCell>
      <TableCell><ChevronRight className="size-4 text-muted-foreground transition-transform group-hover:translate-x-0.5 group-hover:text-foreground" /></TableCell>
    </TableRow>
  )
}

function PipelineTableSkeleton() {
  return <div>{Array.from({ length: 6 }).map((_, index) => <div className="flex h-[73px] items-center gap-4 border-b px-4 last:border-0" key={index}><Skeleton className="size-7 rounded-full" /><div className="w-[34%] space-y-2"><Skeleton className="h-3.5 w-36" /><Skeleton className="h-2.5 w-24" /></div><Skeleton className="h-4 w-24" /><Skeleton className="ml-auto h-4 w-28" /><Skeleton className="h-5 w-44" /></div>)}</div>
}
