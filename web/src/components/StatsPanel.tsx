import type { Stats } from '../view-models'
import { Badge } from './ui/badge'

type Props = {
  stats: Stats
}

export default function StatsPanel({ stats }: Props) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <Badge variant="secondary">
        Reads {stats.readCount}
      </Badge>
      <Badge variant="secondary">
        Errors {stats.errorCount}
      </Badge>
      <Badge variant="outline">
        Last {stats.lastLatencyMs} ms
      </Badge>
    </div>
  )
}
