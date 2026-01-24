import { useState } from 'react'
import type { LogEntry } from '../view-models'
import { Button } from './ui/button'

type Props = {
  logs: LogEntry[]
}

export default function RawLog({ logs }: Props) {
  const [open, setOpen] = useState(false)

  return (
    <section className="space-y-2">
      <div className="flex items-center justify-between gap-2">
        <div>
          <p>Raw frames</p>
          <h3>Frame log</h3>
        </div>
        <Button size="sm" variant="outline" onClick={() => setOpen((prev) => !prev)}>
          {open ? 'Hide' : 'Show'} ({logs.length})
        </Button>
      </div>
      {open && (
        <div className="max-h-72 overflow-auto">
          {logs.length === 0 && <p>No frames captured yet.</p>}
          {logs.map((entry, index) => (
            <div key={`${entry.time}-${index}`} className="flex gap-2">
              <span>{new Date(entry.time).toLocaleTimeString()}</span>
              <span>{entry.direction}</span>
              <span>{entry.message}</span>
            </div>
          ))}
        </div>
      )}
      {!open && <p>Expand to view captured frames.</p>}
    </section>
  )
}
