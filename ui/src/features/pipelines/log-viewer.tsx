import * as React from "react"
import { Check, Copy, Download, Search, Text, WrapText } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import type { PipelineRunStep } from "@/lib/api/types"
import { cn, titleCase } from "@/lib/utils"

const ANSI_PATTERN = new RegExp(`${String.fromCharCode(27)}\\[([0-9;]*)m`, "g")
const ansiClasses: Record<number, string> = {
  31: "text-red-400",
  32: "text-emerald-400",
  33: "text-amber-300",
  34: "text-sky-400",
  35: "text-fuchsia-400",
  36: "text-cyan-400",
  37: "text-zinc-300",
  92: "text-emerald-300",
  97: "text-white",
}

function renderAnsiLine(line: string) {
  const nodes: React.ReactNode[] = []
  let className = "text-zinc-300"
  let cursor = 0
  let match: RegExpExecArray | null
  ANSI_PATTERN.lastIndex = 0
  while ((match = ANSI_PATTERN.exec(line))) {
    if (match.index > cursor) nodes.push(<span className={className} key={`${cursor}-${match.index}`}>{line.slice(cursor, match.index)}</span>)
    const codes = match[1].split(";").map(Number)
    if (codes.includes(0) || codes.length === 0) className = "text-zinc-300"
    for (const code of codes) if (ansiClasses[code]) className = ansiClasses[code]
    cursor = ANSI_PATTERN.lastIndex
  }
  if (cursor < line.length) nodes.push(<span className={className} key={cursor}>{line.slice(cursor)}</span>)
  return nodes.length ? nodes : " "
}

function buildLogs(steps: PipelineRunStep[], selectedStep: number | null) {
  const visible = selectedStep == null ? steps : steps.filter((step) => step.step_index === selectedStep)
  return visible.map((step) => ({
    step,
    lines: step.logs.replaceAll("\r\n", "\n").split("\n"),
  }))
}

export function LogViewer({ steps, selectedStep }: { steps: PipelineRunStep[]; selectedStep: number | null }) {
  const [search, setSearch] = React.useState("")
  const [wrap, setWrap] = React.useState(false)
  const [copied, setCopied] = React.useState(false)
  const groups = React.useMemo(() => buildLogs(steps, selectedStep), [selectedStep, steps])
  const normalizedSearch = search.toLowerCase()
  const rawOutput = groups.map(({ step, lines }) => `--- Step ${step.step_index + 1}: ${titleCase(step.step_type)} ---\n${lines.join("\n")}`).join("\n\n")

  const copy = async () => { await navigator.clipboard.writeText(rawOutput); setCopied(true); window.setTimeout(() => setCopied(false), 1_500) }
  const download = () => { const url = URL.createObjectURL(new Blob([rawOutput], { type: "text/plain" })); const anchor = document.createElement("a"); anchor.href = url; anchor.download = selectedStep == null ? "shogun-run.log" : `shogun-step-${selectedStep + 1}.log`; anchor.click(); URL.revokeObjectURL(url) }

  return (
    <section className="overflow-hidden rounded-xl border border-zinc-800 bg-[#0d0f11] text-zinc-300 shadow-[0_16px_50px_rgba(0,0,0,.14)]">
      <div className="flex min-h-12 flex-wrap items-center gap-2 border-b border-zinc-800 bg-[#121416] px-3 py-2">
        <div className="flex items-center gap-2 px-1 text-xs font-medium text-zinc-200"><Text className="size-3.5 text-[#58cdb2]" />{selectedStep == null ? "Combined output" : `Step ${selectedStep + 1} output`}</div>
        <div className="relative ml-auto w-44 sm:w-56"><Search className="absolute top-1/2 left-2.5 size-3 -translate-y-1/2 text-zinc-500" /><Input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search logs" className="h-7 border-zinc-700 bg-zinc-900 pl-8 text-xs text-zinc-200 placeholder:text-zinc-600 focus:border-zinc-600" /></div>
        <Button variant="ghost" size="icon-sm" className={cn("text-zinc-500 hover:bg-zinc-800 hover:text-zinc-200", wrap && "bg-zinc-800 text-zinc-200")} onClick={() => setWrap((value) => !value)} aria-label="Toggle line wrapping">{wrap ? <WrapText /> : <Text />}</Button>
        <Button variant="ghost" size="icon-sm" className="text-zinc-500 hover:bg-zinc-800 hover:text-zinc-200" onClick={copy} aria-label="Copy logs">{copied ? <Check className="text-emerald-400" /> : <Copy />}</Button>
        <Button variant="ghost" size="icon-sm" className="text-zinc-500 hover:bg-zinc-800 hover:text-zinc-200" onClick={download} aria-label="Download logs"><Download /></Button>
      </div>
      <div className="max-h-[60vh] min-h-[420px] overflow-auto p-4 font-mono text-[12px] leading-[1.75]">
        {!groups.length ? <p className="text-zinc-600">No persisted logs are available for this selection.</p> : groups.map(({ step, lines }) => (
          <div className="mb-6 last:mb-0" key={step.step_index}>
            {selectedStep == null ? <div className="sticky top-0 z-10 mb-2 flex items-center gap-2 bg-[#0d0f11]/95 py-1 text-[10px] font-semibold tracking-[.12em] text-[#58cdb2] uppercase backdrop-blur"><span className="size-1.5 rounded-full bg-current" />Step {step.step_index + 1} · {titleCase(step.step_type)}</div> : null}
            {lines.map((line, index) => {
              const matches = !normalizedSearch || line.toLowerCase().includes(normalizedSearch)
              return <div key={index} className={cn("flex min-w-max", wrap && "min-w-0", normalizedSearch && matches && "bg-amber-400/8")}><span className="mr-4 w-8 shrink-0 select-none text-right text-zinc-700">{index + 1}</span><code className={cn("block", wrap ? "whitespace-pre-wrap break-words" : "whitespace-pre")}>{renderAnsiLine(line)}</code></div>
            })}
          </div>
        ))}
      </div>
    </section>
  )
}
