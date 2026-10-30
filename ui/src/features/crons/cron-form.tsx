import { zodResolver } from "@hookform/resolvers/zod"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { AlertCircle, ArrowUpRight, Clock3, Plus, Save, Trash2 } from "lucide-react"
import { Controller, useFieldArray, useForm } from "react-hook-form"
import { Link, useNavigate } from "react-router-dom"
import { toast } from "sonner"
import { z } from "zod"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { usePipelines } from "@/features/pipelines/hooks"
import { cronApi, cronDate, type CronDefinition, type CronInput } from "./api"

const schema = z.object({
  name: z.string().trim().min(1, "Give this cron a name.").max(100),
  pipeline: z.string().min(1, "Select a pipeline."),
  schedule: z.string().trim().refine(value => value.split(/\s+/).length === 5 && !value.includes("="), "Use five cron fields: minute hour day month weekday."),
  enabled: z.boolean(),
  values: z.array(z.object({
    key: z.string().trim().refine(key => key === "" || /^[A-Za-z_][A-Za-z0-9_]*$/.test(key), "Use letters, numbers, and underscores; start with a letter or underscore."),
    value: z.string(),
  })),
}).superRefine(({ values }, context) => {
  const seen = new Set<string>()
  values.forEach(({ key }, index) => {
    if (!key) return
    if (seen.has(key)) context.addIssue({ code: "custom", path: ["values", index, "key"], message: "Each value name must be unique." })
    seen.add(key)
  })
})
type FormValues = z.infer<typeof schema>
const presets = [{ label: "Hourly", value: "0 * * * *" }, { label: "Daily", value: "0 2 * * *" }, { label: "Weekdays", value: "0 2 * * 1-5" }]

export function CronForm({ cron, timezone }: { cron?: CronDefinition; timezone: string }) {
  const pipelines = usePipelines()
  const client = useQueryClient()
  const navigate = useNavigate()
  const form = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: {
    name: cron?.name ?? "", pipeline: cron?.pipeline ?? "", schedule: cron?.schedule ?? "0 2 * * *", enabled: cron?.enabled ?? true,
    values: Object.entries(cron?.payload ?? {}).map(([key, value]) => ({ key, value })),
  } })
  const values = useFieldArray({ control: form.control, name: "values" })
  const runtimeValues = form.watch("values")
  const emptyNameCount = runtimeValues.filter(({ key }) => key.trim() === "").length
  const canAddValue = emptyNameCount < 5
  const selectedName = form.watch("pipeline")
  const schedule = form.watch("schedule")
  const selected = pipelines.data?.find(pipeline => pipeline.name === selectedName)
  const supportsCron = selected?.triggers.some(trigger => trigger.type === "cron_trigger")
  const canSavePipeline = !!supportsCron || (!!cron && selectedName === cron.pipeline)
  const hasEligible = pipelines.data?.some(pipeline => pipeline.triggers.some(trigger => trigger.type === "cron_trigger"))
  const mutation = useMutation({
    mutationFn: (input: CronInput) => cronApi.save(cron?.id, input),
    onSuccess: async (saved, input) => {
      form.reset({ name: input.name, pipeline: input.pipeline, schedule: input.schedule, enabled: input.enabled, values: Object.entries(input.payload).map(([key, value]) => ({ key, value })) })
      await client.invalidateQueries({ queryKey: ["crons"] })
      toast.success(cron ? "Cron settings saved" : "Cron created")
      if (!cron) navigate(`/crons/${saved.id}?tab=settings`, { replace: true })
    },
  })

  return <form onSubmit={form.handleSubmit(input => {
    if (!canSavePipeline) return
    mutation.mutate({ name: input.name, pipeline: input.pipeline, schedule: input.schedule, enabled: input.enabled, payload: Object.fromEntries(input.values.filter(({ key }) => key !== "").map(({ key, value }) => [key, value])) })
  })} className="grid items-start gap-6 xl:grid-cols-[minmax(0,1fr)_280px]">
    <div className="min-w-0 overflow-hidden rounded-xl border bg-card">
      <div className="border-b bg-muted/20 px-6 py-5"><p className="text-[10px] font-semibold tracking-[.16em] text-muted-foreground uppercase">Configuration</p><h2 className="mt-1 text-base font-semibold">{cron ? "Cron settings" : "Set up your schedule"}</h2><p className="mt-2 text-xs leading-5 text-muted-foreground">Choose what to run, when it runs, and the values it receives.</p></div>
      <div className="divide-y">
        <section className="space-y-5 px-6 py-6">
          <label className="block"><span className="mb-2 block text-xs font-semibold">Cron name</span><Input className="bg-muted/30 focus:bg-muted/40" autoComplete="off" placeholder="Nightly maintenance" {...form.register("name")} aria-invalid={!!form.formState.errors.name} />{form.formState.errors.name ? <FieldError message={form.formState.errors.name.message} /> : null}</label>
          <div>
            <label id="cron-pipeline-label" className="mb-2 block text-xs font-semibold">Pipeline</label>
            <Controller control={form.control} name="pipeline" render={({ field }) => <Select value={field.value} onValueChange={field.onChange} disabled={pipelines.isLoading}>
              <SelectTrigger aria-labelledby="cron-pipeline-label" className="w-full min-w-0 bg-muted/30 [&>span]:truncate [&>span]:text-left [&>svg]:shrink-0"><SelectValue placeholder={pipelines.isLoading ? "Loading pipelines…" : "Select pipeline"} /></SelectTrigger>
              <SelectContent className="w-[var(--radix-select-trigger-width)]">
                {cron && !pipelines.data?.some(pipeline => pipeline.name === cron.pipeline) ? <SelectItem value={cron.pipeline}>{cron.pipeline} · unavailable</SelectItem> : null}
                {pipelines.data?.map(pipeline => <SelectItem key={pipeline.name} value={pipeline.name}><span className="font-mono text-xs">{pipeline.name}</span><span className={`ml-3 text-[10px] ${pipeline.triggers.some(trigger => trigger.type === "cron_trigger") ? "text-muted-foreground" : "text-destructive"}`}>{!pipeline.triggers.some(trigger => trigger.type === "cron_trigger") ? "Requires cron_trigger" : !pipeline.enabled ? "Disabled" : ""}</span></SelectItem>)}
                {pipelines.isSuccess && !pipelines.data.length && !cron ? <p className="p-3 text-xs text-muted-foreground">No indexed pipelines.</p> : null}
              </SelectContent>
            </Select>} />
            {form.formState.errors.pipeline ? <FieldError message={form.formState.errors.pipeline.message} /> : null}
            {pipelines.isError ? <div role="alert" className="mt-3 flex items-center justify-between gap-3 text-xs text-destructive"><span>Could not load pipelines.</span><Button type="button" size="sm" variant="outline" onClick={() => pipelines.refetch()}>Try again</Button></div> : null}
            {pipelines.isSuccess && ((!hasEligible && !selectedName) || (selectedName && !supportsCron)) ? <div className="mt-3 rounded-lg border border-primary/20 bg-primary/5 p-3 text-xs leading-5">
              <p className="flex items-center gap-2 font-semibold"><AlertCircle className="size-3.5 shrink-0" />{selectedName ? `${selectedName} ${selected ? "needs cron_trigger" : "is no longer indexed"}` : pipelines.data.length ? "No pipelines declare cron_trigger yet" : "No pipelines have been indexed yet"}</p>
              <p className="mt-1 text-muted-foreground">{selected ? "Add the trigger below to this pipeline in Git. It will become available after Shogun indexes the change." : selectedName ? "Existing settings can still be saved. Scheduled attempts will record a start failure until the pipeline is available." : "Add cron_trigger to a pipeline’s YAML in Git, then refresh the list after Shogun indexes it."}</p>
              {!selectedName || selected ? <pre className="mt-2 rounded-md border bg-background/60 p-3 font-mono text-[11px]">{"spec:\n  triggers:\n    - type: cron_trigger"}</pre> : null}
            </div> : <p className="mt-2 text-[11px] text-muted-foreground">The pipeline must declare <code>cron_trigger</code> in Git.</p>}
            {selected && !selected.enabled ? <p className="mt-2 text-xs text-muted-foreground">This pipeline is disabled. Scheduled occurrences will record that the pipeline could not start.</p> : null}
          </div>
        </section>
        <section className="px-6 py-6">
          <label className="mb-2 block text-xs font-semibold" htmlFor="cron-schedule">Schedule</label>
          <Input id="cron-schedule" className="bg-muted/30 font-mono focus:bg-muted/40" autoComplete="off" {...form.register("schedule")} aria-invalid={!!form.formState.errors.schedule} />
          {form.formState.errors.schedule ? <FieldError message={form.formState.errors.schedule.message} /> : null}
          <p className="mt-2 font-mono text-[10px] text-muted-foreground">minute · hour · day of month · month · day of week</p>
          <div className="mt-3 flex flex-wrap gap-2">{presets.map(preset => <Button key={preset.label} type="button" size="sm" variant={schedule === preset.value ? "secondary" : "outline"} onClick={() => form.setValue("schedule", preset.value, { shouldDirty: true, shouldValidate: true })}>{preset.label}</Button>)}</div>
          <p className="mt-3 text-[11px] text-muted-foreground">All times use Shogun’s timezone, <span className="font-mono text-foreground">{timezone}</span>.</p>
        </section>
        <section className="px-6 py-6">
          <div className="flex items-center justify-between gap-3"><div className="flex items-center gap-2"><span className="size-1.5 rounded-full bg-primary" /><h3 className="text-xs font-semibold">Runtime values</h3></div><span className="rounded-full bg-muted px-2 py-0.5 font-mono text-[10px] text-muted-foreground">optional</span></div>
          <p className="mt-1 text-[11px] leading-5 text-muted-foreground">Names match <code className="text-foreground">{"{{VARIABLE}}"}</code> in your pipeline. Values can include secret references such as <code className="text-foreground">{"{{DEPLOY_TOKEN}}"}</code>.</p>
          <div className="mt-4 space-y-2.5">
            {values.fields.map((field, index) => <div className="value-row-enter grid gap-2 rounded-lg border bg-muted/15 p-3 transition-colors focus-within:border-ring/40 sm:grid-cols-[minmax(0,.9fr)_minmax(0,1.1fr)_auto]" key={field.id}>
              <label className="min-w-0"><span className="mb-1.5 block text-[10px] font-semibold tracking-[.1em] text-muted-foreground uppercase">Name</span><Input autoComplete="off" className="bg-muted/30 focus:bg-muted/40" placeholder="Variable name" {...form.register(`values.${index}.key`)} />{form.formState.errors.values?.[index]?.key ? <FieldError message={form.formState.errors.values[index].key.message} /> : null}</label>
              <label className="min-w-0"><span className="mb-1.5 block text-[10px] font-semibold tracking-[.1em] text-muted-foreground uppercase">Value</span><Input autoComplete="off" className="bg-muted/30 focus:bg-muted/40" placeholder="Value or {{SECRET_NAME}}" {...form.register(`values.${index}.value`)} /></label>
              <Button type="button" variant="ghost" size="icon-sm" className="self-end text-muted-foreground hover:text-destructive" onClick={() => values.remove(index)} aria-label={`Remove value ${index + 1}`}><Trash2 /></Button>
            </div>)}
            {!values.fields.length ? <p className="rounded-lg border border-dashed px-4 py-5 text-center text-xs text-muted-foreground">No runtime values. This cron will use an empty payload.</p> : null}
            <div className="flex justify-center pt-1"><Button type="button" variant="outline" size="sm" className="rounded-full border-dashed" disabled={!canAddValue} onClick={() => values.append({ key: "", value: "" }, { shouldFocus: true })}><Plus />Add runtime value</Button></div>
          </div>
        </section>
        <section className="px-6 py-5"><label className="flex cursor-pointer items-start gap-3"><input type="checkbox" className="mt-0.5 size-4 accent-primary" {...form.register("enabled")} /><span><span className="block text-xs font-semibold">Schedule enabled</span><span className="mt-1 block text-[11px] text-muted-foreground">Paused schedules keep their settings and run history.</span></span></label></section>
      </div>
      {mutation.error ? <p role="alert" className="mx-6 mb-5 rounded-lg border border-destructive/20 bg-destructive/10 p-3 text-xs text-destructive">{mutation.error.message}</p> : null}
      <div className="flex items-center justify-end gap-2 border-t bg-muted/20 px-6 py-4"><Button asChild variant="outline"><Link to={cron ? `/crons/${cron.id}` : "/crons"}>Cancel</Link></Button><Button type="submit" disabled={mutation.isPending || !canSavePipeline}><Save />{mutation.isPending ? "Saving…" : cron ? "Save changes" : "Create cron"}</Button></div>
    </div>
    <aside className="space-y-4 xl:sticky xl:top-[72px]">
      <div className="rounded-xl border bg-card p-5"><div className="flex items-center gap-2"><Clock3 className="size-4 text-primary" /><h2 className="text-xs font-semibold">Schedule context</h2></div><dl className="mt-5 space-y-4 text-xs"><div><dt className="text-muted-foreground">Shogun timezone</dt><dd className="mt-1 font-mono">{timezone}</dd></div><div><dt className="text-muted-foreground">Expression</dt><dd className="mt-1 break-words font-mono">{schedule || "—"}</dd></div>{cron?.next_run ? <div><dt className="text-muted-foreground">Next saved occurrence</dt><dd className="mt-1">{cronDate(cron.next_run, timezone)}</dd></div> : null}</dl><p className="mt-4 border-t pt-4 text-[11px] leading-5 text-muted-foreground">Each occurrence gets one start attempt. Runs may overlap. Shogun resumes with future occurrences after downtime.</p></div>
      <div className="rounded-xl border bg-muted/15 p-5"><h2 className="text-xs font-semibold">Secrets in values</h2><p className="mt-2 text-[11px] leading-5 text-muted-foreground">Use <code className="text-foreground">{"{{SECRET_NAME}}"}</code> to reference a stored secret. Its current value is resolved when the run starts.</p><Link to="/secrets" className="mt-3 inline-flex items-center gap-1 text-xs font-medium text-primary">Manage secrets<ArrowUpRight className="size-3" /></Link></div>
    </aside>
  </form>
}

function FieldError({ message }: { message?: string }) {
  return <span role="alert" className="mt-1.5 block text-xs text-destructive">{message}</span>
}
