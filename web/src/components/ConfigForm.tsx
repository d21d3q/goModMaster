import { useEffect, useState } from 'react'
import type { Config } from '../types'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select'

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
    return <p>Waiting for config...</p>
  }

  const update = (partial: Partial<Config>) => {
    setDraft({ ...draft, ...partial })
  }

  const actionLabel = connected ? 'Apply & reconnect' : 'Apply'

  return (
    <div className="grid gap-3 md:grid-cols-2">
      <div className="grid gap-1">
        <Label htmlFor="protocol">Protocol</Label>
        <Select value={draft.protocol} onValueChange={(value) => update({ protocol: value as 'tcp' | 'rtu' })}>
          <SelectTrigger className="w-full" id="protocol">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="tcp">TCP</SelectItem>
            <SelectItem value="rtu">RTU</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {draft.protocol === 'tcp' ? (
        <>
          <div className="grid gap-1 md:col-span-2">
            <Label htmlFor="tcp-host">Host</Label>
            <Input
              id="tcp-host"
              value={draft.tcp.host}
              onChange={(event) => update({ tcp: { ...draft.tcp, host: event.target.value } })}
            />
          </div>
          <div className="grid gap-1">
            <Label htmlFor="tcp-port">Port</Label>
            <Input
              id="tcp-port"
              type="number"
              value={draft.tcp.port}
              onChange={(event) => update({ tcp: { ...draft.tcp, port: Number(event.target.value) } })}
            />
          </div>
        </>
      ) : (
        <>
          {serialDevices.length > 0 && (
            <div className="grid gap-1 md:col-span-2">
              <Label htmlFor="serial-detected">Detected</Label>
              <Select
                value={serialDevices.includes(draft.serial.device) ? draft.serial.device : ''}
                onValueChange={(value) => update({ serial: { ...draft.serial, device: value } })}
              >
                <SelectTrigger className="w-full" id="serial-detected">
                  <SelectValue placeholder="Select device" />
                </SelectTrigger>
                <SelectContent>
                  {serialDevices.map((device) => (
                    <SelectItem key={device} value={device}>
                      {device}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}
          <div className="grid gap-1 md:col-span-2">
            <Label htmlFor="serial-device">Device</Label>
            <Input
              id="serial-device"
              value={draft.serial.device}
              onChange={(event) => update({ serial: { ...draft.serial, device: event.target.value } })}
              placeholder="/dev/ttyUSB0"
            />
          </div>
          <div className="grid gap-1">
            <Label htmlFor="serial-speed">Speed</Label>
            <Input
              id="serial-speed"
              type="number"
              value={draft.serial.speed}
              onChange={(event) => update({ serial: { ...draft.serial, speed: Number(event.target.value) } })}
            />
          </div>
          <div className="grid gap-1">
            <Label htmlFor="serial-data">Data</Label>
            <Input
              id="serial-data"
              type="number"
              value={draft.serial.dataBits}
              onChange={(event) => update({ serial: { ...draft.serial, dataBits: Number(event.target.value) } })}
            />
          </div>
          <div className="grid gap-1">
            <Label htmlFor="serial-stop">Stop</Label>
            <Input
              id="serial-stop"
              type="number"
              value={draft.serial.stopBits}
              onChange={(event) => update({ serial: { ...draft.serial, stopBits: Number(event.target.value) } })}
            />
          </div>
          <div className="grid gap-1">
            <Label htmlFor="serial-parity">Parity</Label>
            <Select
              value={draft.serial.parity}
              onValueChange={(value) => update({ serial: { ...draft.serial, parity: value } })}
            >
              <SelectTrigger className="w-full" id="serial-parity">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="none">None</SelectItem>
                <SelectItem value="even">Even</SelectItem>
                <SelectItem value="odd">Odd</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </>
      )}

      <div className="grid gap-1">
        <Label htmlFor="unit-id">Unit</Label>
        <Input
          id="unit-id"
          type="number"
          value={draft.unitId}
          onChange={(event) => update({ unitId: Number(event.target.value) })}
        />
      </div>
      <div className="grid gap-1">
        <Label htmlFor="timeout-ms">Timeout</Label>
        <Input
          id="timeout-ms"
          type="number"
          value={draft.timeoutMs}
          onChange={(event) => update({ timeoutMs: Number(event.target.value) })}
        />
      </div>

      <div className="flex items-center justify-end md:col-span-2">
        <Button size="sm" onClick={() => onSave(draft)}>
          {actionLabel}
        </Button>
      </div>
    </div>
  )
}
