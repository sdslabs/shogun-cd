import * as React from "react"
import { Menu, Moon, Palette, Sun } from "lucide-react"
import { Outlet, useLocation } from "react-router-dom"

import { AppSidebar } from "@/app/app-sidebar"
import { Button } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuContent, DropdownMenuLabel, DropdownMenuRadioGroup, DropdownMenuRadioItem, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip"
import { useTheme } from "@/lib/theme"
import { cn } from "@/lib/utils"

const COLLAPSED_KEY = "shogun.sidebar-collapsed"

function getLocationLabel(pathname: string) {
  if (pathname.includes("/runs/")) return "Execution detail"
  if (pathname.startsWith("/pipelines/") && pathname.endsWith("/runs")) return "Run history"
  if (pathname.startsWith("/pipelines/")) return "Pipeline"
  if (pathname.startsWith("/pipelines")) return "Pipelines"
  if (pathname.startsWith("/targets/")) return "Target"
  if (pathname.startsWith("/targets")) return "Targets"
  if (pathname.startsWith("/webhooks")) return "Webhooks"
  if (pathname.startsWith("/secrets")) return "Secrets"
  return "Shogun"
}

export function AppShell() {
  const location = useLocation()
  const { theme, resolvedTheme, setTheme, colorTheme, setColorTheme } = useTheme()
  const [collapsed, setCollapsed] = React.useState(() => window.localStorage.getItem(COLLAPSED_KEY) === "true")
  const [mobileOpen, setMobileOpen] = React.useState(false)

  const updateCollapsed = (next: boolean) => {
    window.localStorage.setItem(COLLAPSED_KEY, String(next))
    setCollapsed(next)
  }

  return (
    <TooltipProvider>
      <div className="min-h-screen bg-background">
        <AppSidebar collapsed={collapsed} onCollapsedChange={updateCollapsed} mobileOpen={mobileOpen} onMobileOpenChange={setMobileOpen} />
        <div className={cn("min-h-screen transition-[margin] duration-200", collapsed ? "lg:ml-[68px]" : "lg:ml-[236px]")}>
          <div className="sticky top-0 z-30 flex h-12 items-center border-b border-border/70 bg-background/88 px-4 backdrop-blur-xl sm:px-6 lg:px-8">
            <Button variant="ghost" size="icon-sm" className="mr-2 lg:hidden" onClick={() => setMobileOpen(true)} aria-label="Open navigation"><Menu /></Button>
            <span className="text-xs font-medium text-muted-foreground">{getLocationLabel(location.pathname)}</span>
            <span className="mx-3 h-3 w-px bg-border" />
            <span className="flex items-center gap-1.5 text-[11px] text-muted-foreground"><span className="size-1.5 rounded-full bg-muted-foreground/65" />HTTP polling</span>
            <DropdownMenu>
              <Tooltip>
                <TooltipTrigger asChild>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon-sm" className="ml-auto" aria-label={`Color theme: ${colorTheme === "karma" ? "Karma" : "Sukhoi"}`}>
                      <Palette />
                    </Button>
                  </DropdownMenuTrigger>
                </TooltipTrigger>
                <TooltipContent>Choose theme</TooltipContent>
              </Tooltip>
              <DropdownMenuContent align="end" className="min-w-52">
                <DropdownMenuLabel>Color theme</DropdownMenuLabel>
                <DropdownMenuRadioGroup value={colorTheme} onValueChange={(value) => setColorTheme(value === "sukhoi" ? "sukhoi" : "karma")}>
                  <DropdownMenuRadioItem value="karma">Karma</DropdownMenuRadioItem>
                  <DropdownMenuRadioItem value="sukhoi">Sukhoi</DropdownMenuRadioItem>
                </DropdownMenuRadioGroup>
                <DropdownMenuSeparator />
                <DropdownMenuLabel>Appearance</DropdownMenuLabel>
                <DropdownMenuRadioGroup value={theme} onValueChange={(value) => setTheme(value === "light" || value === "dark" ? value : "system")}>
                  <DropdownMenuRadioItem value="light">Light</DropdownMenuRadioItem>
                  <DropdownMenuRadioItem value="dark">Dark</DropdownMenuRadioItem>
                  <DropdownMenuRadioItem value="system">System</DropdownMenuRadioItem>
                </DropdownMenuRadioGroup>
              </DropdownMenuContent>
            </DropdownMenu>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="ghost" size="icon-sm" onClick={() => setTheme(resolvedTheme === "dark" ? "light" : "dark")} aria-label={`Use ${resolvedTheme === "dark" ? "light" : "dark"} theme`}>
                  {resolvedTheme === "dark" ? <Sun /> : <Moon />}
                </Button>
              </TooltipTrigger>
              <TooltipContent>{resolvedTheme === "dark" ? "Light mode" : "Dark mode"}</TooltipContent>
            </Tooltip>
          </div>
          <main className="min-w-0"><div key={location.pathname} className="route-transition"><Outlet /></div></main>
        </div>
      </div>
    </TooltipProvider>
  )
}
