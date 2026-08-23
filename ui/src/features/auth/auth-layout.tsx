import { Link, Outlet } from "react-router-dom"

import { Brand } from "@/components/brand"

export function AuthLayout() {
  return (
    <main className="grid min-h-screen bg-background lg:grid-cols-[minmax(0,1.08fr)_minmax(440px,.92fr)]">
      <section className="relative hidden overflow-hidden border-r bg-[#101817] text-[#edf7f4] lg:flex lg:flex-col">
        <div className="surface-grid absolute inset-0 opacity-25" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_18%_12%,rgba(53,184,155,.1),transparent_32%),radial-gradient(circle_at_80%_80%,rgba(53,184,155,.045),transparent_34%)]" />
        <div className="relative z-10 p-9"><Link to="/login"><Brand /></Link></div>
        <div className="relative z-10 my-auto max-w-2xl px-10 py-16 xl:px-16">
          <p className="mb-5 text-[10px] font-semibold tracking-[0.24em] text-[#58cdb2] uppercase">The deployment room</p>
          <h1 className="text-balance text-[clamp(2.8rem,4.7vw,5.2rem)] leading-[.98] font-semibold tracking-[-0.06em]">Every release,<br /><span className="text-[#c8d8d4]">under command.</span></h1>
          <p className="mt-7 max-w-lg text-base leading-7 text-[#93a8a3]">Follow pipelines from intent to execution. Know what ran, where it ran, and exactly where it stopped.</p>
          <div className="mt-12 flex items-center gap-3" aria-hidden="true">
            {["success", "success", "success", "active", "idle"].map((state, index) => <div className="flex items-center" key={`${state}-${index}`}>{index > 0 ? <span className="h-px w-9 bg-[#344844]" /> : null}<span className={`size-3 rounded-full border ${state === "success" ? "border-[#58cdb2] bg-[#58cdb2]" : state === "active" ? "border-[#7ae0c9] bg-[#7ae0c9] shadow-[0_0_0_5px_rgba(88,205,178,.14)]" : "border-[#48605b] bg-[#172522]"}`} /></div>)}
          </div>
        </div>
        <p className="relative z-10 p-9 text-[11px] text-[#718580]">Self-hosted continuous delivery</p>
      </section>
      <section className="relative flex min-h-screen items-center justify-center px-5 py-12 sm:px-10">
        <div className="absolute top-6 left-6 lg:hidden"><Brand /></div>
        <div className="w-full max-w-[390px]"><Outlet /></div>
      </section>
    </main>
  )
}
