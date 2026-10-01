import { QueryClient } from "@tanstack/react-query"

import { ApiError } from "@/lib/api/client"

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 15_000,
      retry: (failureCount, error) => {
        const status = error instanceof ApiError ? error.status : 0
        return status !== 401 && status !== 403 && failureCount < 2
      },
      refetchOnWindowFocus: true,
    },
    mutations: { retry: false },
  },
})

export const queryKeys = {
  pipelines: ["pipelines"] as const,
  targets: ["targets"] as const,
  runs: (pipeline: string) => ["pipelines", pipeline, "runs"] as const,
  run: (pipeline: string, runId: number) => ["pipelines", pipeline, "runs", runId] as const,
  webhooks: (pipeline?: string) => ["webhooks", pipeline ?? "all"] as const,
  secrets: ["secrets"] as const,
}
