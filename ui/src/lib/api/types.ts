export interface ApiResponse<T> {
  success: boolean
  message: string
  data?: T
  error?: string
}

export type UserRole = "admin" | "user"

export interface LoginData {
  token: string
  expiration_hours: number
}

export interface LoginInput {
  email: string
  password: string
}

export type PipelineRunStatus = "queued" | "running" | "succeeded" | "failed" | "cancelled"
export type StepStatus = "in_progress" | "succeeded" | "failed" | "skipped"

export interface PipelineTrigger {
  type: "ci_webhook" | "git_changes" | string
  paths?: string[]
}

export interface PipelineStepSummary {
  index: number
  type: "mutate" | "sync" | "exec" | "apply" | string
  trigger_when?: string
  target?: string
  config: PipelineStepConfig
}

export interface PipelineStepConfig {
  trigger_when?: string
  target?: string
  files?: Array<string | { src: string; dst: string }>
  commands?: string[]
  changes?: Array<{ file: string; update_field: string; value: string }>
  [field: string]: unknown
}

export interface PipelineRunSummary {
  id: number
  status: PipelineRunStatus
  trigger_kind: string
  success: boolean | null
  started_at: string
  finished_at?: string | null
}

export interface PipelineSummary {
  name: string
  enabled: boolean
  triggers: PipelineTrigger[]
  steps: PipelineStepSummary[]
  last_run?: PipelineRunSummary | null
}

export interface PipelineRunStep {
  step_index: number
  step_type: string
  status: StepStatus
  logs: string
  started_at: string
  finished_at?: string | null
}

export interface PipelineRun extends PipelineRunSummary {
  pipeline: string
  steps: PipelineRunStep[]
}

export interface TargetSummary {
  name: string
  type: "server" | "cluster" | string
  host: string
  user: string
  port: number
  access_secret: string
}

export interface Webhook {
  slug: string
  pipeline_name: string
  alias: string
  created_by: string
  created_at: string
  is_active: boolean
}

export interface CreateWebhookInput {
  pipeline: string
  alias: string
}

export interface CreatedWebhook {
  webhook: Webhook
  secret: string
}

export interface SecretSummary {
  name: string
  created_at?: string
  updated_at?: string
}

export interface SecretInput {
  name: string
  value: string
}
