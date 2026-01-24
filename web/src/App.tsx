import { useEffect, useState } from 'react'
import ConfigForm from './components/ConfigForm'
import DecoderPanel from './components/DecoderPanel'
import DisplayPanel from './components/DisplayPanel'
import RawLog from './components/RawLog'
import ReadPanel from './components/ReadPanel'
import StatsPanel from './components/StatsPanel'
import type { Config } from './types'
import type { LogEntry, ReadKind, ReadResult, Stats, WsEvent } from './view-models'

type ConfigResponse = {
  config: Config
  invocation: string
}

const baseUrl = ''

function App() {
  const token = new URLSearchParams(window.location.search).get('token')
  const [config, setConfig] = useState<Config | null>(null)
  const [invocation, setInvocation] = useState('')
  const [selectedKind, setSelectedKind] = useState<ReadKind>('holding_registers')
  const [addressInput, setAddressInput] = useState('0')
  const [quantity, setQuantity] = useState(1)
  const [lastResult, setLastResult] = useState<ReadResult | null>(null)
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [stats, setStats] = useState<Stats>({ readCount: 0, errorCount: 0, lastLatencyMs: 0 })
  const [connected, setConnected] = useState(false)
  const [connecting, setConnecting] = useState(false)
  const [error, setError] = useState('')
  const [autoConnect, setAutoConnect] = useState(true)
  const [columns, setColumns] = useState(8)
  const [version, setVersion] = useState('')
  const [addressError, setAddressError] = useState('')
  const [quantityError, setQuantityError] = useState('')

  useEffect(() => {
    const headers = token ? { 'X-GMM-Token': token } : undefined
    fetch(`${baseUrl}/api/config`, { headers })
      .then((res) => res.json())
      .then((data: ConfigResponse) => {
        setConfig(data.config)
        setInvocation(data.invocation)
      })
      .catch((err) => setError(err.message))

    fetch(`${baseUrl}/api/stats`, { headers })
      .then((res) => res.json())
      .then((data: Stats) => setStats(data))
      .catch(() => undefined)

    fetch(`${baseUrl}/api/status`, { headers })
      .then((res) => res.json())
      .then((data: { connected: boolean; connecting: boolean }) => {
        setConnected(data.connected)
        setConnecting(data.connecting)
      })
      .catch(() => undefined)

    fetch(`${baseUrl}/api/version`, { headers })
      .then((res) => res.json())
      .then((data: { version: string }) => setVersion(data.version))
      .catch(() => undefined)
  }, [token])

  useEffect(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
    const wsUrl = token
      ? `${protocol}://${window.location.host}/ws?token=${encodeURIComponent(token)}`
      : `${protocol}://${window.location.host}/ws`
    const ws = new WebSocket(wsUrl)

    ws.onmessage = (event) => {
      const payload: WsEvent = JSON.parse(event.data)
      if (payload.type === 'data') {
        setLastResult(payload.payload as ReadResult)
      }
      if (payload.type === 'log') {
        const entry = payload.payload as LogEntry
        setLogs((prev) => [...prev.slice(-499), entry])
      }
      if (payload.type === 'stats') {
        setStats(payload.payload as Stats)
      }
      if (payload.type === 'error') {
        const result = payload.payload as ReadResult
        setLastResult(result)
        setError(result.errorMessage ?? 'Unknown error')
      }
      if (payload.type === 'status') {
        const status = payload.payload as { connected: boolean; connecting: boolean; lastError?: string }
        setConnected(Boolean(status.connected))
        setConnecting(Boolean(status.connecting))
        if (status.lastError) {
          setError(status.lastError)
        }
      }
    }

    return () => {
      ws.close()
    }
  }, [token])

  const updateConfig = (next: Config, reconnect: boolean) => {
    const headers = token
      ? { 'Content-Type': 'application/json', 'X-GMM-Token': token }
      : { 'Content-Type': 'application/json' }
    fetch(`${baseUrl}/api/config`, {
      method: 'POST',
      headers,
      body: JSON.stringify(next),
    })
      .then((res) => res.json())
      .then((data: ConfigResponse) => {
        setConfig(data.config)
        setInvocation(data.invocation)
        if (reconnect) {
          return apiPost('/api/disconnect', token)
            .then(() => apiPost('/api/connect', token))
            .then((data) => {
              setConnected(Boolean(data.connected))
              setConnecting(Boolean(data.connecting))
            })
        }
      })
      .catch((err) => setError(err.message))
  }

  const handleConnect = () => {
    apiPost('/api/connect', token)
      .then((data) => {
        setConnected(Boolean(data.connected))
        setConnecting(Boolean(data.connecting))
      })
      .catch((err) => setError(err.message))
  }

  const handleDisconnect = () => {
    apiPost('/api/disconnect', token)
      .then((data) => {
        setConnected(Boolean(data.connected))
        setConnecting(Boolean(data.connecting))
      })
      .catch((err) => setError(err.message))
  }

  const handleRead = () => {
    if (!config) return
    const parsedAddress = parseAddress(addressInput)
    if (parsedAddress === null) {
      setAddressError('Invalid address')
      return
    }
    if (!Number.isFinite(quantity) || quantity < 1) {
      setQuantityError('Quantity must be >= 1')
      return
    }
    const payload = {
      kind: selectedKind,
      address: parsedAddress,
      quantity,
      unitId: config.unitId,
    }
    const headers = token
      ? { 'Content-Type': 'application/json', 'X-GMM-Token': token }
      : { 'Content-Type': 'application/json' }

    const runRead = () =>
      fetch(`${baseUrl}/api/read`, {
        method: 'POST',
        headers,
        body: JSON.stringify(payload),
      })
        .then((res) => res.json())
        .then((data: ReadResult) => {
          setLastResult(data)
        })
        .catch((err) => setError(err.message))

    if (!connected && autoConnect) {
      apiPost('/api/connect', token)
        .then((data) => {
          const isConnected = Boolean(data.connected)
          setConnected(isConnected)
          setConnecting(Boolean(data.connecting))
          if (!isConnected) {
            return
          }
          return runRead()
        })
        .catch((err) => setError(err.message))
      return
    }
    if (!connected) {
      setError('Not connected')
      return
    }
    runRead()
  }

  const updateDecoder = (nextDecoder: { type: string; endianness: string; wordOrder: string; enabled: boolean }) => {
    if (!config) return
    const existing = config.decoders.filter((decoder) => decoder.type !== nextDecoder.type)
    const next = {
      ...config,
      decoders: [...existing, nextDecoder],
    }
    updateConfig(next, false)
  }

  const setAddressBase = (base: number) => {
    if (!config) return
    updateConfig({ ...config, addressBase: base }, false)
  }

  const setAddressFormat = (format: number) => {
    if (!config) return
    updateConfig({ ...config, addressFormat: format }, false)
  }

  const setValueBase = (base: number) => {
    if (!config) return
    updateConfig({ ...config, valueBase: base }, false)
  }

  const handleAddressChange = (value: string) => {
    setAddressInput(value)
    const parsed = parseAddress(value)
    if (parsed === null) {
      setAddressError('Invalid address')
    } else {
      setAddressError('')
    }
  }

  const handleQuantityChange = (value: number) => {
    setQuantity(value)
    if (!Number.isFinite(value) || value < 1) {
      setQuantityError('Quantity must be >= 1')
    } else {
      setQuantityError('')
    }
  }

  return (
    <div className="min-h-screen bg-slate-100 text-slate-900">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-4 px-6 py-4">
          <div>
            <p className="text-xs uppercase tracking-[0.25em] text-slate-500">goModMaster</p>
            <h1 className="text-2xl font-semibold">Real-time Modbus diagnostics</h1>
          </div>
          <div className="flex items-center gap-3">
            <div className="rounded-full border border-slate-200 bg-slate-100 px-4 py-2 text-xs font-semibold uppercase tracking-[0.2em] text-slate-600">
              {connected ? 'online' : connecting ? 'connecting' : 'offline'}
            </div>
            <button
              className={`rounded-full px-4 py-2 text-sm font-semibold ${
                connected
                  ? 'bg-emerald-500 text-white'
                  : connecting
                    ? 'bg-amber-500 text-white'
                    : 'bg-slate-900 text-white'
              }`}
              onClick={connected || connecting ? handleDisconnect : handleConnect}
            >
              {connected || connecting ? 'Disconnect' : 'Connect'}
            </button>
          </div>
        </div>
      </header>

      <main className="mx-auto flex max-w-6xl flex-col gap-3 px-6 py-5">
        <div className="rounded-3xl border border-slate-200 bg-white p-3 shadow-sm">
          <div className="flex items-center justify-between gap-4">
            <h2 className="text-lg font-semibold">Connection settings</h2>
          </div>
          <div className="mt-2">
            <ConfigForm config={config} onSave={(next) => updateConfig(next, connected)} connected={connected} />
          </div>
          <div className="mt-3 border-t border-slate-200 pt-3">
            <div className="flex items-center justify-between gap-3">
              <h3 className="text-sm font-semibold uppercase tracking-[0.2em] text-slate-500">Interpretations</h3>
            </div>
            <div className="mt-2">
              <DecoderPanel decoders={config?.decoders ?? []} onUpdate={updateDecoder} />
            </div>
          </div>
        </div>

        <DisplayPanel
          config={config}
          columns={columns}
          onColumnsChange={setColumns}
          onAddressBaseChange={setAddressBase}
          onAddressFormatChange={setAddressFormat}
          onValueBaseChange={setValueBase}
        />
        <div className="grid gap-3 lg:grid-cols-[2fr_1fr]">
          <section className="space-y-4">
            <ReadPanel
              selectedKind={selectedKind}
              addressInput={addressInput}
              quantity={quantity}
              addressBase={config?.addressBase ?? 0}
              addressFormat={config?.addressFormat ?? 10}
              valueBase={config?.valueBase ?? 10}
              decoders={config?.decoders ?? []}
              lastResult={lastResult}
              columns={columns}
              connected={connected}
              autoConnect={autoConnect}
              addressError={addressError}
              quantityError={quantityError}
              onKindChange={setSelectedKind}
              onAddressChange={handleAddressChange}
              onQuantityChange={handleQuantityChange}
              onRead={handleRead}
              onAutoConnectChange={setAutoConnect}
            />
            <RawLog logs={logs} />
          </section>
          <aside className="space-y-6"></aside>
        </div>
      </main>

      <div className="border-t border-slate-200 bg-white">
        <div className="mx-auto flex max-w-6xl flex-col gap-2 px-6 py-3 text-xs text-slate-500">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <span>goModMaster {version ? `v${version}` : 'MVP'} · read-only</span>
            <span>{error ? `Last error: ${error}` : 'Ready for requests'}</span>
            {lastResult && (
              <span className="text-slate-600">
                Last: {lastResult.kind} @ {lastResult.address} · {lastResult.latencyMs} ms
              </span>
            )}
          </div>
          <div className="flex flex-wrap items-center justify-between gap-4">
            <StatsPanel stats={stats} />
          </div>
          {invocation && <div className="font-mono text-[11px] text-slate-600">{invocation}</div>}
        </div>
      </div>
    </div>
  )
}

function parseAddress(input: string): number | null {
  const trimmed = input.trim().toLowerCase()
  if (trimmed === '') {
    return null
  }
  if (trimmed.startsWith('0x')) {
    const hex = trimmed.slice(2)
    if (!/^[0-9a-f]+$/.test(hex)) {
      return null
    }
    return Number.parseInt(hex, 16)
  }
  if (!/^[0-9]+$/.test(trimmed)) {
    return null
  }
  const value = Number.parseInt(trimmed, 10)
  if (Number.isNaN(value)) {
    return null
  }
  return value
}

function apiPost(path: string, token: string | null): Promise<any> {
  return fetch(`${baseUrl}${path}`, {
    method: 'POST',
    headers: token ? { 'X-GMM-Token': token } : undefined,
  }).then(async (res) => {
    const data = await res.json().catch(() => ({}))
    if (!res.ok) {
      const message = typeof data?.error === 'string' ? data.error : res.statusText
      throw new Error(message || 'Request failed')
    }
    return data
  })
}

export default App
