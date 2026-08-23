import { ChevronRight, KeyRound, RefreshCw, Search, Server, Waypoints } from "lucide-react"
import { useDeferredValue, useMemo, useState } from "react"
import { useNavigate } from "react-router-dom"

import { EmptyState, ErrorState, PageBody, PageHeader } from "@/components/page"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { usePipelines } from "@/features/pipelines/hooks"
import { useTargets } from "@/features/targets/hooks"
import type { TargetSummary } from "@/lib/api/types"
import { pluralize, titleCase } from "@/lib/utils"

export function TargetsPage() {
  const navigate = useNavigate()
  const targetsQuery = useTargets()
  const pipelinesQuery = usePipelines()
  const [search, setSearch] = useState("")
  const deferredSearch = useDeferredValue(search.toLowerCase().trim())
  const targets = useMemo(() => [...(targetsQuery.data ?? [])].filter((target) => !deferredSearch || [target.name, target.type, target.host, target.user, target.access_secret].some((value) => value.toLowerCase().includes(deferredSearch))).sort((a, b) => a.name.localeCompare(b.name)), [deferredSearch, targetsQuery.data])

  const referenceCount = (targetName: string) => pipelinesQuery.data?.filter((pipeline) => pipeline.steps.some((step) => step.target && !step.target.includes("{{") && step.target.toLowerCase() === targetName.toLowerCase())).length ?? 0

  return (
    <div>
      <PageHeader eyebrow="Infrastructure" title="Targets" description="Servers and clusters available to pipeline steps. Definitions are indexed from Git and remain read-only here." actions={<Button variant="outline" size="sm" onClick={() => targetsQuery.refetch()} disabled={targetsQuery.isFetching}><RefreshCw className={targetsQuery.isFetching ? "animate-spin" : ""} />Refresh</Button>} />
      <PageBody>
        <div className="mb-5 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"><div className="flex items-center gap-2 text-xs text-muted-foreground"><Waypoints className="size-3.5" /><span>{pluralize(targetsQuery.data?.length ?? 0, "execution target")}</span></div><div className="relative sm:w-64"><Search className="absolute top-1/2 left-3 size-3.5 -translate-y-1/2 text-muted-foreground" /><Input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Find a target…" className="pl-9" /></div></div>
        <section className="overflow-hidden rounded-xl border bg-card">
          {targetsQuery.isLoading ? <TargetSkeleton /> : targetsQuery.isError ? <ErrorState error={targetsQuery.error} retry={() => targetsQuery.refetch()} compact /> : targets.length === 0 ? <EmptyState title={targetsQuery.data?.length ? "No targets match" : "No targets indexed"} description={targetsQuery.data?.length ? "Try a different search." : "Add a Target manifest to the configured Git repository."} compact /> : (
            <Table><TableHeader><TableRow className="hover:bg-transparent"><TableHead>Target</TableHead><TableHead>Type</TableHead><TableHead>Connection</TableHead><TableHead>Access secret</TableHead><TableHead>Static references</TableHead><TableHead className="w-12"><span className="sr-only">Open</span></TableHead></TableRow></TableHeader><TableBody>{targets.map((target) => <TargetRow target={target} references={referenceCount(target.name)} onOpen={() => navigate(`/targets/${encodeURIComponent(target.name)}`)} key={target.name} />)}</TableBody></Table>
          )}
        </section>
        <p className="mt-3 text-[11px] text-muted-foreground">Runtime target values supplied by webhooks cannot be inferred from pipeline definitions and are intentionally excluded from reference counts.</p>
      </PageBody>
    </div>
  )
}

function TargetRow({ target, references, onOpen }: { target: TargetSummary; references: number; onOpen: () => void }) {
  return <TableRow role="link" tabIndex={0} onClick={onOpen} onKeyDown={(event) => { if (event.key === "Enter" || event.key === " ") onOpen() }} className="group cursor-pointer"><TableCell><div className="flex items-center gap-3"><span className="flex size-8 items-center justify-center rounded-lg border bg-muted/35 text-muted-foreground"><Server className="size-3.5" /></span><span className="font-semibold">{target.name}</span></div></TableCell><TableCell><Badge variant={target.type === "server" ? "info" : "secondary"}>{titleCase(target.type)}</Badge></TableCell><TableCell><code className="text-xs text-muted-foreground"><span className="text-foreground">{target.user}</span>@{target.host}:{target.port}</code></TableCell><TableCell><span className="inline-flex items-center gap-1.5 text-xs text-muted-foreground"><KeyRound className="size-3" /><code className="text-foreground">{target.access_secret || "—"}</code></span></TableCell><TableCell><span className="text-xs">{pluralize(references, "pipeline")}</span></TableCell><TableCell><ChevronRight className="size-4 text-muted-foreground transition-transform group-hover:translate-x-0.5" /></TableCell></TableRow>
}

function TargetSkeleton() { return <div>{Array.from({ length: 5 }).map((_, index) => <div className="flex h-16 items-center gap-5 border-b px-4 last:border-0" key={index}><Skeleton className="size-8" /><Skeleton className="h-3 w-32" /><Skeleton className="ml-auto h-5 w-16" /><Skeleton className="h-3 w-52" /><Skeleton className="h-3 w-32" /><Skeleton className="h-3 w-20" /></div>)}</div> }
