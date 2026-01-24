import type { DecoderConfig } from '../types'
import type { ReadKind, ReadResult } from '../view-models'

const readKinds: { label: string; value: ReadKind; code: string }[] = [
  { label: 'Coils', value: 'coils', code: '01' },
  { label: 'Discrete Inputs', value: 'discrete_inputs', code: '02' },
  { label: 'Holding Registers', value: 'holding_registers', code: '03' },
  { label: 'Input Registers', value: 'input_registers', code: '04' },
]

type Props = {
  selectedKind: ReadKind
  addressInput: string
  quantity: number
  addressBase: number
  addressFormat: number
  valueBase: number
  decoders: DecoderConfig[]
  lastResult: ReadResult | null
  columns: number
  connected: boolean
  autoConnect: boolean
  addressError: string
  quantityError: string
  onKindChange: (kind: ReadKind) => void
  onAddressChange: (value: string) => void
  onQuantityChange: (value: number) => void
  onRead: () => void
  onAutoConnectChange: (value: boolean) => void
}

type Cell = { value: string; colSpan: number }

type RenderRow = {
  label: string
  cells: Cell[]
  tone?: string
}

export default function ReadPanel({
  selectedKind,
  addressInput,
  quantity,
  addressBase,
  addressFormat,
  valueBase,
  decoders,
  lastResult,
  columns,
  connected,
  autoConnect,
  addressError,
  quantityError,
  onKindChange,
  onAddressChange,
  onQuantityChange,
  onRead,
  onAutoConnectChange,
}: Props) {
  const rows = buildRows(lastResult, decoders, addressBase, addressFormat, valueBase, columns)

  return (
    <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-slate-500">Read operations</p>
          <h2 className="text-xl font-semibold">Manual read</h2>
        </div>
        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.2em] text-slate-500">
            <input
              type="checkbox"
              checked={autoConnect}
              onChange={(event) => onAutoConnectChange(event.target.checked)}
              className="h-4 w-4 accent-slate-900"
            />
            Auto connect
          </label>
          <button
            className={`rounded-full px-5 py-2 text-sm font-semibold ${
              connected && !addressError && !quantityError
                ? 'bg-slate-900 text-white'
                : 'cursor-not-allowed bg-slate-200 text-slate-400'
            }`}
            onClick={onRead}
            disabled={!connected || Boolean(addressError) || Boolean(quantityError)}
          >
            Read now
          </button>
        </div>
      </div>
      <div className="mt-6 grid gap-4 md:grid-cols-4">
        <div className="space-y-2 md:col-span-2">
          <label className="text-xs font-semibold uppercase tracking-[0.15em] text-slate-500">Function</label>
          <select
            className="w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm"
            value={selectedKind}
            onChange={(event) => onKindChange(event.target.value as ReadKind)}
          >
            {readKinds.map((kind) => (
              <option key={kind.value} value={kind.value}>
                {kind.code} | {kind.label}
              </option>
            ))}
          </select>
        </div>
        <div className="space-y-2 md:col-span-1">
          <label className="text-xs font-semibold uppercase tracking-[0.15em] text-slate-500">Start address</label>
          <input
            type="text"
            className="w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm"
            value={addressInput}
            onChange={(event) => onAddressChange(event.target.value)}
            placeholder="0x10 or 16"
          />
        </div>
        <div className="space-y-2 md:col-span-1">
          <label className="text-xs font-semibold uppercase tracking-[0.15em] text-slate-500">Quantity</label>
          <input
            type="number"
            className="w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm"
            value={quantity}
            onChange={(event) => onQuantityChange(Number(event.target.value))}
          />
          {quantityError && <p className="text-[11px] font-semibold text-rose-600">{quantityError}</p>}
        </div>
      </div>

      <div className="mt-4">
        <p className="text-[11px] text-slate-500">Prefix hex with 0x, decimal otherwise.</p>
        {addressError && <p className="text-[11px] font-semibold text-rose-600">{addressError}</p>}
      </div>

      <div className="mt-4 rounded-2xl border border-slate-200 bg-slate-50 p-4">
        <div className="flex items-center justify-between text-xs uppercase tracking-[0.2em] text-slate-500">
          <span>Values</span>
          <span>{lastResult ? new Date(lastResult.completedAt).toLocaleTimeString() : '—'}</span>
        </div>
        {lastResult?.errorMessage ? (
          <div className="mt-4 rounded-2xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700">
            {lastResult.errorMessage}
          </div>
        ) : rows.length === 0 ? (
          <p className="mt-4 text-sm text-slate-500">No data read yet.</p>
        ) : (
          <div className="mt-4 overflow-x-auto">
            <table className="min-w-full border-collapse text-xs">
              <thead>
                <tr className="text-left text-[10px] uppercase tracking-[0.2em] text-slate-500">
                  <th className="px-2 py-1">Base</th>
                  {Array.from({ length: columns }).map((_, index) => (
                    <th key={index} className="px-2 py-1">
                      +{index}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {rows.map((row, rowIndex) => (
                  <tr key={`${row.label}-${rowIndex}`} className={row.tone ?? ''}>
                    <td className="whitespace-nowrap border-t border-slate-200 px-2 py-1 font-semibold text-slate-600">
                      {row.label}
                    </td>
                    {row.cells.map((cell, cellIndex) => (
                      <td
                        key={`${rowIndex}-${cellIndex}`}
                        colSpan={cell.colSpan}
                        className="border-t border-slate-200 px-2 py-1 font-mono text-[11px] text-slate-900"
                      >
                        {cell.value}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}

function buildRows(
  result: ReadResult | null,
  decoders: DecoderConfig[],
  addressBase: number,
  addressFormat: number,
  valueBase: number,
  columns: number,
): RenderRow[] {
  if (!result) {
    return []
  }
  const rows: RenderRow[] = []
  const values = result.regValues ?? result.boolValues?.map((value) => (value ? 1 : 0)) ?? []

  for (let offset = 0; offset < values.length; offset += columns) {
    const slice = values.slice(offset, offset + columns)
    const baseAddress = result.address + offset
    rows.push({
      label: formatAddress(baseAddress, addressBase, addressFormat),
      cells: slice.map((value) => ({ value: formatValue(value, valueBase), colSpan: 1 })),
    })

    const enabledDecoders = decoders.filter((decoder) => decoder.enabled)
    if (enabledDecoders.length && result.regValues) {
      const regSlice = result.regValues.slice(offset, offset + columns)
      for (const decoder of enabledDecoders) {
        const decoded = decodeRegisters(regSlice, decoder, columns)
        rows.push({
          label: `↳ ${decoder.type}`,
          cells: decoded,
          tone: 'bg-white/60',
        })
      }
    }
  }

  return rows
}

function formatAddress(address: number, _addressBase: number, format: number) {
  if (format === 16) {
    return `0x${address.toString(16).padStart(4, '0')}`
  }
  return `${address}`
}

function formatValue(value: number, base: number) {
  if (base === 16) {
    return `0x${value.toString(16).padStart(4, '0')}`
  }
  return `${value}`
}

function decodeRegisters(regs: number[], decoder: DecoderConfig, columns: number): Cell[] {
  if (decoder.type === 'uint16') {
    return regs.map((value) => ({ value: formatValue(value, 10), colSpan: 1 }))
  }
  if (decoder.type === 'int16') {
    return regs.map((value) => ({ value: `${toInt16(value)}`, colSpan: 1 }))
  }
  if (decoder.type === 'uint32' || decoder.type === 'int32' || decoder.type === 'float32') {
    const cells: Cell[] = []
    for (let i = 0; i + 1 < regs.length; i += 2) {
      const high = regs[i]
      const low = regs[i + 1]
      const [first, second] = decoder.wordOrder === 'low-first' ? [low, high] : [high, low]
      const bytes = decoder.endianness === 'little'
        ? [first & 0xff, first >> 8, second & 0xff, second >> 8]
        : [first >> 8, first & 0xff, second >> 8, second & 0xff]
      const view = new DataView(new Uint8Array(bytes).buffer)
      let value: string
      if (decoder.type === 'uint32') {
        value = `${view.getUint32(0, false)}`
      } else if (decoder.type === 'int32') {
        value = `${view.getInt32(0, false)}`
      } else {
        value = `${view.getFloat32(0, false).toFixed(4)}`
      }
      cells.push({ value, colSpan: 2 })
    }
    if (cells.length === 0) {
      cells.push({ value: '—', colSpan: columns })
    }
    return cells
  }
  return regs.map((value) => ({ value: `${value}`, colSpan: 1 }))
}

function toInt16(value: number) {
  const masked = value & 0xffff
  return masked & 0x8000 ? masked - 0x10000 : masked
}
