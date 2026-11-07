import { beforeEach, describe, expect, it, vi } from "vitest"
import { render, screen, fireEvent, waitFor, within } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { MemoryRouter, Route, Routes } from "react-router-dom"

import { CronsPage } from "./crons-page"
import { CronPage } from "./cron-page"

const mocks = vi.hoisted(() => ({ role: "user", request: vi.fn(), pipelines: vi.fn() }))
vi.mock("@/features/auth/auth-provider", () => ({ useAuth: () => ({ claims: { role: mocks.role } }) }))
vi.mock("@/lib/api/client", () => ({ apiRequest: mocks.request }))
vi.mock("@/lib/api/endpoints", () => ({ api: { pipelines: mocks.pipelines } }))

function mount(path = "/crons") {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })
  const view = render(<QueryClientProvider client={client}><MemoryRouter initialEntries={[path]}><Routes>
    <Route path="/crons" element={<CronsPage />} />
    <Route path="/crons/new" element={<CronPage />} />
    <Route path="/crons/:cronId" element={<CronPage />} />
  </Routes></MemoryRouter></QueryClientProvider>)
  return { ...view, client }
}

const pipeline = { name: "deploy", enabled: true, triggers: [{ type: "cron_trigger" }], steps: [] }
beforeEach(() => {
  mocks.role = "user"
  mocks.request.mockReset()
  mocks.pipelines.mockReset().mockResolvedValue([pipeline])
  mocks.request.mockImplementation(async (path: string, options?: { method?: string; body?: unknown }) => {
    if (options?.method === "POST" || options?.method === "PATCH") return { id: 7, ...options.body as object }
    if (path === "/crons") return { timezone: "Asia/Kolkata", crons: [{ id: 7, name: "Nightly", pipeline: "deploy", schedule: "0 2 * * *", payload: { TOKEN: "{{DEPLOY_TOKEN}}" }, enabled: true, next_run: "2026-10-29T20:30:00Z" }] }
    return { has_more: false, runs: [{ id: 8, cron_id: 7, pipeline: "deploy", scheduled_at: "2026-10-28T20:30:00Z", status: "failed", error: "pipeline is disabled" }, { id: 9, cron_id: 7, pipeline: "deploy", scheduled_at: "2026-10-29T20:30:00Z", status: "started", pipeline_run_id: 42, pipeline_run: { status: "succeeded" } }] }
  })
  HTMLElement.prototype.scrollIntoView = vi.fn()
  HTMLElement.prototype.hasPointerCapture = vi.fn(() => false)
  HTMLElement.prototype.setPointerCapture = vi.fn()
  HTMLElement.prototype.releasePointerCapture = vi.fn()
})

describe("Cron dashboard pages", () => {
  it("opens a dedicated cron page from the list and shows persisted run results", async () => {
    mount()
    fireEvent.click(await screen.findByRole("link", { name: "Nightly" }))
    expect(await screen.findByText("pipeline is disabled")).toBeInTheDocument()
    expect(screen.getByRole("link", { name: "View pipeline run" })).toHaveAttribute("href", "/pipelines/deploy/runs/42?cron=7")
    expect(screen.queryByRole("link", { name: "Settings" })).not.toBeInTheDocument()
    expect(screen.queryByRole("combobox")).not.toBeInTheDocument()
    expect(screen.queryByText("{{DEPLOY_TOKEN}}")).not.toBeInTheDocument()
  })

  it("preserves old history links by redirecting them to the dedicated page", async () => {
    mount("/crons?cron=7")
    expect(await screen.findByRole("heading", { name: "Nightly" })).toBeInTheDocument()
    expect(await screen.findByText("pipeline is disabled")).toBeInTheDocument()
  })

  it("allows admins to pause a schedule", async () => {
    mocks.role = "admin"
    mount()
    fireEvent.click(await screen.findByRole("button", { name: "Pause" }))
    await waitFor(() => expect(mocks.request).toHaveBeenCalledWith("/admin/crons/7", expect.objectContaining({ method: "PATCH", body: expect.objectContaining({ enabled: false, payload: { TOKEN: "{{DEPLOY_TOKEN}}" } }) })))
  })

  it("opens admin settings as an inline form with editable name/value rows", async () => {
    mocks.role = "admin"
    mount("/crons/7")
    fireEvent.click(await screen.findByRole("link", { name: "Settings" }))
    expect(await screen.findByDisplayValue("{{DEPLOY_TOKEN}}")).toBeInTheDocument()
    expect(screen.getByLabelText("Cron name")).toHaveValue("Nightly")
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument()
    fireEvent.change(screen.getByLabelText("Cron name"), { target: { value: "Updated nightly" } })
    fireEvent.click(screen.getByRole("button", { name: "Save changes" }))
    await waitFor(() => expect(mocks.request).toHaveBeenCalledWith("/admin/crons/7", expect.objectContaining({ body: expect.objectContaining({ name: "Updated nightly", payload: { TOKEN: "{{DEPLOY_TOKEN}}" } }) })))
  })

  it("shows all pipelines and explains why a pipeline needs cron_trigger", async () => {
    mocks.role = "admin"
    mocks.pipelines.mockResolvedValue([{ ...pipeline, name: "webhook-only", triggers: [{ type: "ci_webhook" }] }])
    mount("/crons/new")
    expect(await screen.findByText("No pipelines declare cron_trigger yet")).toBeInTheDocument()
    const user = userEvent.setup()
    await user.click(screen.getByRole("combobox", { name: "Pipeline" }))
    await user.click(screen.getByRole("option", { name: /webhook-only/ }))
    expect(screen.getByText("webhook-only needs cron_trigger")).toBeInTheDocument()
    expect(screen.getByRole("button", { name: "Create cron" })).toBeDisabled()
  })

  it("creates a cron using runtime rows and keeps secret references intact", async () => {
    mocks.role = "admin"
    mount("/crons/new")
    const user = userEvent.setup()
    await screen.findByRole("heading", { name: "Create cron" })
    await waitFor(() => expect(screen.getByRole("combobox", { name: "Pipeline" })).not.toBeDisabled())
    await user.type(screen.getByLabelText("Cron name"), "Cleanup")
    await user.click(screen.getByRole("combobox", { name: "Pipeline" }))
    await user.click(screen.getByRole("option", { name: /deploy/ }))
    await user.click(screen.getByRole("button", { name: "Add runtime value" }))
    fireEvent.change(screen.getByPlaceholderText("Variable name"), { target: { value: "TOKEN" } })
    fireEvent.change(screen.getByPlaceholderText("Value or {{SECRET_NAME}}"), { target: { value: "{{DEPLOY_TOKEN}}" } })
    await user.click(screen.getByRole("button", { name: "Create cron" }))
    await waitFor(() => expect(mocks.request).toHaveBeenCalledWith("/admin/crons", { method: "POST", body: { name: "Cleanup", pipeline: "deploy", schedule: "0 2 * * *", payload: { TOKEN: "{{DEPLOY_TOKEN}}" }, enabled: true } }))
  })

  it("allows at most five unnamed rows and frees a slot when a row is named or removed", async () => {
    mocks.role = "admin"
    mount("/crons/new")
    await screen.findByRole("heading", { name: "Create cron" })
    const add = screen.getByRole("button", { name: "Add runtime value" })
    for (let i = 0; i < 5; i++) {
      expect(add).toBeEnabled()
      fireEvent.click(add)
    }
    expect(screen.getAllByPlaceholderText("Variable name")).toHaveLength(5)
    expect(add).toBeDisabled()
    fireEvent.change(screen.getAllByPlaceholderText("Variable name")[0], { target: { value: "   " } })
    expect(add).toBeDisabled()
    fireEvent.change(screen.getAllByPlaceholderText("Variable name")[0], { target: { value: "TOKEN" } })
    expect(add).toBeEnabled()
    fireEvent.click(add)
    expect(screen.getAllByPlaceholderText("Variable name")).toHaveLength(6)
    expect(add).toBeDisabled()
    fireEvent.click(screen.getByRole("button", { name: "Remove value 6" }))
    expect(add).toBeEnabled()
  })

  it("omits unnamed rows from the request while keeping named rows with empty values", async () => {
    mocks.role = "admin"
    mount("/crons/7?tab=settings")
    await screen.findByDisplayValue("{{DEPLOY_TOKEN}}")
    const add = screen.getByRole("button", { name: "Add runtime value" })
    for (let i = 0; i < 3; i++) fireEvent.click(add)
    fireEvent.change(screen.getAllByPlaceholderText("Variable name")[1], { target: { value: "   " } })
    fireEvent.change(screen.getAllByPlaceholderText("Value or {{SECRET_NAME}}")[1], { target: { value: "ignored without a name" } })
    fireEvent.change(screen.getAllByPlaceholderText("Variable name")[3], { target: { value: "EMPTY_VALUE" } })
    fireEvent.click(screen.getByRole("button", { name: "Save changes" }))
    await waitFor(() => expect(mocks.request).toHaveBeenCalledWith("/admin/crons/7", expect.objectContaining({ method: "PATCH", body: expect.objectContaining({ payload: { TOKEN: "{{DEPLOY_TOKEN}}", EMPTY_VALUE: "" } }) })))
  })

  it("saves an empty payload when every runtime row is unnamed", async () => {
    mocks.role = "admin"
    mount("/crons/7?tab=settings")
    await screen.findByDisplayValue("{{DEPLOY_TOKEN}}")
    fireEvent.change(screen.getByPlaceholderText("Variable name"), { target: { value: "" } })
    fireEvent.click(screen.getByRole("button", { name: "Add runtime value" }))
    fireEvent.click(screen.getByRole("button", { name: "Save changes" }))
    await waitFor(() => expect(mocks.request).toHaveBeenCalledWith("/admin/crons/7", expect.objectContaining({ body: expect.objectContaining({ payload: {} }) })))
  })

  it("rejects duplicate payload keys instead of silently overwriting a value", async () => {
    mocks.role = "admin"
    mount("/crons/7?tab=settings")
    await screen.findByDisplayValue("{{DEPLOY_TOKEN}}")
    fireEvent.click(screen.getByRole("button", { name: "Add runtime value" }))
    fireEvent.change(screen.getAllByPlaceholderText("Variable name")[1], { target: { value: "TOKEN" } })
    fireEvent.click(screen.getByRole("button", { name: "Save changes" }))
    expect(await screen.findByText("Each value name must be unique.")).toBeInTheDocument()
    expect(mocks.request.mock.calls.some(([, options]) => options?.method === "PATCH")).toBe(false)
  })

  it("shows an error and retry control when loading pipeline choices fails", async () => {
    mocks.role = "admin"
    mocks.pipelines.mockRejectedValue(new Error("unreachable"))
    mount("/crons/new")
    expect(await screen.findByText("Could not load pipelines.")).toBeInTheDocument()
    expect(screen.getByRole("button", { name: "Try again" })).toBeInTheDocument()
    expect(screen.getByRole("button", { name: "Create cron" })).toBeDisabled()
  })

  it("paginates ten newest runs with older, newer, and latest controls", async () => {
    const original = mocks.request.getMockImplementation()!
    const records = Array.from({ length: 25 }, (_, index) => ({ id: 25 - index, cron_id: 7, pipeline: "deploy", scheduled_at: "2026-10-29T20:30:00Z", status: "failed", error: `Occurrence ${25 - index}` }))
    mocks.request.mockImplementation(async (path: string, options?: unknown) => {
      if (!path.startsWith("/crons/7/runs")) return original(path, options)
      const before = Number(new URL(path, "http://test").searchParams.get("before"))
      const remaining = records.filter(run => !before || run.id < before)
      return { runs: remaining.slice(0, 10), has_more: remaining.length > 10 }
    })
    mount("/crons/7")
    expect(await screen.findByText("Occurrence 25")).toBeInTheDocument()
    expect(screen.getAllByRole("row")).toHaveLength(11)
    expect(screen.getByRole("button", { name: "Newer" })).toBeDisabled()
    fireEvent.click(screen.getByRole("button", { name: "Older" }))
    expect(await screen.findByText("Occurrence 15")).toBeInTheDocument()
    expect(screen.queryByText("Occurrence 25")).not.toBeInTheDocument()
    fireEvent.click(screen.getByRole("button", { name: "Older" }))
    expect(await screen.findByText("Occurrence 5")).toBeInTheDocument()
    expect(screen.getAllByRole("row")).toHaveLength(6)
    expect(screen.getByRole("button", { name: "Older" })).toBeDisabled()
    fireEvent.click(screen.getByRole("button", { name: "Newer" }))
    expect(await screen.findByText("Occurrence 15")).toBeInTheDocument()
    fireEvent.click(screen.getByRole("button", { name: "Newer" }))
    expect(await screen.findByText("Occurrence 25")).toBeInTheDocument()
    fireEvent.click(screen.getByRole("button", { name: "Older" }))
    expect(await screen.findByText("Occurrence 15")).toBeInTheDocument()
    fireEvent.click(screen.getByRole("button", { name: "Latest runs" }))
    expect(await screen.findByText("Occurrence 25")).toBeInTheDocument()
    expect(screen.getByRole("button", { name: "Newer" })).toBeDisabled()
  })

  it("shows the next scheduled date for paused crons and marks it as a preview", async () => {
    const original = mocks.request.getMockImplementation()!
    mocks.request.mockImplementation(async (path: string, options?: unknown) => {
      const data = await original(path, options)
      return path === "/crons" ? { ...data, crons: data.crons.map((cron: { enabled: boolean }) => ({ ...cron, enabled: false })) } : data
    })
    mount("/crons/7")
    expect(await screen.findByText("Paused · will not execute")).toBeInTheDocument()
    const summary = screen.getByText("Next scheduled time").parentElement!
    expect(summary).toHaveTextContent("2026")
    expect(summary).not.toHaveTextContent("Not available")
  })

  it.each(["admin", "user"])("shows deleted cron metadata to %s with admin-only saved values", async (role) => {
    mocks.role = role
    const original = mocks.request.getMockImplementation()!
    mocks.request.mockImplementation(async (path: string, options?: unknown) => {
      if (path === "/crons") return { timezone: "Asia/Kolkata", crons: [] }
      if (path === "/crons/7") return { id: 7, name: "Archived nightly", pipeline: "deploy", schedule: "0 2 * * *", enabled: false, deleted_at: "2026-10-29T06:00:00Z", ...(role === "admin" ? { payload: { TOKEN: "{{DEPLOY_TOKEN}}" } } : {}) }
      return original(path, options)
    })
    mount("/crons/7")
    expect(await screen.findByRole("heading", { name: "Archived nightly" })).toBeInTheDocument()
    expect(screen.getByText("Deleted · no future runs")).toBeInTheDocument()
    expect(screen.getByText("0 2 * * *")).toBeInTheDocument()
    expect(await screen.findByText("pipeline is disabled")).toBeInTheDocument()
    expect(screen.queryByRole("link", { name: "Settings" })).not.toBeInTheDocument()
    expect(screen.queryByRole("button", { name: "Resume" })).not.toBeInTheDocument()
    if (role === "admin") {
      expect(screen.getByText("{{DEPLOY_TOKEN}}")).toBeInTheDocument()
      expect(screen.getByRole("button", { name: "Permanently delete" })).toBeInTheDocument()
    } else {
      expect(screen.queryByText("{{DEPLOY_TOKEN}}")).not.toBeInTheDocument()
      expect(screen.queryByRole("button", { name: "Permanently delete" })).not.toBeInTheDocument()
    }
  })

  it.each(["admin", "user"])("shows deleted crons to %s and restricts permanent deletion", async role => {
    mocks.role = role
    const original = mocks.request.getMockImplementation()!
    mocks.request.mockImplementation(async (path: string, options?: unknown) => {
      if (path === "/crons/deleted") return [{ id: 9, name: "Old cron", pipeline: "deploy", schedule: "0 2 * * *", enabled: false, deleted_at: "2026-10-29T06:00:00Z" }]
      return original(path, options)
    })
    mount("/crons")
    fireEvent.click(screen.getByRole("link", { name: "Deleted crons" }))
    expect(await screen.findByRole("link", { name: "Old cron" })).toHaveAttribute("href", "/crons/9")
    expect(screen.queryByRole("button", { name: "Permanently delete" })).not.toBeInTheDocument()
  })

  it("requires confirmation on the deleted cron page and returns to the list", async () => {
    mocks.role = "admin"
    const original = mocks.request.getMockImplementation()!
    const oldCron = { id: 9, name: "Old cron", pipeline: "deploy", schedule: "0 2 * * *", enabled: false, deleted_at: "2026-10-29T06:00:00Z" }
    let removed = false
    mocks.request.mockImplementation(async (path: string, options?: { method?: string }) => {
      if (path === "/crons/deleted") return removed ? [] : [oldCron]
      if (path === "/crons/9") return oldCron
      if (path === "/admin/crons/9/permanent") { removed = true; return undefined }
      return original(path, options)
    })
    mount("/crons?view=deleted")
    fireEvent.click(await screen.findByRole("link", { name: "Old cron" }))
    fireEvent.click(await screen.findByRole("button", { name: "Permanently delete" }))
    const dialog = screen.getByRole("alertdialog")
    expect(within(dialog).getByText(/permanently deletes the cron and all its cron run records/i)).toBeInTheDocument()
    expect(mocks.request).not.toHaveBeenCalledWith("/admin/crons/9/permanent", expect.anything())
    fireEvent.click(within(dialog).getByRole("button", { name: "Permanently delete" }))
    await waitFor(() => expect(mocks.request).toHaveBeenCalledWith("/admin/crons/9/permanent", { method: "DELETE" }))
    expect(await screen.findByText("No deleted crons")).toBeInTheDocument()
  })

  it("keeps non-admin users on history when they open a settings URL", async () => {
    mount("/crons/7?tab=settings")
    expect(await screen.findByText("pipeline is disabled")).toBeInTheDocument()
    expect(screen.queryByLabelText("Cron name")).not.toBeInTheDocument()
    expect(screen.queryByRole("button", { name: "Save changes" })).not.toBeInTheDocument()
  })
})
