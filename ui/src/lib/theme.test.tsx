import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { beforeEach, describe, expect, it } from "vitest"

import { ThemeProvider, useTheme } from "@/lib/theme"

function ThemeProbe() {
  const { colorTheme, setColorTheme } = useTheme()
  return <button onClick={() => setColorTheme("sukhoi")}>{colorTheme}</button>
}

describe("ThemeProvider", () => {
  beforeEach(() => {
    window.localStorage.clear()
    document.documentElement.className = ""
    delete document.documentElement.dataset.colorTheme
  })

  it("persists and applies the selected color theme", async () => {
    const user = userEvent.setup()
    render(<ThemeProvider><ThemeProbe /></ThemeProvider>)

    expect(screen.getByRole("button", { name: "karma" })).toBeInTheDocument()
    expect(document.documentElement.dataset.colorTheme).toBe("karma")

    await user.click(screen.getByRole("button", { name: "karma" }))

    expect(screen.getByRole("button", { name: "sukhoi" })).toBeInTheDocument()
    expect(document.documentElement.dataset.colorTheme).toBe("sukhoi")
    expect(window.localStorage.getItem("shogun.color-theme")).toBe("sukhoi")
  })
})
