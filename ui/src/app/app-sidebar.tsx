import {
  ChevronLeft,
  KeyRound,
  LogOut,
  PanelLeftClose,
  PanelLeftOpen,
  Server,
  Webhook,
  Workflow,
} from "lucide-react"
import { Link, NavLink } from "react-router-dom"

import { Brand } from "@/components/brand"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { useAuth } from "@/features/auth/auth-provider"
import { cn } from "@/lib/utils"

const mainNavigation = [
  { label: "Pipelines", href: "/pipelines", icon: Workflow },
  { label: "Targets", href: "/targets", icon: Server },
  { label: "Webhooks", href: "/webhooks", icon: Webhook },
]

export function AppSidebar({ collapsed, onCollapsedChange, mobileOpen, onMobileOpenChange }: { collapsed: boolean; onCollapsedChange: (collapsed: boolean) => void; mobileOpen: boolean; onMobileOpenChange: (open: boolean) => void }) {
  const { claims, logout } = useAuth()

  return (
    <>
      {mobileOpen ? <button type="button" aria-label="Close navigation" className="fixed inset-0 z-40 bg-black/35 backdrop-blur-[1px] lg:hidden" onClick={() => onMobileOpenChange(false)} /> : null}
      <aside className={cn("fixed inset-y-0 left-0 z-50 flex flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground transition-[width,transform] duration-200 lg:translate-x-0", collapsed ? "w-[68px]" : "w-[236px]", mobileOpen ? "translate-x-0" : "-translate-x-full")}>
        <div className={cn("flex h-16 items-center border-b border-sidebar-border px-4", collapsed && "justify-center px-2")}>
          <Link to="/pipelines" onClick={() => onMobileOpenChange(false)}><Brand compact={collapsed} /></Link>
          <Button variant="ghost" size="icon-sm" className="ml-auto lg:hidden" onClick={() => onMobileOpenChange(false)} aria-label="Close sidebar"><ChevronLeft /></Button>
        </div>

        <nav className="flex-1 overflow-y-auto px-2.5 py-5">
          {!collapsed ? <p className="mb-2 px-2 text-[9px] font-semibold tracking-[0.19em] text-muted-foreground uppercase">Operate</p> : null}
          <div className="space-y-1">
            {mainNavigation.map((item) => <SidebarLink key={item.href} {...item} collapsed={collapsed} onClick={() => onMobileOpenChange(false)} />)}
          </div>

          {claims?.role === "admin" ? (
            <>
              <Separator className="my-5 bg-sidebar-border" />
              {!collapsed ? <p className="mb-2 px-2 text-[9px] font-semibold tracking-[0.19em] text-muted-foreground uppercase">Administration</p> : null}
              <SidebarLink label="Secrets" href="/secrets" icon={KeyRound} collapsed={collapsed} onClick={() => onMobileOpenChange(false)} />
            </>
          ) : null}
        </nav>

        <div className="border-t border-sidebar-border p-2.5">
          <div className={cn("mb-2 flex items-center gap-2 rounded-lg px-2 py-2", collapsed && "justify-center px-0")}>
            <span className="flex size-8 shrink-0 items-center justify-center rounded-full border border-sidebar-border bg-background text-[11px] font-semibold uppercase text-foreground">
              {claims?.email.slice(0, 2)}
            </span>
            {!collapsed ? (
              <span className="min-w-0 flex-1">
                <span className="block truncate text-xs font-medium">{claims?.email}</span>
                <span className="mt-0.5 block text-[10px] capitalize text-muted-foreground">{claims?.role}</span>
              </span>
            ) : null}
            {!collapsed ? <Button variant="ghost" size="icon-sm" onClick={logout} aria-label="Sign out"><LogOut /></Button> : null}
          </div>
          <Button variant="ghost" size={collapsed ? "icon" : "sm"} className={cn("hidden text-muted-foreground lg:flex", !collapsed && "w-full justify-start")} onClick={() => onCollapsedChange(!collapsed)}>
            {collapsed ? <PanelLeftOpen /> : <PanelLeftClose />}
            {!collapsed ? "Collapse sidebar" : <span className="sr-only">Expand sidebar</span>}
          </Button>
        </div>
      </aside>
    </>
  )
}

function SidebarLink({ label, href, icon: Icon, collapsed, onClick }: { label: string; href: string; icon: typeof Workflow; collapsed: boolean; onClick: () => void }) {
  const link = (
    <NavLink
      to={href}
      onClick={onClick}
      className={({ isActive }) => cn(
        "group flex h-10 items-center gap-3 rounded-lg px-3 text-[13px] font-medium text-muted-foreground transition-colors hover:bg-sidebar-accent/70 hover:text-sidebar-foreground",
        isActive && "bg-sidebar-accent text-sidebar-foreground shadow-[inset_2px_0_0_var(--primary)]",
        collapsed && "justify-center px-0",
      )}
    >
      <Icon className="size-[17px] shrink-0" strokeWidth={1.8} />
      {!collapsed ? <span>{label}</span> : null}
    </NavLink>
  )
  return collapsed ? <Tooltip><TooltipTrigger asChild>{link}</TooltipTrigger><TooltipContent side="right">{label}</TooltipContent></Tooltip> : link
}
