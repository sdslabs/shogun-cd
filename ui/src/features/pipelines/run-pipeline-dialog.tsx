import { zodResolver } from "@hookform/resolvers/zod"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { Play, Plus, Trash2 } from "lucide-react"
import { useFieldArray, useForm } from "react-hook-form"
import { toast } from "sonner"
import { z } from "zod"

import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { api } from "@/lib/api/endpoints"
import { queryKeys } from "@/lib/query"

const valueKeyPattern = /^[A-Za-z_][A-Za-z0-9_]*$/

const triggerValueSchema = z.object({
  key: z.string().regex(valueKeyPattern, "Start with a letter or underscore; use only letters, numbers, and underscores."),
  value: z.string(),
})

const triggerValuesFormSchema = z.object({
  values: z.array(triggerValueSchema),
}).superRefine(({ values }, ctx) => {
  const seen = new Set<string>()
  values.forEach(({ key }, index) => {
    if (!key || seen.has(key)) {
      if (key) {
        ctx.addIssue({
          code: "custom",
          path: ["values", index, "key"],
          message: "Each value name must be unique.",
        })
      }
      return
    }
    seen.add(key)
  })
})

type TriggerValuesForm = z.infer<typeof triggerValuesFormSchema>

interface RunPipelineDialogProps {
  pipelineName: string
  open: boolean
  onOpenChange: (open: boolean) => void
  onStarted: (runID: number) => void
}

export function RunPipelineDialog({ pipelineName, open, onOpenChange, onStarted }: RunPipelineDialogProps) {
  const queryClient = useQueryClient()
  const form = useForm<TriggerValuesForm>({
    resolver: zodResolver(triggerValuesFormSchema),
    defaultValues: { values: [] },
  })
  const values = useFieldArray({ control: form.control, name: "values" })

  const mutation = useMutation({
    mutationFn: (input: Record<string, string>) => api.startPipelineRun(pipelineName, { values: input }),
    onSuccess: async ({ run_id }) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.pipelines }),
        queryClient.invalidateQueries({ queryKey: queryKeys.runs(pipelineName) }),
      ])
      form.reset({ values: [] })
      onOpenChange(false)
      toast.success(`Run #${run_id} started`)
      onStarted(run_id)
    },
  })

  const close = (nextOpen: boolean) => {
    if (!nextOpen) {
      form.reset({ values: [] })
      mutation.reset()
    }
    onOpenChange(nextOpen)
  }

  const submit = ({ values: entries }: TriggerValuesForm) => {
    const triggerValues = Object.fromEntries(entries.map(({ key, value }) => [key, value]))
    mutation.mutate(triggerValues)
  }

  const addValue = () => values.append({ key: "", value: "" }, { shouldFocus: true })

  return (
    <Dialog open={open} onOpenChange={close}>
      <DialogContent className="flex max-h-[calc(100dvh-2rem)] max-w-xl flex-col gap-0 overflow-hidden border-primary/15 bg-popover p-0">
        <form className="flex min-h-0 flex-1 flex-col" onSubmit={form.handleSubmit(submit)}>
          <DialogHeader className="shrink-0 border-b border-primary/10 bg-gradient-to-br from-primary/12 via-primary/[.035] to-transparent px-6 pt-6 pb-5">
            <div className="flex items-start gap-3">
              <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-sm">
                <Play className="size-4" fill="currentColor" />
              </span>
              <div>
                <p className="text-[10px] font-semibold tracking-[.16em] text-muted-foreground uppercase">Manual execution</p>
                <DialogTitle className="mt-1 text-lg tracking-[-0.02em]">Run pipeline</DialogTitle>
              </div>
            </div>
            <DialogDescription className="mt-3 max-w-md text-xs leading-5">
              Run <code className="rounded bg-background/70 px-1.5 py-0.5 font-mono text-[11px] text-foreground">{pipelineName}</code> with the values below. They are available to this execution as pipeline runtime values.
            </DialogDescription>
          </DialogHeader>

          <div className="min-h-0 flex-1 overflow-y-auto px-6 py-5">
            <div className="space-y-5">
              <section>
                <div className="flex items-center justify-between gap-3">
                  <div className="flex items-center gap-2">
                    <span className="size-1.5 rounded-full bg-primary" />
                    <p className="text-xs font-semibold">Runtime values</p>
                  </div>
                  <span className="rounded-full bg-muted px-2 py-0.5 font-mono text-[10px] text-muted-foreground">optional</span>
                </div>
                <p className="mt-1 text-[11px] leading-4 text-muted-foreground">Use names that match <code className="font-mono text-[10px] text-foreground/80">{"{{Variables}}"}</code> in your pipeline.</p>

                <div className="mt-3 space-y-2.5">
                  {values.fields.length ? values.fields.map((field, index) => (
                    <div className="value-row-enter grid gap-2 rounded-lg border bg-muted/40 p-3 shadow-sm transition-[background-color,border-color,box-shadow] hover:border-primary/25 hover:bg-muted/55 sm:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)_auto]" key={field.id}>
                      <label className="min-w-0">
                        <span className="mb-1.5 block text-[10px] font-semibold tracking-[.1em] text-muted-foreground uppercase">Name</span>
                        <Input autoComplete="off" className="bg-popover focus:bg-popover" placeholder="Variable name" {...form.register(`values.${index}.key`)} />
                        {form.formState.errors.values?.[index]?.key ? <span className="mt-1.5 block text-xs text-destructive">{form.formState.errors.values[index].key.message}</span> : null}
                      </label>
                      <label className="min-w-0">
                        <span className="mb-1.5 block text-[10px] font-semibold tracking-[.1em] text-muted-foreground uppercase">Value</span>
                        <Input autoComplete="off" className="bg-popover focus:bg-popover" placeholder="Variable value" {...form.register(`values.${index}.value`)} />
                      </label>
                      <Button type="button" variant="ghost" size="icon-sm" className="self-end text-muted-foreground hover:text-destructive" onClick={() => values.remove(index)} aria-label="Remove value">
                        <Trash2 />
                      </Button>
                    </div>
                  )) : <p className="px-2 py-4 text-center text-xs text-muted-foreground">No values added yet.</p>}

                  <div className="flex justify-center pt-0.5">
                    <Button type="button" variant="outline" size="icon-sm" className="rounded-full border-dashed bg-background transition-transform hover:rotate-90 hover:border-primary/50 hover:bg-accent" onClick={addValue} aria-label="Add runtime value">
                      <Plus />
                    </Button>
                  </div>
                </div>
              </section>

              {mutation.error ? <p className="rounded-lg border border-destructive/20 bg-destructive/10 p-3 text-xs text-destructive">{mutation.error.message}</p> : null}
            </div>
          </div>

          <DialogFooter className="shrink-0 border-t bg-muted/30 px-6 py-4">
            <Button type="button" variant="outline" onClick={() => close(false)}>Cancel</Button>
            <Button type="submit" disabled={mutation.isPending}>{mutation.isPending ? "Starting…" : "Run pipeline"}</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
