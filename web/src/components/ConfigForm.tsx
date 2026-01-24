import { useEffect, useState } from 'react'
import type { Config } from '../types'

type Props = {
  config: Config | null
  onSave: (config: Config) => void
  connected: boolean
}

type SerialDevicesResponse = {
  devices: string[]
}

export default function ConfigForm({ config, onSave, connected }: Props) {
  const [draft, setDraft] = useState<Config | null>(config)
  const [serialDevices, setSerialDevices] = useState<string[]>([])

  useEffect(() => {
    setDraft(config)
  }, [config])

  useEffect(() => {
    if (!draft || draft.protocol !== 'rtu') return
    const token = new URLSearchParams(window.location.search).get('token')
    const headers = token ? { 'X-GMM-Token': token } : undefined
    fetch('/api/serial-devices', { headers })
      .then((res) => res.json())
      .then((data: SerialDevicesResponse) => setSerialDevices(data.devices ?? []))
      .catch(() => setSerialDevices([]))
  }, [draft?.protocol])

  if (!draft) {
    return <p className="text-sm text-slate-500">Waiting for config...</p>
  }

  const update = (partial: Partial<Config>) => {
    setDraft({ ...draft, ...partial })
  }

  return (
    <div className="flex flex-wrap items-center gap-3 text-sm">
      <div className="min-w-[120px]">
        <label className="text-[10px] font-semibold uppercase tracking-[0.15em] text-slate-500">Protocol</label>
        <select
          className="mt-1 w-full rounded-lg border border-slate-200 bg-white px-2 py-2 text-sm"
          value={draft.protocol}
          onChange={(event) => update({ protocol: event.target.value as 'tcp' | 'rtu' })}
        >
          <option value="tcp">TCP</option>
          <option value="rtu">RTU</option>
        </select>
      </div>

      {draft.protocol === 'tcp' ? (
        <>
          <div className="min-w-[180px] flex-1">
            <label className="text-[10px] font-semibold uppercase tracking-[0.15em] text-slate-500">Host</label>
            <input
              className="mt-1 w-full rounded-lg border border-slate-200 bg-white px-2 py-2 text-sm"
              value={draft.tcp.host}
              onChange={(event) => update({ tcp: { ...draft.tcp, host: event.target.value } })}
            />
          </div>
          <div className="min-w-[110px]">
            <label className="text-[10px] font-semibold uppercase tracking-[0.15em] text-slate-500">Port</label>
            <input
              type="number"
              className="mt-1 w-full rounded-lg border border-slate-200 bg-white px-2 py-2 text-sm"
              value={draft.tcp.port}
              onChange={(event) => update({ tcp: { ...draft.tcp, port: Number(event.target.value) } })}
            />
          </div>
        </>
      ) : (
        <>
          {serialDevices.length > 0 && (
            <div className="min-w-[200px] flex-1">
              <label className="text-[10px] font-semibold uppercase tracking-[0.15em] text-slate-500">Detected</label>
              <select
                className="mt-1 w-full rounded-lg border border-slate-200 bg-white px-2 py-2 text-sm"
                value=""
                onChange={(event) => update({ serial: { ...draft.serial, device: event.target.value } })}
              >
                <option value="">Select device</option>
                {serialDevices.map((device) => (
                  <option key={device} value={device}>
                    {device}
                  </option>
                ))}
              </select>
            </div>
          )}
          <div className="min-w-[180px] flex-1">
            <label className="text-[10px] font-semibold uppercase tracking-[0.15em] text-slate-500">Device</label>
            <input
              className="mt-1 w-full rounded-lg border border-slate-200 bg-white px-2 py-2 text-sm"
              value={draft.serial.device}
              onChange={(event) => update({ serial: { ...draft.serial, device: event.target.value } })}
              placeholder="/dev/ttyUSB0"
            />
          </div>
          <div className="min-w-[110px]">
            <label className="text-[10px] font-semibold uppercase tracking-[0.15em] text-slate-500">Speed</label>
            <input
              type="number"
              className="mt-1 w-full rounded-lg border border-slate-200 bg-white px-2 py-2 text-sm"
              value={draft.serial.speed}
              onChange={(event) => update({ serial: { ...draft.serial, speed: Number(event.target.value) } })}
            />
          </div>
          <div className="min-w-[80px]">
            <label className="text-[10px] font-semibold uppercase tracking-[0.15em] text-slate-500">Data</label>
            <input
              type="number"
              className="mt-1 w-full rounded-lg border border-slate-200 bg-white px-2 py-2 text-sm"
              value={draft.serial.dataBits}
              onChange={(event) => update({ serial: { ...draft.serial, dataBits: Number(event.target.value) } })}
            />
          </div>
          <div className="min-w-[80px]">
            <label className="text-[10px] font-semibold uppercase tracking-[0.15em] text-slate-500">Stop</label>
            <input
              type="number"
              className="mt-1 w-full rounded-lg border border-slate-200 bg-white px-2 py-2 text-sm"
              value={draft.serial.stopBits}
              onChange={(event) => update({ serial: { ...draft.serial, stopBits: Number(event.target.value) } })}
            />
          </div>
          <div className="min-w-[110px]">
            <label className="text-[10px] font-semibold uppercase tracking-[0.15em] text-slate-500">Parity</label>
            <select
              className="mt-1 w-full rounded-lg border border-slate-200 bg-white px-2 py-2 text-sm"
              value={draft.serial.parity}
              onChange={(event) => update({ serial: { ...draft.serial, parity: event.target.value } })}
            >
              <option value="none">None</option>
              <option value="even">Even</option>
              <option value="odd">Odd</option>
            </select>
          </div>
        </>
      )}

      <div className="min-w-[90px]">
        <label className="text-[10px] font-semibold uppercase tracking-[0.15em] text-slate-500">Unit</label>
        <input
          type="number"
          className="mt-1 w-full rounded-lg border border-slate-200 bg-white px-2 py-2 text-sm"
          value={draft.unitId}
          onChange={(event) => update({ unitId: Number(event.target.value) })}
        />
      </div>
      <div className="min-w-[120px]">
        <label className="text-[10px] font-semibold uppercase tracking-[0.15em] text-slate-500">Timeout</label>
        <input
          type="number"
          className="mt-1 w-full rounded-lg border border-slate-200 bg-white px-2 py-2 text-sm"
          value={draft.timeoutMs}
          onChange={(event) => update({ timeoutMs: Number(event.target.value) })}
        />
      </div>

      <button
        className="rounded-full bg-slate-900 px-3 py-2 text-[11px] font-semibold uppercase tracking-[0.2em] text-white"
        onClick={() => onSave(draft)}
      >
        Apply
      </button>
    </div>
  )
}
