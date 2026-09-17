import { Link, Outlet } from "react-router-dom"

import { Brand } from "@/components/brand"

export function AuthLayout() {
  return (
    <main className="grid min-h-screen bg-background lg:grid-cols-[minmax(0,1.08fr)_minmax(440px,.92fr)]">
      <section className="relative hidden overflow-hidden border-r bg-[var(--auth-panel)] text-[var(--auth-foreground)] lg:flex lg:flex-col">
        <div className="surface-grid absolute inset-0 opacity-25" />
        <div className="auth-ambient absolute inset-0" />
        <div className="relative z-10 p-9"><Link to="/login"><Brand /></Link></div>
        <div className="relative z-10 my-auto max-w-2xl px-10 py-16 xl:px-16">
          <p className="mb-5 text-[10px] font-semibold tracking-[0.24em] text-[var(--auth-primary)] uppercase">The deployment room</p>
          <h1 className="text-balance text-[clamp(2.8rem,4.7vw,5.2rem)] leading-[.98] font-semibold tracking-[-0.06em]">Every release,<br /><span className="text-[var(--auth-heading-muted)]">under command.</span></h1>
          <p className="mt-7 max-w-lg text-base leading-7 text-[var(--auth-copy)]">Follow pipelines from intent to execution. Know what ran, where it ran, and exactly where it stopped.</p>
          <div className="mt-12 flex items-center gap-3" aria-hidden="true">
            {["success", "success", "success", "active", "idle"].map((state, index) => <div className="flex items-center" key={`${state}-${index}`}>{index > 0 ? <span className="h-px w-9 bg-[var(--auth-connector)]" /> : null}<span className={`size-3 rounded-full border ${state === "success" ? "border-[var(--auth-primary)] bg-[var(--auth-primary)]" : state === "active" ? "auth-active-dot border-[var(--auth-active)] bg-[var(--auth-active)] motion-safe:animate-[auth-beacon-glow_1.8s_ease-in-out_infinite]" : "border-[var(--auth-idle-border)] bg-[var(--auth-idle-background)]"}`} /></div>)}
          </div>
        </div>
        <p className="relative z-10 p-9 text-[11px] text-[var(--auth-footer)]">Self-hosted continuous delivery</p>
      </section>
      <section className="relative flex min-h-screen items-center justify-center px-5 py-12 sm:px-10">
        <div className="absolute top-6 left-6 lg:hidden"><Brand /></div>
        <div className="w-full max-w-[390px]"><Outlet /></div>
      </section>
    </main>
  )
}
