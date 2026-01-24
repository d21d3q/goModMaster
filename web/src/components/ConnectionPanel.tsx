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
    <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
      <p className="text-xs uppercase tracking-[0.2em] text-slate-500">Session</p>
      <h2 className="text-xl font-semibold">Connection</h2>
      <div className="mt-4 space-y-3 text-sm">
        <div className="flex items-center justify-between">
          <span className="text-slate-500">Protocol</span>
          <span className="font-semibold uppercase">{protocolLabels[config?.protocol ?? 'tcp']}</span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-slate-500">Target</span>
          <span className="font-semibold">{target}</span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-slate-500">Unit ID</span>
          <span className="font-semibold">{config?.unitId ?? 1}</span>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-slate-500">Timeout</span>
          <span className="font-semibold">{config?.timeoutMs ?? 0} ms</span>
        </div>
      </div>
      <div className="mt-4 rounded-2xl border border-slate-200 bg-slate-50 p-3 text-xs">
        <p className="font-semibold text-slate-500">Invocation</p>
        <p className="mt-2 break-all font-mono text-[11px] text-slate-800">{invocation}</p>
      </div>
    </div>
  )
}
