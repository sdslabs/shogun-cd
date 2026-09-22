import { ArrowLeft, ExternalLink, GitBranch, KeyRound, Network, Server, Terminal, UserRound } from "lucide-react"
import { Link, useParams } from "react-router-dom"

import { ErrorState, PageBody, PageHeader } from "@/components/page"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { usePipelines } from "@/features/pipelines/hooks"
import { useTargets } from "@/features/targets/hooks"
import { pluralize, titleCase } from "@/lib/utils"

export function TargetPage() {
  const { target: targetParam = "" } = useParams()
  const targetName = decodeURIComponent(targetParam)
  const targetsQuery = useTargets()
  const pipelinesQuery = usePipelines()
  const target = targetsQuery.data?.find((item) => item.name.toLowerCase() === targetName.toLowerCase())

  if (targetsQuery.isLoading || pipelinesQuery.isLoading) return <TargetPageSkeleton />
  if (targetsQuery.isError) return <ErrorState error={targetsQuery.error} retry={() => targetsQuery.refetch()} />
  if (!target) return <ErrorState title="Target not found" error={new Error(`“${targetName}” is not present in the current Git index.`)} />

  const pipelines = (pipelinesQuery.data ?? []).map((pipeline) => ({ pipeline, steps: pipeline.steps.filter((step) => step.target && !step.target.includes("{{") && step.target.toLowerCase() === target.name.toLowerCase()) })).filter(({ steps }) => steps.length)

  return (
    <div>
      <PageHeader eyebrow="Execution target" title={target.name} description={`${titleCase(target.type)} · ${target.user}@${target.host}:${target.port}`} actions={<Button asChild variant="ghost" size="sm"><Link to="/targets"><ArrowLeft />All targets</Link></Button>} />
      <PageBody className="space-y-8">
        <section className="grid overflow-hidden rounded-xl border bg-card sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
          <Metric icon={Server} label="Type" value={titleCase(target.type)} />
          <Metric icon={Network} label="Host" value={target.host} />
          <Metric icon={UserRound} label="User" value={target.user} />
          <Metric icon={Terminal} label="Port" value={String(target.port)} />
          <Metric icon={KeyRound} label="Access secret" value={target.access_secret || "—"} />
        </section>

        <section>
          <div className="mb-4"><p className="text-[10px] font-semibold tracking-[.16em] text-muted-foreground uppercase">Static relationships</p><h2 className="mt-1 text-lg font-semibold tracking-[-0.025em]">Referenced by {pluralize(pipelines.length, "pipeline")}</h2></div>
          <div className="overflow-hidden rounded-xl border bg-card">
            {!pipelines.length ? <div className="px-5 py-12 text-center"><p className="font-medium">No static pipeline references</p><p className="mx-auto mt-2 max-w-md text-sm leading-6 text-muted-foreground">This target may still be selected through runtime webhook values. Shogun does not infer those relationships from definitions.</p></div> : pipelines.map(({ pipeline, steps }) => <Link to={`/pipelines/${encodeURIComponent(pipeline.name)}`} className="group flex items-center gap-4 border-b px-5 py-4 transition-colors last:border-0 hover:bg-muted/35" key={pipeline.name}><span className="flex size-8 items-center justify-center rounded-lg bg-muted text-muted-foreground"><GitBranch className="size-3.5" /></span><div className="min-w-0 flex-1"><p className="font-semibold">{pipeline.name}</p><div className="mt-1.5 flex flex-wrap gap-1.5">{steps.map((step) => <Badge variant="secondary" key={`${step.index}-${step.type}`}>Step {step.index + 1} · {titleCase(step.type)}</Badge>)}</div></div><ExternalLink className="size-4 text-muted-foreground group-hover:text-foreground" /></Link>)}
          </div>
        </section>
      </PageBody>
    </div>
  )
}

function Metric({ icon: Icon, label, value }: { icon: typeof Server; label: string; value: string }) { return <div className="border-b p-5 last:border-0 sm:border-r xl:border-b-0"><div className="flex items-center gap-2 text-[10px] font-semibold tracking-[.13em] text-muted-foreground uppercase"><Icon className="size-3.5" />{label}</div><p className="mt-3 truncate font-mono text-sm font-medium" title={value}>{value}</p></div> }
function TargetPageSkeleton() { return <div><div className="border-b px-9 py-8"><Skeleton className="h-3 w-24" /><Skeleton className="mt-4 h-8 w-56" /></div><PageBody><Skeleton className="h-32 rounded-xl" /><Skeleton className="mt-8 h-64 rounded-xl" /></PageBody></div> }
