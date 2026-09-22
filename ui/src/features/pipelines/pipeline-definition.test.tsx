import { useState } from "react"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { describe, expect, it } from "vitest"

import { PipelineDefinition } from "@/features/pipelines/pipeline-definition"
import type { PipelineStepSummary } from "@/lib/api/types"

const steps: PipelineStepSummary[] = [
  {
    index: 0,
    type: "exec",
    target: "production",
    config: { target: "production", commands: ["systemctl restart shogun"] },
  },
]

function DefinitionHarness() {
  const [open, setOpen] = useState(false)
  return <PipelineDefinition steps={steps} open={open} selectedStep={0} onOpenChange={setOpen} />
}

describe("PipelineDefinition", () => {
  it("opens and closes the selected step configuration", async () => {
    const user = userEvent.setup()
    render(<DefinitionHarness />)

    const trigger = screen.getByRole("button", { name: /pipeline configuration/i })
    expect(screen.queryByText("systemctl restart shogun")).not.toBeInTheDocument()

    await user.click(trigger)
    expect(screen.getByText("systemctl restart shogun")).toBeVisible()

    await user.click(trigger)
    expect(screen.queryByText("systemctl restart shogun")).not.toBeInTheDocument()
  })
})
