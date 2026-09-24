import { describe, expect, it } from "vitest"

import { runNumbersById } from "@/features/pipelines/run-numbers"

describe("runNumbersById", () => {
  it("numbers a pipeline's runs from one despite global IDs and newest-first responses", () => {
    expect([...runNumbersById([{ id: 11 }, { id: 5 }, { id: 2 }])]).toEqual([
      [2, 1],
      [5, 2],
      [11, 3],
    ])
    expect(runNumbersById([{ id: 12 }]).get(12)).toBe(1)
  })
})
