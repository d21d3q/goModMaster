import { useEffect, useRef, useState } from 'react'
import type { LogEntry } from '../view-models'

type Props = {
  logs: LogEntry[]
}

export default function RawLog({ logs }: Props) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const [atBottom, setAtBottom] = useState(true)

  useEffect(() => {
    if (!atBottom) return
    const el = containerRef.current
    if (!el) return
    el.scrollTop = el.scrollHeight
  }, [logs, atBottom])

  const handleScroll = () => {
    const el = containerRef.current
    if (!el) return
    const nearBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 4
    setAtBottom(nearBottom)
  }

  return (
    <section className="space-y-2">
      <div ref={containerRef} className="max-h-72 overflow-auto" onScroll={handleScroll}>
        {logs.length === 0 && <p>No frames captured yet.</p>}
        {logs.map((entry, index) => (
          <div key={`${entry.time}-${index}`} className="flex gap-2">
            <span>{new Date(entry.time).toLocaleTimeString()}</span>
            <span>{entry.direction}</span>
            <span>{entry.message}</span>
          </div>
        ))}
      </div>
    </section>
  )
}
