export type DecoderConfig = {
  type: string
  endianness: string
  wordOrder: string
  enabled: boolean
}

export type Config = {
  protocol: 'tcp' | 'rtu'
  unitId: number
  timeoutMs: number
  addressBase: number
  addressFormat: number
  valueBase: number
  serial: {
    device: string
    speed: number
    dataBits: number
    parity: string
    stopBits: number
  }
  tcp: {
    host: string
    port: number
  }
  listenAddr: string
  requireToken: boolean
  token: string
  decoders: DecoderConfig[]
}
