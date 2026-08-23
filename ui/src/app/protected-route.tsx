import { Navigate, Outlet, useLocation } from "react-router-dom"

import { useAuth } from "@/features/auth/auth-provider"

export function ProtectedRoute() {
  const { isAuthenticated } = useAuth()
  const location = useLocation()
  return isAuthenticated ? <Outlet /> : <Navigate to="/login" replace state={{ from: location }} />
}

export function AdminRoute() {
  const { claims } = useAuth()
  return claims?.role === "admin" ? <Outlet /> : <Navigate to="/pipelines" replace />
}

export function PublicOnlyRoute() {
  const { isAuthenticated } = useAuth()
  return isAuthenticated ? <Navigate to="/pipelines" replace /> : <Outlet />
}
