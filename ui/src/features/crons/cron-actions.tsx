import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { Pause, Play, Trash2 } from "lucide-react"
import { toast } from "sonner"

import { AlertDialog, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { cronApi, type CronDefinition } from "./api"

export function CronActions({ cron, onDeleted }: { cron: CronDefinition; onDeleted?: () => void }) {
  const client = useQueryClient()
  const [deleting, setDeleting] = useState(false)
  const change = useMutation({
    mutationFn: () => cronApi.save(cron.id, { name: cron.name, pipeline: cron.pipeline, schedule: cron.schedule, payload: cron.payload ?? {}, enabled: !cron.enabled }),
    onSuccess: () => client.invalidateQueries({ queryKey: ["crons"] }),
    onError: (error) => toast.error(error.message),
  })
  const remove = useMutation({
    mutationFn: () => cronApi.remove(cron.id),
    onSuccess: async () => {
      await Promise.all([client.invalidateQueries({ queryKey: ["crons"] }), client.invalidateQueries({ queryKey: ["deleted-crons"] })])
      setDeleting(false)
      toast.success("Cron deleted")
      onDeleted?.()
    },
    onError: (error) => toast.error(error.message),
  })

  return <>
    <Button size="sm" variant="outline" disabled={change.isPending} onClick={() => change.mutate()}>{cron.enabled ? <Pause /> : <Play />}{cron.enabled ? "Pause" : "Resume"}</Button>
    <Button size="icon-sm" variant="ghost" aria-label={`Delete ${cron.name}`} onClick={() => setDeleting(true)}><Trash2 /></Button>
    <AlertDialog open={deleting} onOpenChange={setDeleting}>
      <AlertDialogContent>
        <AlertDialogHeader><AlertDialogTitle>Delete {cron.name}?</AlertDialogTitle><AlertDialogDescription>Future executions will stop. Recorded run history is retained.</AlertDialogDescription></AlertDialogHeader>
        <AlertDialogFooter><AlertDialogCancel>Cancel</AlertDialogCancel><Button variant="destructive" disabled={remove.isPending} onClick={() => remove.mutate()}>{remove.isPending ? "Deleting…" : "Delete cron"}</Button></AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </>
}

export function PermanentlyDeleteCron({ cron, onDeleted }: { cron: CronDefinition; onDeleted: () => void }) {
  const client = useQueryClient()
  const [open, setOpen] = useState(false)
  const remove = useMutation({
    mutationFn: () => cronApi.permanentlyDelete(cron.id),
    onSuccess: async () => {
      onDeleted()
      client.removeQueries({ queryKey: ["cron-detail", cron.id] })
      client.removeQueries({ queryKey: ["cron-runs", cron.id] })
      await client.invalidateQueries({ queryKey: ["deleted-crons"] })
      setOpen(false)
      toast.success("Cron and its run records permanently deleted")
    },
    onError: (error) => toast.error(error.message),
  })
  return <>
    <Button variant="destructive" size="sm" onClick={() => setOpen(true)}><Trash2 />Permanently delete</Button>
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogContent><AlertDialogHeader><AlertDialogTitle>Permanently delete {cron.name}?</AlertDialogTitle><AlertDialogDescription>This permanently deletes the cron and all its cron run records. Linked pipeline runs remain in pipeline history. This cannot be undone.</AlertDialogDescription></AlertDialogHeader><AlertDialogFooter><AlertDialogCancel disabled={remove.isPending}>Cancel</AlertDialogCancel><Button variant="destructive" disabled={remove.isPending} onClick={() => remove.mutate()}>{remove.isPending ? "Deleting…" : "Permanently delete"}</Button></AlertDialogFooter></AlertDialogContent>
    </AlertDialog>
  </>
}
