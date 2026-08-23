import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { describe, expect, it, vi } from "vitest"

import { StepChain } from "@/components/step-chain"

const steps = [
  { index: 0, type: "sync", target: "prod-api" },
  { index: 1, type: "exec", target: "{{TARGET}}" },
]

describe("StepChain", () => {
  it("keeps runtime targets intentionally generic", () => {
    render(<StepChain steps={steps} />)
    expect(screen.getByText("prod-api")).toBeInTheDocument()
    expect(screen.getByText("runtime target")).toBeInTheDocument()
    expect(screen.queryByText("All logs")).not.toBeInTheDocument()
  })

  it("selects a step when used as a log control", async () => {
    const onSelect = vi.fn()
    render(<StepChain steps={steps} selectedStep={null} onSelect={onSelect} />)
    await userEvent.click(screen.getByRole("button", { name: /2\. Exec/ }))
    expect(onSelect).toHaveBeenCalledWith(1)
  })
})
