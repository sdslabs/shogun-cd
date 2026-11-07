import { useQuery } from "@tanstack/react-query"
import { Archive, ArrowRight, Clock3, Plus } from "lucide-react"
import { Link, Navigate, useSearchParams } from "react-router-dom"

import { EmptyState, ErrorState, PageBody, PageHeader } from "@/components/page"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAuth } from "@/features/auth/auth-provider"
import { cronApi, cronDate, useCrons } from "./api"
import { CronActions } from "./cron-actions"
import { CronStatus } from "./cron-status"

export function CronsPage() {
  const { claims } = useAuth()
  const admin = claims?.role === "admin"
  const query = useCrons()
  const [params] = useSearchParams()
  const showDeleted = params.get("view") === "deleted"
  const deleted = useQuery({ queryKey: ["deleted-crons"], queryFn: cronApi.deleted, enabled: showDeleted })
  const legacyCron = Number(params.get("cron"))
  if (Number.isInteger(legacyCron) && legacyCron > 0) return <Navigate to={`/crons/${legacyCron}`} replace />
  if (query.isError) return <ErrorState error={query.error} retry={() => query.refetch()} />

  return <div>
    <PageHeader eyebrow="Automate" title="Crons" description="Schedule your pipelines. Open a cron to view its runs and manage its settings." actions={admin ? <Button asChild><Link to="/crons/new"><Plus />Create cron</Link></Button> : undefined} />
    <PageBody className="space-y-5">
      <nav aria-label="Cron lists" className="flex flex-wrap gap-2"><Button asChild variant={showDeleted ? "outline" : "secondary"} size="sm"><Link to="/crons">Active crons</Link></Button><Button asChild variant={showDeleted ? "secondary" : "outline"} size="sm"><Link to="/crons?view=deleted"><Archive />Deleted crons</Link></Button></nav>
      <div className="flex flex-wrap items-center justify-between gap-3 text-xs text-muted-foreground">
        <span className="flex items-center gap-2"><Clock3 className="size-3.5 text-primary" />Shogun timezone: <span className="font-mono text-foreground">{query.data?.timezone ?? "Loading…"}</span></span>
        <span>{showDeleted ? deleted.data?.length ?? 0 : query.data?.crons.length ?? 0} schedules</span>
      </div>
      <div className="overflow-hidden rounded-xl border bg-card">
        {showDeleted ? deleted.isLoading ? <Skeleton className="m-5 h-40" /> : deleted.isError ? <ErrorState compact error={deleted.error} retry={() => deleted.refetch()} /> : !deleted.data?.length ? <EmptyState compact title="No deleted crons" description="Deleted schedules and their history will appear here." /> : <Table>
          <TableHeader><TableRow><TableHead>Cron</TableHead><TableHead>Pipeline</TableHead><TableHead>Schedule</TableHead><TableHead>Deleted</TableHead><TableHead><span className="sr-only">Actions</span></TableHead></TableRow></TableHeader>
          <TableBody>{deleted.data.map(cron => <TableRow key={cron.id}>
            <TableCell><Link className="font-semibold hover:text-primary" to={`/crons/${cron.id}`}>{cron.name}</Link></TableCell>
            <TableCell className="font-mono text-xs"><Link className="text-muted-foreground hover:text-primary hover:underline" to={`/pipelines/${encodeURIComponent(cron.pipeline)}`}>{cron.pipeline}</Link></TableCell>
            <TableCell className="font-mono text-xs">{cron.schedule}</TableCell>
            <TableCell className="text-xs">{cron.deleted_at && query.data?.timezone ? cronDate(cron.deleted_at, query.data.timezone) : "—"}</TableCell>
            <TableCell><div className="flex items-center justify-end gap-2"><Button asChild variant="ghost" size="icon-sm"><Link to={`/crons/${cron.id}`} aria-label={`Open ${cron.name}`}><ArrowRight /></Link></Button></div></TableCell>
          </TableRow>)}</TableBody>
        </Table> : query.isLoading ? <Skeleton className="m-5 h-40" /> : !query.data?.crons.length ? <EmptyState title="No cron schedules" description="Create a schedule, choose a pipeline, and supply its runtime values." action={admin ? <Button asChild variant="outline"><Link to="/crons/new"><Plus />Create cron</Link></Button> : undefined} /> : <Table>
          <TableHeader><TableRow><TableHead>Cron</TableHead><TableHead>Pipeline</TableHead><TableHead>Schedule</TableHead><TableHead>Next run</TableHead><TableHead>Status</TableHead><TableHead><span className="sr-only">Actions</span></TableHead></TableRow></TableHeader>
          <TableBody>{[...query.data.crons].sort((a, b) => b.id - a.id).map(cron => <TableRow key={cron.id}>
            <TableCell><Link className="font-semibold hover:text-primary" to={`/crons/${cron.id}`}>{cron.name}</Link></TableCell>
            <TableCell className="font-mono text-xs"><Link className="text-muted-foreground hover:text-primary hover:underline" to={`/pipelines/${encodeURIComponent(cron.pipeline)}`}>{cron.pipeline}</Link></TableCell>
            <TableCell className="font-mono text-xs">{cron.schedule}</TableCell>
            <TableCell className="text-xs">{cron.next_run ? cronDate(cron.next_run, query.data.timezone) : "—"}{!cron.enabled && cron.next_run ? <span className="mt-1 block text-[10px] text-muted-foreground">Paused · preview only</span> : null}</TableCell>
            <TableCell><CronStatus enabled={cron.enabled} /></TableCell>
            <TableCell><div className="flex items-center justify-end gap-2">{admin ? <CronActions cron={cron} /> : null}<Button asChild variant="ghost" size="icon-sm"><Link to={`/crons/${cron.id}`} aria-label={`Open ${cron.name}`}><ArrowRight /></Link></Button></div></TableCell>
          </TableRow>)}</TableBody>
        </Table>}
      </div>
    </PageBody>
  </div>
}
