export type ReadKind = 'coils' | 'discrete_inputs' | 'holding_registers' | 'input_registers'

export type ReadResult = {
  kind: ReadKind
  address: number
  quantity: number
  boolValues?: boolean[]
  regValues?: number[]
  decoded?: { type: string; value: number }[]
  latencyMs: number
  completedAt: string
  errorMessage?: string
}

export type LogEntry = {
  time: string
  direction: string
  message: string
}

export type Stats = {
  readCount: number
  errorCount: number
  lastLatencyMs: number
}

export type WsEvent = {
  type: 'data' | 'log' | 'stats' | 'error' | 'status'
  payload: any
}

export type ConnectionStatus = {
  connected: boolean
  connecting: boolean
  lastError?: string
}
