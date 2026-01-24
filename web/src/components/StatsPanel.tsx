import type { Stats } from '../view-models'

type Props = {
  stats: Stats
}

export default function StatsPanel({ stats }: Props) {
  return (
    <div className="flex flex-wrap items-center gap-4 text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">
      <span>Reads: {stats.readCount}</span>
      <span>Errors: {stats.errorCount}</span>
      <span>Last: {stats.lastLatencyMs} ms</span>
    </div>
  )
}
