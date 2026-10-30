import { expect, it, vi } from "vitest"
import { render, screen, fireEvent } from "@testing-library/react"
import { MemoryRouter, Routes, Route } from "react-router-dom"
import { RunPage } from "./run-page"
vi.mock("@/features/pipelines/hooks", () => ({
 usePipeline: () => ({ pipeline: undefined, isLoading: false }),
 usePipelineRun: () => ({ data: { id: 42, status: "succeeded", trigger_kind: "cron_trigger", started_at: "2026-10-29T10:00:00Z", finished_at: "2026-10-29T10:01:00Z", steps: [] }, isLoading: false }),
 usePipelineRuns: () => ({ data: [], isLoading: false }),
}))
it("returns to the originating cron even if the pipeline was removed from Git", () => {
 render(<MemoryRouter initialEntries={["/pipelines/deploy/runs/42?cron=7"]}><Routes><Route path="/pipelines/:pipeline/runs/:runId" element={<RunPage />} /><Route path="/crons/:cronId" element={<p>Cron history</p>} /></Routes></MemoryRouter>)
 const back = screen.getByRole("link", { name: "Back to cron" })
 expect(back).toHaveAttribute("href", "/crons/7")
 fireEvent.click(back)
 expect(screen.getByText("Cron history")).toBeInTheDocument()
})
