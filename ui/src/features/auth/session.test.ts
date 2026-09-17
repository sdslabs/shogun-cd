import { afterEach, describe, expect, it } from "vitest"

import { clearSessionToken, parseSessionToken, readSession, saveSessionToken } from "@/features/auth/session"

function tokenFor(payload: object) {
  const encode = (value: object) => btoa(JSON.stringify(value)).replaceAll("+", "-").replaceAll("/", "_").replaceAll("=", "")
  return `${encode({ alg: "none" })}.${encode(payload)}.signature`
}

describe("session tokens", () => {
  afterEach(clearSessionToken)

  it("reads the email, role, and expiry from a valid token", () => {
    const token = tokenFor({ email: "admin@shogun.dev", role: "admin", exp: Math.floor(Date.now() / 1000) + 60 })
    expect(parseSessionToken(token)).toMatchObject({ email: "admin@shogun.dev", role: "admin" })
  })

  it("removes an expired persisted session", () => {
    saveSessionToken(tokenFor({ email: "user@shogun.dev", role: "user", exp: 1 }))
    expect(readSession()).toBeNull()
    expect(window.localStorage.length).toBe(0)
  })

  it("rejects unknown roles", () => {
    expect(parseSessionToken(tokenFor({ email: "user@shogun.dev", role: "owner", exp: 9999999999 }))).toBeNull()
  })
})
