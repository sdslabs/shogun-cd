import { StrictMode } from "react"
import { createRoot } from "react-dom/client"

import "@fontsource-variable/jetbrains-mono"
import "@fontsource-variable/manrope"

import { App } from "@/app/app"
import "@/index.css"

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
