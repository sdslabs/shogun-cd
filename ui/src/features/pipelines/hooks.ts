import { useQuery } from "@tanstack/react-query"

import { api } from "@/lib/api/endpoints"
import { queryKeys } from "@/lib/query"

export function usePipelines() {
  return useQuery({ queryKey: queryKeys.pipelines, queryFn: api.pipelines })
}

export function usePipeline(pipelineName?: string) {
  const query = usePipelines()
  const pipeline = query.data?.find((item) => item.name.toLowerCase() === pipelineName?.toLowerCase())
  return { ...query, pipeline }
}

export function usePipelineRuns(pipelineName?: string) {
  return useQuery({
    queryKey: queryKeys.runs(pipelineName ?? ""),
    queryFn: () => api.pipelineRuns(pipelineName!),
    enabled: Boolean(pipelineName),
    refetchInterval: (query) => query.state.data?.some((run) => run.status === "running" || run.status === "queued") ? 2_500 : false,
  })
}

export function usePipelineRun(pipelineName?: string, runId?: number) {
  return useQuery({
    queryKey: queryKeys.run(pipelineName ?? "", runId ?? 0),
    queryFn: () => api.pipelineRun(pipelineName!, runId!),
    enabled: Boolean(pipelineName && runId),
    refetchInterval: (query) => {
      const run = query.state.data
      return run?.status === "running" || run?.status === "queued" ? 2_000 : false
    },
  })
}
