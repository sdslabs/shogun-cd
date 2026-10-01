import { QueryClientProvider } from "@tanstack/react-query"
import { RouterProvider } from "react-router-dom"
import { Toaster } from "sonner"

import { AuthProvider } from "@/features/auth/auth-provider"
import { ThemeProvider } from "@/lib/theme"
import { queryClient } from "@/lib/query"
import { router } from "@/app/router"

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <AuthProvider>
          <RouterProvider router={router} />
          <Toaster richColors closeButton position="bottom-right" />
        </AuthProvider>
      </ThemeProvider>
    </QueryClientProvider>
  )
}
