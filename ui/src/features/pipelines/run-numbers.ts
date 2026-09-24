import type { PipelineRunSummary } from "@/lib/api/types"

// The API uses global IDs. Keep them for URLs and lookups, but number runs
// within each pipeline for display.
export function runNumbersById(runs: ReadonlyArray<Pick<PipelineRunSummary, "id">> = []): Map<number, number> {
  return new Map(
    [...runs]
      .sort((a, b) => a.id - b.id)
      .map((run, index) => [run.id, index + 1]),
  )
}
