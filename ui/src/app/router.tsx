import { createBrowserRouter, Navigate } from "react-router-dom"

import { AdminRoute, ProtectedRoute, PublicOnlyRoute } from "@/app/protected-route"
import { AppShell } from "@/app/app-shell"
import { NotFoundPage } from "@/app/not-found-page"
import { AuthLayout } from "@/features/auth/auth-layout"

export const router = createBrowserRouter([
  {
    element: <PublicOnlyRoute />,
    children: [{
      element: <AuthLayout />,
      children: [
        { path: "/login", lazy: async () => ({ Component: (await import("@/features/auth/login-page")).LoginPage }) },
        { path: "/register", lazy: async () => ({ Component: (await import("@/features/auth/register-page")).RegisterPage }) },
      ],
    }],
  },
  {
    element: <ProtectedRoute />,
    children: [{
      element: <AppShell />,
      children: [
        { index: true, element: <Navigate to="/pipelines" replace /> },
        { path: "/pipelines", lazy: async () => ({ Component: (await import("@/features/pipelines/pipelines-page")).PipelinesPage }) },
        { path: "/pipelines/:pipeline", lazy: async () => ({ Component: (await import("@/features/pipelines/pipeline-page")).PipelinePage }) },
        { path: "/pipelines/:pipeline/runs", element: <Navigate to=".." replace /> },
        { path: "/pipelines/:pipeline/runs/:runId", lazy: async () => ({ Component: (await import("@/features/pipelines/run-page")).RunPage }) },
        { path: "/targets", lazy: async () => ({ Component: (await import("@/features/targets/targets-page")).TargetsPage }) },
        { path: "/targets/:target", lazy: async () => ({ Component: (await import("@/features/targets/target-page")).TargetPage }) },
        { path: "/webhooks", lazy: async () => ({ Component: (await import("@/features/webhooks/webhooks-page")).WebhooksPage }) },
        {
          element: <AdminRoute />,
          children: [{ path: "/secrets", lazy: async () => ({ Component: (await import("@/features/secrets/secrets-page")).SecretsPage }) }],
        },
        { path: "*", element: <NotFoundPage /> },
      ],
    }],
  },
])
