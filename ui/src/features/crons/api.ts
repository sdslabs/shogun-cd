import { useQuery } from "@tanstack/react-query"

import { apiRequest } from "@/lib/api/client"
import type { PipelineRun } from "@/lib/api/types"

export interface CronDefinition {
  id: number
  name: string
  pipeline: string
  schedule: string
  payload?: Record<string, string>
  enabled: boolean
  next_run?: string
  deleted_at?: string
}

export interface CronRun {
  id: number
  cron_id: number
  pipeline: string
  scheduled_at: string
  status: string
  error?: string
  pipeline_run_id?: number
  pipeline_run?: PipelineRun
}

export interface CronList {
  timezone: string
  crons: CronDefinition[]
}

export interface CronInput {
  name: string
  pipeline: string
  schedule: string
  payload: Record<string, string>
  enabled: boolean
}

export const cronApi = {
  list: () => apiRequest<CronList>("/crons"),
  deleted: () => apiRequest<CronDefinition[]>("/crons/deleted"),
  permanentlyDelete: (id: number) => apiRequest<undefined>(`/admin/crons/${id}/permanent`, { method: "DELETE" }),
  get: (id: number) => apiRequest<CronDefinition>(`/crons/${id}`),
  save: (id: number | undefined, body: CronInput) =>
    apiRequest<CronDefinition>(id ? `/admin/crons/${id}` : "/admin/crons", { method: id ? "PATCH" : "POST", body }),
  remove: (id: number) => apiRequest<undefined>(`/admin/crons/${id}`, { method: "DELETE" }),
  runs: (id: number, before: number) =>
    apiRequest<{ runs: CronRun[]; has_more: boolean }>(`/crons/${id}/runs${before ? `?before=${before}` : ""}`),
}

export function useCrons() {
  return useQuery({ queryKey: ["crons"], queryFn: cronApi.list, refetchInterval: 10_000 })
}

export function cronDate(value: string, timezone: string) {
  return new Intl.DateTimeFormat(undefined, { timeZone: timezone, dateStyle: "medium", timeStyle: "short" }).format(new Date(value))
}
