import * as React from "react"
import { useQueryClient } from "@tanstack/react-query"

import { api } from "@/lib/api/endpoints"
import type { LoginInput } from "@/lib/api/types"
import {
  clearSessionToken,
  readSession,
  saveSessionToken,
  type SessionClaims,
} from "@/features/auth/session"

interface AuthContextValue {
  claims: SessionClaims | null
  isAuthenticated: boolean
  login: (input: LoginInput) => Promise<void>
  logout: () => void
}

const AuthContext = React.createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const queryClient = useQueryClient()
  const [claims, setClaims] = React.useState<SessionClaims | null>(() => readSession()?.claims ?? null)

  const logout = React.useCallback(() => {
    clearSessionToken()
    setClaims(null)
    queryClient.clear()
  }, [queryClient])

  React.useEffect(() => {
    window.addEventListener("shogun:unauthorized", logout)
    return () => window.removeEventListener("shogun:unauthorized", logout)
  }, [logout])

  const login = React.useCallback(async (input: LoginInput) => {
    const result = await api.login(input)
    saveSessionToken(result.token)
    const session = readSession()
    if (!session) throw new Error("The server returned an invalid session")
    setClaims(session.claims)
  }, [])

  const value = React.useMemo(
    () => ({ claims, isAuthenticated: Boolean(claims), login, logout }),
    [claims, login, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = React.useContext(AuthContext)
  if (!context) throw new Error("useAuth must be used inside AuthProvider")
  return context
}
