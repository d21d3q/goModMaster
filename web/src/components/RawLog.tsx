import { useState } from 'react'
import type { LogEntry } from '../view-models'

type Props = {
  logs: LogEntry[]
}

export default function RawLog({ logs }: Props) {
  const [open, setOpen] = useState(false)

  return (
    <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-slate-500">Raw frames</p>
          <h2 className="text-xl font-semibold">Frame log</h2>
        </div>
        <button
          className="rounded-full border border-slate-200 px-4 py-2 text-xs font-semibold uppercase tracking-[0.2em] text-slate-600"
          onClick={() => setOpen((prev) => !prev)}
        >
          {open ? 'Hide' : 'Show'} ({logs.length})
        </button>
      </div>
      {open && (
        <div className="mt-4 max-h-72 overflow-auto rounded-2xl border border-slate-200 bg-slate-950 p-4 text-xs text-emerald-200">
          {logs.length === 0 && <p className="text-slate-500">No frames captured yet.</p>}
          {logs.map((entry, index) => (
            <div key={`${entry.time}-${index}`} className="flex gap-3">
              <span className="text-slate-500">{new Date(entry.time).toLocaleTimeString()}</span>
              <span className="uppercase text-cyan-200">{entry.direction}</span>
              <span>{entry.message}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
