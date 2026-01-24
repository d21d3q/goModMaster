import { useCallback, useEffect, useState } from 'react'
import ConfigForm from './components/ConfigForm'
import DecoderPanel from './components/DecoderPanel'
import DisplayPanel from './components/DisplayPanel'
import RawLog from './components/RawLog'
import ReadPanel from './components/ReadPanel'
import StatsPanel from './components/StatsPanel'
import { Badge } from './components/ui/badge'
import { Button } from './components/ui/button'
import { Tooltip, TooltipContent, TooltipTrigger } from './components/ui/tooltip'
import { CircleHelp, Github } from 'lucide-react'
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarProvider,
  SidebarRail,
  SidebarSeparator,
  SidebarTrigger,
} from './components/ui/sidebar'
import type { Config } from './types'
import type { LogEntry, ReadKind, ReadResult, Stats, WsEvent } from './view-models'

type ConfigResponse = {
  config: Config
  invocation: string
}

type PendingRead = {
  kind: ReadKind
  address: number
  quantity: number
  unitId: number
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
  const [autoConnect, setAutoConnect] = useState(true)
  const [columns, setColumns] = useState(8)
  const [version, setVersion] = useState('')
  const [connectionError, setConnectionError] = useState('')
  const [addressError, setAddressError] = useState('')
  const [quantityError, setQuantityError] = useState('')
  const [showLogs, setShowLogs] = useState(false)
  const [pendingRead, setPendingRead] = useState<PendingRead | null>(null)

  useEffect(() => {
    const headers = token ? { 'X-GMM-Token': token } : undefined
    fetch(`${baseUrl}/api/config`, { headers })
      .then((res) => res.json())
      .then((data: ConfigResponse) => {
        setConfig(data.config)
        setInvocation(data.invocation)
      })
      .catch(() => undefined)

    fetch(`${baseUrl}/api/stats`, { headers })
      .then((res) => res.json())
      .then((data: Stats) => setStats(data))
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
      }
      if (payload.type === 'status') {
        const status = payload.payload as { connected: boolean; connecting: boolean; lastError?: string }
        const isConnected = Boolean(status.connected)
        const isConnecting = Boolean(status.connecting)
        setConnected(isConnected)
        setConnecting(isConnecting)
        setConnectionError(typeof status.lastError === 'string' ? status.lastError : '')
        if (!isConnected && !isConnecting) {
          setPendingRead(null)
        }
      }
    }

    return () => {
      ws.close()
    }
  }, [token])

  const updateConfig = (next: Config, reconnect: boolean) => {
    const headers = buildJsonHeaders(token)
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
          return apiPost('/api/disconnect', token).then(() => apiPost('/api/connect', token))
        }
      })
      .catch(() => undefined)
  }

  const handleConnect = () => {
    apiPost('/api/connect', token).catch(() => undefined)
  }

  const handleDisconnect = () => {
    apiPost('/api/disconnect', token).catch(() => undefined)
  }

  const runRead = useCallback(
    (payload: PendingRead) => {
      const headers = buildJsonHeaders(token)
      return fetch(`${baseUrl}/api/read`, {
        method: 'POST',
        headers,
        body: JSON.stringify(payload),
      })
        .then((res) => res.json())
        .then((data: ReadResult) => {
          setLastResult(data)
        })
        .catch(() => undefined)
    },
    [token],
  )

  useEffect(() => {
    if (!pendingRead || !connected) {
      return
    }
    const payload = pendingRead
    setPendingRead(null)
    runRead(payload)
  }, [pendingRead, connected, runRead])

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

    if (!connected && autoConnect) {
      setPendingRead(payload)
      apiPost('/api/connect', token).catch(() => undefined)
      return
    }
    if (!connected) {
      return
    }
    runRead(payload)
  }

  const updateDecoder = (nextDecoder: { type: string; endianness: string; wordOrder: string; enabled: boolean }) => {
    if (!config) return
    const index = config.decoders.findIndex((decoder) => decoder.type === nextDecoder.type)
    const decoders =
      index >= 0
        ? config.decoders.map((decoder, idx) => (idx === index ? nextDecoder : decoder))
        : [...config.decoders, nextDecoder]
    const next = {
      ...config,
      decoders,
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

  const statusLabel = connected ? 'online' : connecting ? 'connecting' : 'offline'
  const statusVariant = connected ? 'default' : connecting ? 'secondary' : 'outline'
  const actionVariant = connected || connecting ? 'secondary' : 'default'

  return (
    <SidebarProvider defaultOpen>
      <Sidebar collapsible="offcanvas" variant="inset">
        <SidebarHeader className="flex h-12 flex-row items-center gap-0 border-b px-4 py-0">
          <span>Settings</span>
        </SidebarHeader>
        <SidebarContent className="overflow-x-hidden">
            <SidebarGroup>
              <SidebarGroupLabel>Connection settings</SidebarGroupLabel>
              <SidebarGroupContent>
              <ConfigForm
                config={config}
                onSave={(next) => updateConfig(next, connected || connecting)}
                connected={connected}
                connecting={connecting}
              />
              </SidebarGroupContent>
            </SidebarGroup>
          <SidebarSeparator />
          <SidebarGroup>
            <SidebarGroupLabel>
              <span className="flex items-center gap-1">
                Interpretations
                <Tooltip>
                  <TooltipTrigger asChild>
                    <button
                      type="button"
                      className="text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring size-5 inline-flex items-center justify-center rounded-sm"
                      aria-label="Interpretation shortcuts"
                    >
                      <CircleHelp className="size-3.5" />
                    </button>
                  </TooltipTrigger>
                  <TooltipContent>
                    <div className="space-y-1 text-xs">
                      <div>BE = big endian</div>
                      <div>LE = little endian</div>
                      <div>HF = high first</div>
                      <div>LF = low first</div>
                    </div>
                  </TooltipContent>
                </Tooltip>
              </span>
            </SidebarGroupLabel>
            <SidebarGroupContent>
              <DecoderPanel decoders={config?.decoders ?? []} onUpdate={updateDecoder} />
            </SidebarGroupContent>
          </SidebarGroup>
          <SidebarSeparator />
          <SidebarGroup>
            <SidebarGroupLabel>Addressing</SidebarGroupLabel>
            <SidebarGroupContent>
              <DisplayPanel
                config={config}
                columns={columns}
                onColumnsChange={setColumns}
                onAddressBaseChange={setAddressBase}
                onAddressFormatChange={setAddressFormat}
                onValueBaseChange={setValueBase}
              />
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
        <SidebarRail />
      </Sidebar>
      <SidebarInset className="min-w-0 md:peer-data-[variant=inset]:my-0 pt-2">
        <header className="border-b px-4">
          <div className="mx-auto flex h-12 w-full min-w-0 items-center gap-4">
            <div className="flex items-center gap-2 shrink-0">
              <SidebarTrigger />
              <div>
                <h1>GoModMaster</h1>
              </div>
            </div>
            <div className="flex min-w-0 flex-1 items-center justify-end gap-2">
              {connectionError && (
                <span className="min-w-0 max-w-[320px] flex-1 truncate text-xs text-destructive" title={connectionError}>
                  {connectionError}
                </span>
              )}
              <div className="flex items-center gap-2 shrink-0">
                <Badge variant={statusVariant}>{statusLabel}</Badge>
                <Button size="sm" variant={actionVariant} onClick={connected || connecting ? handleDisconnect : handleConnect}>
                  {connected || connecting ? 'Disconnect' : 'Connect'}
                </Button>
              </div>
            </div>
          </div>
        </header>
        <section className="flex-1 p-4">
          <div className="mx-auto w-full space-y-4">
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
          </div>
        </section>
        {showLogs && (
          <div className="border-t bg-background p-4 sticky bottom-0">
            <div className="mx-auto w-full max-w-6xl">
              <RawLog logs={logs} />
            </div>
          </div>
        )}
        <footer className="border-t p-4">
          <div className="mx-auto flex w-full max-w-6xl flex-col gap-2">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="flex items-center gap-3">
                <StatsPanel stats={stats} />
                <Button size="sm" variant="outline" onClick={() => setShowLogs((prev) => !prev)}>
                  {showLogs ? 'Hide logs' : 'Show logs'} ({logs.length})
                </Button>
              </div>
              <span className="flex items-center gap-2">
                goModMaster {version ? `v${version}` : 'MVP'}
                <a
                  href="https://github.com/d21d3q/goModMaster"
                  target="_blank"
                  rel="noreferrer"
                  aria-label="goModMaster on GitHub"
                >
                  <Github className="h-4 w-4" />
                </a>
              </span>
            </div>
            {invocation && (
              <pre className="overflow-x-auto rounded-md bg-muted p-2 text-xs">
                <code>{invocation}</code>
              </pre>
            )}
          </div>
        </footer>
      </SidebarInset>
    </SidebarProvider>
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

function buildJsonHeaders(token: string | null): HeadersInit {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  if (token) {
    headers['X-GMM-Token'] = token
  }
  return headers
}

export default App
