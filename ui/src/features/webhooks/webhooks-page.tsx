import { zodResolver } from "@hookform/resolvers/zod"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Check, Copy, MoreHorizontal, Pause, Play, Plus, RefreshCw, Search, Trash2, Webhook as WebhookIcon } from "lucide-react"
import * as React from "react"
import { useForm } from "react-hook-form"
import { toast } from "sonner"
import { z } from "zod"

import { EmptyState, ErrorState, PageBody, PageHeader } from "@/components/page"
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from "@/components/ui/alert-dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useAuth } from "@/features/auth/auth-provider"
import { usePipelines } from "@/features/pipelines/hooks"
import { absoluteApiUrl } from "@/lib/api/client"
import { api } from "@/lib/api/endpoints"
import type { CreatedWebhook, Webhook } from "@/lib/api/types"
import { formatDateTime, formatRelativeTime } from "@/lib/format"
import { queryKeys } from "@/lib/query"

export function WebhooksPage() {
  const { claims } = useAuth()
  const isAdmin = claims?.role === "admin"
  const queryClient = useQueryClient()
  const pipelinesQuery = usePipelines()
  const hooksQuery = useQuery({ queryKey: queryKeys.webhooks(), queryFn: () => api.webhooks() })
  const [search, setSearch] = React.useState("")
  const [pipeline, setPipeline] = React.useState("all")
  const [createOpen, setCreateOpen] = React.useState(false)
  const [deleteTarget, setDeleteTarget] = React.useState<Webhook | null>(null)
  const deferredSearch = React.useDeferredValue(search.toLowerCase().trim())
  const hooks = React.useMemo(() => [...(hooksQuery.data ?? [])].filter((hook) => pipeline === "all" || hook.pipeline_name === pipeline).filter((hook) => !deferredSearch || [hook.alias, hook.slug, hook.pipeline_name, hook.created_by].some((value) => value.toLowerCase().includes(deferredSearch))).sort((a, b) => a.alias.localeCompare(b.alias)), [deferredSearch, hooksQuery.data, pipeline])

  const statusMutation = useMutation({ mutationFn: ({ hook, active }: { hook: Webhook; active: boolean }) => api.setWebhookStatus(hook.slug, active), onSuccess: async (_, variables) => { await queryClient.invalidateQueries({ queryKey: ["webhooks"] }); toast.success(`${variables.hook.alias} ${variables.active ? "resumed" : "paused"}`) }, onError: (error) => toast.error(error.message) })
  const deleteMutation = useMutation({ mutationFn: (hook: Webhook) => api.deleteWebhook(hook.slug), onSuccess: async (_, hook) => { setDeleteTarget(null); await queryClient.invalidateQueries({ queryKey: ["webhooks"] }); toast.success(`${hook.alias} deleted`) }, onError: (error) => toast.error(error.message) })

  return (
    <div>
      <PageHeader eyebrow="Ingress" title="Webhooks" description="Authenticated entry points that let CI systems start Shogun pipelines." actions={<><Button variant="outline" size="sm" onClick={() => hooksQuery.refetch()} disabled={hooksQuery.isFetching}><RefreshCw className={hooksQuery.isFetching ? "animate-spin" : ""} />Refresh</Button>{isAdmin ? <Button size="sm" onClick={() => setCreateOpen(true)}><Plus />New webhook</Button> : null}</>} />
      <PageBody>
        <div className="mb-5 flex flex-col gap-2 sm:flex-row sm:justify-end"><div className="relative sm:w-64"><Search className="absolute top-1/2 left-3 size-3.5 -translate-y-1/2 text-muted-foreground" /><Input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Find a webhook…" className="pl-9" /></div><Select value={pipeline} onValueChange={setPipeline}><SelectTrigger className="sm:w-48"><SelectValue placeholder="All pipelines" /></SelectTrigger><SelectContent><SelectItem value="all">All pipelines</SelectItem>{pipelinesQuery.data?.map((item) => <SelectItem value={item.name} key={item.name}>{item.name}</SelectItem>)}</SelectContent></Select></div>
        <section className="overflow-hidden rounded-xl border bg-card">
          {hooksQuery.isLoading ? <WebhookSkeleton /> : hooksQuery.isError ? <ErrorState error={hooksQuery.error} retry={() => hooksQuery.refetch()} compact /> : hooks.length === 0 ? <EmptyState title={hooksQuery.data?.length ? "No webhooks match" : "No webhooks configured"} description={isAdmin ? "Create a webhook to let an external CI system start a pipeline." : "No webhook entry points are visible on this instance."} action={isAdmin && !hooksQuery.data?.length ? <Button size="sm" onClick={() => setCreateOpen(true)}><Plus />Create webhook</Button> : undefined} compact /> : (
            <Table><TableHeader><TableRow className="hover:bg-transparent"><TableHead>Webhook</TableHead><TableHead>Pipeline</TableHead><TableHead>Endpoint</TableHead><TableHead>Created</TableHead><TableHead>Status</TableHead>{isAdmin ? <TableHead className="w-12"><span className="sr-only">Actions</span></TableHead> : null}</TableRow></TableHeader><TableBody>{hooks.map((hook) => <TableRow key={hook.slug}><TableCell><div className="flex items-center gap-3"><span className="flex size-8 items-center justify-center rounded-lg border bg-muted/35 text-muted-foreground"><WebhookIcon className="size-3.5" /></span><div><span className="font-semibold">{hook.alias}</span><p className="mt-1 text-[10px] text-muted-foreground">by {hook.created_by}</p></div></div></TableCell><TableCell><Badge variant="secondary">{hook.pipeline_name}</Badge></TableCell><TableCell><CopyValue value={absoluteApiUrl(`/hook/${hook.slug}`)} display={`/hook/${hook.slug}`} /></TableCell><TableCell><span className="text-xs">{formatRelativeTime(hook.created_at)}</span><p className="mt-1 text-[10px] text-muted-foreground" title={formatDateTime(hook.created_at)}>{formatDateTime(hook.created_at)}</p></TableCell><TableCell><Badge variant={hook.is_active ? "success" : "outline"}>{hook.is_active ? "Active" : "Paused"}</Badge></TableCell>{isAdmin ? <TableCell><DropdownMenu><DropdownMenuTrigger asChild><Button variant="ghost" size="icon-sm"><MoreHorizontal /><span className="sr-only">Webhook actions</span></Button></DropdownMenuTrigger><DropdownMenuContent align="end"><DropdownMenuItem onSelect={() => statusMutation.mutate({ hook, active: !hook.is_active })}>{hook.is_active ? <Pause /> : <Play />}{hook.is_active ? "Pause" : "Resume"}</DropdownMenuItem><DropdownMenuSeparator /><DropdownMenuItem variant="destructive" onSelect={() => setDeleteTarget(hook)}><Trash2 />Delete</DropdownMenuItem></DropdownMenuContent></DropdownMenu></TableCell> : null}</TableRow>)}</TableBody></Table>
          )}
        </section>
      </PageBody>
      <CreateWebhookDialog open={createOpen} onOpenChange={setCreateOpen} pipelines={pipelinesQuery.data?.map((item) => item.name) ?? []} />
      <AlertDialog open={Boolean(deleteTarget)} onOpenChange={(open) => !open && setDeleteTarget(null)}><AlertDialogContent><AlertDialogHeader><AlertDialogTitle>Delete {deleteTarget?.alias}?</AlertDialogTitle><AlertDialogDescription>This permanently removes the endpoint. CI systems using it will stop triggering the pipeline.</AlertDialogDescription></AlertDialogHeader><AlertDialogFooter><AlertDialogCancel>Keep webhook</AlertDialogCancel><AlertDialogAction onClick={() => deleteTarget && deleteMutation.mutate(deleteTarget)} disabled={deleteMutation.isPending}>{deleteMutation.isPending ? "Deleting…" : "Delete webhook"}</AlertDialogAction></AlertDialogFooter></AlertDialogContent></AlertDialog>
    </div>
  )
}

const webhookSchema = z.object({ pipeline: z.string().min(1, "Select a pipeline"), alias: z.string().min(3, "Use at least 3 characters").max(32, "Use 32 characters or fewer") })
type WebhookForm = z.infer<typeof webhookSchema>

function CreateWebhookDialog({ open, onOpenChange, pipelines }: { open: boolean; onOpenChange: (open: boolean) => void; pipelines: string[] }) {
  const queryClient = useQueryClient()
  const [created, setCreated] = React.useState<CreatedWebhook | null>(null)
  const { register, handleSubmit, setValue, watch, reset, formState: { errors } } = useForm<WebhookForm>({ resolver: zodResolver(webhookSchema), defaultValues: { pipeline: "", alias: "" } })
  const mutation = useMutation({ mutationFn: api.createWebhook, onSuccess: async (result) => { setCreated(result); await queryClient.invalidateQueries({ queryKey: ["webhooks"] }) } })
  const close = (next: boolean) => { if (!next) { reset(); setCreated(null); mutation.reset() } onOpenChange(next) }
  return <Dialog open={open} onOpenChange={close}><DialogContent>{created ? <SecretReveal created={created} onDone={() => close(false)} /> : <form onSubmit={handleSubmit((values) => mutation.mutate(values))}><DialogHeader><DialogTitle>Create webhook</DialogTitle><DialogDescription>Generate an authenticated endpoint for a pipeline. Its secret is shown exactly once.</DialogDescription></DialogHeader><div className="my-6 space-y-4"><label className="block"><span className="mb-2 block text-xs font-medium">Pipeline</span><Select value={watch("pipeline")} onValueChange={(value) => setValue("pipeline", value, { shouldValidate: true })}><SelectTrigger className="w-full"><SelectValue placeholder="Select a pipeline" /></SelectTrigger><SelectContent>{pipelines.map((pipeline) => <SelectItem value={pipeline} key={pipeline}>{pipeline}</SelectItem>)}</SelectContent></Select>{errors.pipeline ? <span className="mt-1.5 block text-xs text-destructive">{errors.pipeline.message}</span> : null}</label><label className="block"><span className="mb-2 block text-xs font-medium">Alias</span><Input placeholder="Production image webhook" {...register("alias")} />{errors.alias ? <span className="mt-1.5 block text-xs text-destructive">{errors.alias.message}</span> : null}</label>{mutation.error ? <p className="rounded-lg border border-destructive/20 bg-destructive/7 p-3 text-xs text-destructive">{mutation.error.message}</p> : null}</div><DialogFooter><Button type="button" variant="outline" onClick={() => close(false)}>Cancel</Button><Button type="submit" disabled={mutation.isPending}>{mutation.isPending ? "Creating…" : "Create webhook"}</Button></DialogFooter></form>}</DialogContent></Dialog>
}

function SecretReveal({ created, onDone }: { created: CreatedWebhook; onDone: () => void }) {
  return <><DialogHeader><DialogTitle>Webhook ready</DialogTitle><DialogDescription>Copy these values now. The authentication secret cannot be retrieved again.</DialogDescription></DialogHeader><div className="space-y-3"><RevealRow label="Endpoint" value={absoluteApiUrl(`/hook/${created.webhook.slug}`)} /><RevealRow label="Secret" value={created.secret} secret /></div><div className="rounded-lg border border-warning/25 bg-warning/8 px-3.5 py-3 text-xs leading-5 text-foreground">Store the secret in your CI provider before closing this dialog.</div><DialogFooter><Button onClick={onDone}>I’ve saved the secret</Button></DialogFooter></>
}

function RevealRow({ label, value, secret }: { label: string; value: string; secret?: boolean }) { const [copied, setCopied] = React.useState(false); return <div><p className="mb-1.5 text-[10px] font-semibold tracking-[.12em] text-muted-foreground uppercase">{label}</p><div className="flex items-center gap-2 rounded-lg border bg-muted/35 p-2"><code className="min-w-0 flex-1 break-all px-1 text-xs">{secret ? value : value}</code><Button variant="outline" size="icon-sm" onClick={async () => { await navigator.clipboard.writeText(value); setCopied(true); window.setTimeout(() => setCopied(false), 1_500) }}>{copied ? <Check className="text-success" /> : <Copy />}</Button></div></div> }
function CopyValue({ value, display }: { value: string; display: string }) { const [copied, setCopied] = React.useState(false); return <button type="button" className="group flex max-w-48 items-center gap-2 font-mono text-[11px] text-muted-foreground hover:text-foreground" onClick={async () => { await navigator.clipboard.writeText(value); setCopied(true); window.setTimeout(() => setCopied(false), 1_500) }}><span className="truncate">{display}</span>{copied ? <Check className="size-3 text-success" /> : <Copy className="size-3 opacity-0 group-hover:opacity-100" />}</button> }
function WebhookSkeleton() { return <div>{Array.from({ length: 5 }).map((_, index) => <div className="flex h-[69px] items-center gap-5 border-b px-4 last:border-0" key={index}><Skeleton className="size-8" /><Skeleton className="h-3 w-32" /><Skeleton className="ml-auto h-5 w-28" /><Skeleton className="h-3 w-40" /><Skeleton className="h-5 w-14" /></div>)}</div> }
