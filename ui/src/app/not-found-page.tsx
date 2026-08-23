import { ArrowLeft, Workflow } from "lucide-react"
import { Link } from "react-router-dom"

import { Button } from "@/components/ui/button"

export function NotFoundPage() {
  return <div className="flex min-h-[calc(100vh-3rem)] items-center justify-center px-6 py-16 text-center"><div><span className="mx-auto flex size-12 items-center justify-center rounded-xl border bg-muted text-muted-foreground"><Workflow className="size-5" /></span><p className="mt-6 text-[10px] font-semibold tracking-[.2em] text-primary uppercase">404 · Route missing</p><h1 className="mt-3 text-3xl font-semibold tracking-[-0.045em]">Nothing is deployed here.</h1><p className="mx-auto mt-3 max-w-sm text-sm leading-6 text-muted-foreground">The page may have moved, or the resource is no longer indexed by Shogun.</p><Button asChild variant="outline" className="mt-7"><Link to="/pipelines"><ArrowLeft />Return to pipelines</Link></Button></div></div>
}
