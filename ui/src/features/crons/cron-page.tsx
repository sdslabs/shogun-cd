import { useQuery } from "@tanstack/react-query"
import { ArchiveX, ArrowLeft, ArrowUpRight, CheckCircle2, CircleMinus, Clock3, History, LoaderCircle, Settings2, XCircle } from "lucide-react"
import { Link, Navigate, useNavigate, useParams, useSearchParams } from "react-router-dom"

import { EmptyState, ErrorState, PageBody } from "@/components/page"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAuth } from "@/features/auth/auth-provider"
import { cn, titleCase } from "@/lib/utils"
import { cronApi, cronDate, useCrons } from "./api"
import { CronActions, PermanentlyDeleteCron } from "./cron-actions"
import { CronStatus } from "./cron-status"
import { CronForm } from "./cron-form"

export function CronPage() {
  const { cronId } = useParams()
  const creating = cronId === undefined
  const id = Number(cronId)
  const { claims } = useAuth()
  const admin = claims?.role === "admin"
  const query = useCrons()
  const activeCron = query.data?.crons.find(item => item.id === id)
  const archived = useQuery({ queryKey: ["cron-detail", id], queryFn: () => cronApi.get(id), enabled: !creating && Number.isInteger(id) && id > 0 && query.isSuccess && !activeCron })
  const [params] = useSearchParams()
  const navigate = useNavigate()
  if (creating && !admin) return <Navigate to="/crons" replace />
  if (!creating && (!Number.isInteger(id) || id <= 0)) return <ErrorState title="Invalid cron" />
  if (query.isLoading) return <PageBody><Skeleton className="h-20" /><Skeleton className="mt-6 h-80" /></PageBody>
  if (query.isError) return <ErrorState error={query.error} retry={() => query.refetch()} />
  if (!creating && !activeCron && archived.isLoading) return <PageBody><Skeleton className="h-20" /><Skeleton className="mt-6 h-80" /></PageBody>
  if (!creating && !activeCron && archived.isError) return <ErrorState title="Could not load cron" error={archived.error} retry={() => archived.refetch()} />
  const cron = activeCron ?? archived.data
  const deleted = !!cron?.deleted_at
  const timezone = query.data?.timezone ?? "UTC"
  const settings = creating || (admin && !!cron && !deleted && params.get("tab") === "settings")

  return <div>
    <header className="border-b border-primary/10 bg-gradient-to-br from-primary/[.07] via-primary/[.015] to-transparent px-5 pt-5 sm:px-7 lg:px-9">
      <Button asChild variant="ghost" size="sm" className="-ml-2 mb-5"><Link to={deleted ? "/crons?view=deleted" : "/crons"}><ArrowLeft />{deleted ? "Deleted crons" : "Crons"}</Link></Button>
      <div className="flex flex-wrap items-start justify-between gap-5 pb-6">
        <div className="flex items-start gap-3.5">
          <span className="flex size-11 shrink-0 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm"><Clock3 className="size-5" /></span>
          <div><p className="text-[10px] font-semibold tracking-[.18em] text-muted-foreground uppercase">Scheduled execution</p><div className="mt-1 flex flex-wrap items-center gap-3"><h1 className="text-2xl font-semibold tracking-[-.035em]">{creating ? "Create cron" : cron?.name ?? `Cron #${id}`}</h1>{deleted ? <span className="inline-flex h-6 items-center gap-1.5 rounded-md border border-destructive/25 bg-destructive/10 px-2 text-[11px] font-semibold text-destructive"><ArchiveX className="size-3" aria-hidden="true" />Deleted</span> : null}</div><p className="mt-2 text-xs text-muted-foreground">Shogun timezone: <span className="font-mono text-foreground">{timezone}</span></p></div>
        </div>
        {cron ? <div className="flex items-center gap-2">{!deleted ? <CronStatus enabled={cron.enabled} /> : null}{admin && deleted ? <PermanentlyDeleteCron cron={cron} onDeleted={() => navigate("/crons?view=deleted")} /> : null}{admin && !settings && !deleted ? <CronActions cron={cron} onDeleted={() => navigate("/crons")} /> : null}</div> : null}
      </div>
      {!creating ? <nav aria-label="Cron sections" className="flex gap-6">
        <Link to={`/crons/${id}`} className={cn("flex items-center gap-2 border-b-2 py-3 text-xs font-medium", !settings ? "border-primary text-foreground" : "border-transparent text-muted-foreground")}><History className="size-3.5" />Run history</Link>
        {admin && cron && !deleted ? <Link to={`/crons/${id}?tab=settings`} className={cn("flex items-center gap-2 border-b-2 py-3 text-xs font-medium", settings ? "border-primary text-foreground" : "border-transparent text-muted-foreground")}><Settings2 className="size-3.5" />Settings</Link> : null}
      </nav> : null}
    </header>
    <PageBody>
      {settings ? <CronForm key={cron?.id ?? "new"} cron={cron} timezone={timezone} /> : <div className="space-y-6">
        {cron ? <div className="grid gap-4 sm:grid-cols-3"><Summary label="Pipeline"><Link className="flex items-center gap-2 hover:text-primary" to={`/pipelines/${encodeURIComponent(cron.pipeline)}`}>{cron.pipeline}<ArrowUpRight className="size-3.5" /></Link></Summary><Summary label="Schedule"><code className="font-mono text-sm">{cron.schedule}</code></Summary><Summary label={deleted ? "Status" : cron.enabled ? "Next run" : "Next scheduled time"}><span>{deleted ? "Deleted · no future runs" : cron.next_run ? cronDate(cron.next_run, timezone) : "Not available"}</span>{!cron.enabled && !deleted ? <p className="mt-1 text-xs font-normal text-muted-foreground">Paused · will not execute</p> : null}</Summary></div> : <p className="text-sm text-muted-foreground">This schedule is no longer available. Its recorded runs are retained below.</p>}
        {deleted && admin && cron?.payload ? <section className="rounded-xl border bg-card p-5"><h2 className="text-sm font-semibold">Saved runtime values</h2><p className="mt-1 text-xs text-muted-foreground">Settings retained when this cron was deleted.</p><dl className="mt-4 divide-y">{Object.entries(cron.payload).map(([key, value]) => <div key={key} className="grid gap-1 py-3 text-xs sm:grid-cols-2"><dt className="font-mono text-muted-foreground">{key}</dt><dd className="break-all font-mono">{value || "(empty)"}</dd></div>)}</dl>{!Object.keys(cron.payload).length ? <p className="mt-3 text-xs text-muted-foreground">No runtime values.</p> : null}</section> : null}
        <CronHistory id={id} timezone={timezone} />
      </div>}
    </PageBody>
  </div>
}

function Summary({ label, children }: { label: string; children: React.ReactNode }) {
  return <div className="rounded-xl border bg-card px-5 py-4"><p className="mb-2 text-[10px] font-semibold tracking-[.12em] text-muted-foreground uppercase">{label}</p><div className="text-sm font-medium">{children}</div></div>
}

function CronHistory({ id, timezone }: { id: number; timezone: string }) {
  const [params, setParams] = useSearchParams()
  const before = Number(params.get("before")) || 0
  const trail = (params.get("trail") ?? "").split(",").filter(Boolean).map(Number).filter(value => Number.isInteger(value) && value >= 0)
  const changePage = (cursor: number, previous: number[]) => {
    const next = new URLSearchParams(params)
    next.delete("tab")
    if (cursor > 0) next.set("before", String(cursor)); else next.delete("before")
    if (previous.length) next.set("trail", previous.join(",")); else next.delete("trail")
    setParams(next)
  }
  const runs = useQuery({ queryKey: ["cron-runs", id, before], queryFn: () => cronApi.runs(id, before), refetchInterval: 5_000 })
  return <section className="overflow-hidden rounded-xl border bg-card">
    <div className="border-b px-5 py-4"><h2 className="text-sm font-semibold">Run history</h2><p className="mt-1 text-xs text-muted-foreground">Recorded occurrences and their pipeline executions.</p></div>
    {runs.isLoading ? <Skeleton className="m-5 h-36" /> : runs.isError ? <ErrorState compact error={runs.error} retry={() => runs.refetch()} /> : !runs.data?.runs.length ? <EmptyState compact title="No recorded occurrences" description="Scheduled runs will appear here when this cron fires." /> : <Table>
      <TableHeader><TableRow><TableHead><span className="sr-only">Scheduled time · {timezone}</span></TableHead><TableHead>Result</TableHead><TableHead>Pipeline run</TableHead></TableRow></TableHeader>
      <TableBody>{runs.data.runs.map(run => <TableRow key={run.id}>
        <TableCell className="text-xs">{cronDate(run.scheduled_at, timezone)}</TableCell>
        <TableCell><CronRunStatus status={run.pipeline_run?.status ?? run.status} />{run.error ? <p className="mt-1 max-w-xl text-xs text-destructive">{run.error}</p> : null}</TableCell>
        <TableCell>{run.pipeline_run_id ? <Button asChild variant="ghost" size="sm"><Link to={`/pipelines/${encodeURIComponent(run.pipeline)}/runs/${run.pipeline_run_id}?cron=${id}`}>View pipeline run<ArrowUpRight /></Link></Button> : <span className="text-xs text-muted-foreground">{run.status === "dispatching" ? "Starting…" : "Did not start"}</span>}</TableCell>
      </TableRow>)}</TableBody>
    </Table>}
    <div className="flex flex-wrap items-center justify-between gap-3 border-t px-5 py-3">
      <span className="text-xs text-muted-foreground">{runs.data?.runs.length ?? 0} runs · 10 per page</span>
      <div className="flex items-center gap-2">
        {before > 0 ? <Button variant="ghost" size="sm" onClick={() => changePage(0, [])}>Latest runs</Button> : null}
        <Button variant="outline" size="sm" disabled={before === 0 || runs.isLoading} onClick={() => changePage(trail.at(-1) ?? 0, trail.slice(0, -1))}>Newer</Button>
        <Button variant="outline" size="sm" disabled={!runs.data?.has_more || runs.isLoading} onClick={() => changePage(runs.data!.runs.at(-1)!.id, [...trail, before])}>Older</Button>
      </div>
    </div>
  </section>
}

function CronRunStatus({ status }: { status: string }) {
  const running = status === "running" || status === "dispatching"
  const Icon = status === "succeeded" ? CheckCircle2 : status === "failed" ? XCircle : running ? LoaderCircle : CircleMinus
  return <span className={cn(
    "inline-flex h-8 items-center gap-2 rounded-lg border px-3 text-xs font-semibold",
    status === "succeeded" ? "border-emerald-600/25 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400" :
    status === "failed" ? "border-destructive/25 bg-destructive/10 text-destructive" :
    running ? "border-info/25 bg-info/10 text-info" :
    status === "interrupted" ? "border-warning/25 bg-warning/10 text-amber-800 dark:text-warning" :
    "border-border bg-muted/40 text-muted-foreground",
  )}>
    <Icon className={cn("size-4", running && "motion-safe:animate-spin")} aria-hidden="true" />
    {titleCase(status)}
  </span>
}
