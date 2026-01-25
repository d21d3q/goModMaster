import { useCallback, useEffect, useState } from 'react'
import AppLayout from './components/AppLayout'
import UnauthorizedPanel from './components/UnauthorizedPanel'
import { apiPost, buildJsonHeaders, fetchJson } from './lib/api'
import { parseAddress } from './lib/parse'
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
  const [authBlocked, setAuthBlocked] = useState(() => window.location.hash === '#/401')

  const handleUnauthorized = useCallback(() => {
    setAuthBlocked(true)
    window.location.hash = '#/401'
  }, [])

  useEffect(() => {
    const handleHashChange = () => {
      setAuthBlocked(window.location.hash === '#/401')
    }
    window.addEventListener('hashchange', handleHashChange)
    return () => window.removeEventListener('hashchange', handleHashChange)
  }, [])

  useEffect(() => {
    const headers = token ? { 'X-GMM-Token': token } : undefined
    fetchJson<ConfigResponse>('/api/config', { headers }, handleUnauthorized)
      .then((data: ConfigResponse) => {
        setConfig(data.config)
        setInvocation(data.invocation)
      })
      .catch(() => undefined)

    fetchJson<Stats>('/api/stats', { headers }, handleUnauthorized)
      .then((data: Stats) => setStats(data))
      .catch(() => undefined)

    fetchJson<{ version: string }>('/api/version', { headers }, handleUnauthorized).then((data) =>
      setVersion(data.version),
    )
      .catch(() => undefined)
  }, [handleUnauthorized, token])

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
    fetchJson<ConfigResponse>(
      '/api/config',
      {
        method: 'POST',
        headers,
        body: JSON.stringify(next),
      },
      handleUnauthorized,
    )
      .then((data: ConfigResponse) => {
        setConfig(data.config)
        setInvocation(data.invocation)
        if (reconnect) {
          return apiPost('/api/disconnect', token, handleUnauthorized).then(() =>
            apiPost('/api/connect', token, handleUnauthorized),
          )
        }
      })
      .catch(() => undefined)
  }

  const handleConnect = () => {
    apiPost('/api/connect', token, handleUnauthorized).catch(() => undefined)
  }

  const handleDisconnect = () => {
    apiPost('/api/disconnect', token, handleUnauthorized).catch(() => undefined)
  }

  const runRead = useCallback(
    (payload: PendingRead) => {
      const headers = buildJsonHeaders(token)
      return fetchJson<ReadResult>(
        '/api/read',
        {
          method: 'POST',
          headers,
          body: JSON.stringify(payload),
        },
        handleUnauthorized,
      )
        .then((data: ReadResult) => setLastResult(data))
        .catch(() => undefined)
    },
    [handleUnauthorized, token],
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
      apiPost('/api/connect', token, handleUnauthorized).catch(() => undefined)
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

  if (authBlocked) {
    return <UnauthorizedPanel />
  }

  return (
    <AppLayout
      config={config}
      connected={connected}
      connecting={connecting}
      connectionError={connectionError}
      logs={logs}
      showLogs={showLogs}
      stats={stats}
      version={version}
      invocation={invocation}
      columns={columns}
      selectedKind={selectedKind}
      addressInput={addressInput}
      quantity={quantity}
      lastResult={lastResult}
      autoConnect={autoConnect}
      addressError={addressError}
      quantityError={quantityError}
      onSaveConfig={(next) => updateConfig(next, connected || connecting)}
      onUnauthorized={handleUnauthorized}
      onUpdateDecoder={updateDecoder}
      onColumnsChange={setColumns}
      onAddressBaseChange={setAddressBase}
      onAddressFormatChange={setAddressFormat}
      onValueBaseChange={setValueBase}
      onKindChange={setSelectedKind}
      onAddressChange={handleAddressChange}
      onQuantityChange={handleQuantityChange}
      onRead={handleRead}
      onAutoConnectChange={setAutoConnect}
      onToggleLogs={() => setShowLogs((prev) => !prev)}
      onConnect={handleConnect}
      onDisconnect={handleDisconnect}
    />
  )
}

export default App
