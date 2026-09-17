import { useQuery } from "@tanstack/react-query"

import { api } from "@/lib/api/endpoints"
import { queryKeys } from "@/lib/query"

export function useTargets() {
  return useQuery({ queryKey: queryKeys.targets, queryFn: api.targets })
}
