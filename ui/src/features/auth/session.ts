import type { UserRole } from "@/lib/api/types"

const TOKEN_KEY = "shogun.session-token"

export interface SessionClaims {
  email: string
  role: UserRole
  exp: number
  iat?: number
}

function decodeBase64Url(value: string) {
  const normalized = value.replaceAll("-", "+").replaceAll("_", "/")
  const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, "=")
  return decodeURIComponent(
    atob(padded)
      .split("")
      .map((char) => `%${char.charCodeAt(0).toString(16).padStart(2, "0")}`)
      .join(""),
  )
}

export function parseSessionToken(token: string): SessionClaims | null {
  try {
    const [, payload] = token.split(".")
    if (!payload) return null
    const claims = JSON.parse(decodeBase64Url(payload)) as Partial<SessionClaims>
    if (!claims.email || !claims.role || !claims.exp) return null
    if (claims.role !== "admin" && claims.role !== "user") return null
    return claims as SessionClaims
  } catch {
    return null
  }
}

export function getSessionToken() {
  return window.localStorage.getItem(TOKEN_KEY)
}

export function saveSessionToken(token: string) {
  window.localStorage.setItem(TOKEN_KEY, token)
}

export function clearSessionToken() {
  window.localStorage.removeItem(TOKEN_KEY)
}

export function readSession() {
  const token = getSessionToken()
  if (!token) return null

  const claims = parseSessionToken(token)
  if (!claims || claims.exp * 1000 <= Date.now()) {
    clearSessionToken()
    return null
  }

  return { token, claims }
}
