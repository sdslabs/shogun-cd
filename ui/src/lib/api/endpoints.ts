import { ApiError, apiRequest } from "@/lib/api/client"
import type {
  CreateWebhookInput,
  CreatedWebhook,
  LoginData,
  LoginInput,
  PipelineRun,
  PipelineSummary,
  SecretInput,
  SecretSummary,
  TargetSummary,
  Webhook,
} from "@/lib/api/types"

export const api = {
  login: (input: LoginInput) =>
    apiRequest<LoginData>("/auth/login", { method: "POST", body: input, auth: false }),

  register: (input: LoginInput) =>
    apiRequest<undefined>("/auth/register", { method: "POST", body: input, auth: false }),

  pipelines: () => apiRequest<PipelineSummary[]>("/system/pipelines"),

  targets: () => apiRequest<TargetSummary[]>("/system/targets"),

  pipelineRuns: async (pipeline: string) => {
    try {
      return await apiRequest<PipelineRun[]>(`/pipelines/${encodeURIComponent(pipeline)}/runs`)
    } catch (error) {
      if (error instanceof ApiError && error.status === 404) return []
      throw error
    }
  },

  pipelineRun: async (pipeline: string, runId: number) => {
    const runs = await apiRequest<PipelineRun[]>(
      `/pipelines/${encodeURIComponent(pipeline)}/runs/${runId}`,
    )
    return runs[0]
  },

  webhooks: (pipeline?: string) =>
    apiRequest<Webhook[]>(pipeline ? `/hook/${encodeURIComponent(pipeline)}` : "/hook"),

  createWebhook: (input: CreateWebhookInput) =>
    apiRequest<CreatedWebhook>("/admin/hook", { method: "POST", body: input }),

  setWebhookStatus: (slug: string, active: boolean) =>
    apiRequest<undefined>(`/admin/hook/${encodeURIComponent(slug)}/${active ? "resume" : "pause"}`, {
      method: "PATCH",
    }),

  deleteWebhook: (slug: string) =>
    apiRequest<undefined>(`/admin/hook/${encodeURIComponent(slug)}`, { method: "DELETE" }),

  secrets: async () => {
    const data = await apiRequest<Array<SecretSummary & Record<string, unknown>>>("/admin/secrets")
    return data.map((secret) => ({
      name: String(secret.name ?? secret.Name ?? ""),
      created_at: String(secret.created_at ?? secret.CreatedAt ?? "") || undefined,
      updated_at: String(secret.updated_at ?? secret.UpdatedAt ?? "") || undefined,
    }))
  },

  setSecrets: (input: SecretInput[]) =>
    apiRequest<undefined>("/admin/secrets", { method: "POST", body: input }),

  deleteSecret: (name: string) =>
    apiRequest<undefined>(`/admin/secrets/${encodeURIComponent(name)}`, { method: "DELETE" }),
}
