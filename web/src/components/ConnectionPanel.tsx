import type { Config } from '../types'
const protocolLabels: Record<string, string> = {
  tcp: 'TCP',
  rtu: 'RTU',
}

type Props = {
  config: Config | null
  invocation: string
}

export default function ConnectionPanel({ config, invocation }: Props) {
  const target = config?.protocol === 'rtu'
    ? config.serial.device
    : `${config?.tcp.host ?? '127.0.0.1'}:${config?.tcp.port ?? 502}`

  return (
    <section className="space-y-2">
      <div>
        <p>Session</p>
        <h3>Connection</h3>
      </div>
      <div className="space-y-2">
        <div className="flex items-center justify-between gap-2">
          <span>Protocol</span>
          <span>{protocolLabels[config?.protocol ?? 'tcp']}</span>
        </div>
        <div className="flex items-center justify-between gap-2">
          <span>Target</span>
          <span>{target}</span>
        </div>
        <div className="flex items-center justify-between gap-2">
          <span>Unit ID</span>
          <span>{config?.unitId ?? 1}</span>
        </div>
        <div className="flex items-center justify-between gap-2">
          <span>Timeout</span>
          <span>{config?.timeoutMs ?? 0} ms</span>
        </div>
        <div>
          <p>Invocation</p>
          <p>{invocation}</p>
        </div>
      </div>
    </section>
  )
}
