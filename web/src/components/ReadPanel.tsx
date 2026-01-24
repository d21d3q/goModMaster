import type { DecoderConfig } from '../types'
import type { ReadKind, ReadResult } from '../view-models'
import { decoderTypeOrder } from './decoder-order'
import { Badge } from './ui/badge'
import { Button } from './ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card'
import { Checkbox } from './ui/checkbox'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table'
import { Tooltip, TooltipContent, TooltipTrigger } from './ui/tooltip'

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

type Cell = { value: string; colSpan: number; fullValue?: string }

type RenderRow = {
  label: string
  cells: Cell[]
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
  const canRead = connected && !addressError && !quantityError

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <CardDescription>Read operations</CardDescription>
            <CardTitle>Manual read</CardTitle>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <div className="flex items-center gap-2">
              <Checkbox
                id="auto-connect"
                checked={autoConnect}
                onCheckedChange={(value) => onAutoConnectChange(Boolean(value))}
              />
              <Label htmlFor="auto-connect">Auto connect</Label>
            </div>
            <Button size="sm" onClick={onRead} disabled={!canRead}>
              Read now
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-4 md:grid-cols-4">
          <div className="grid gap-1 md:col-span-2">
            <Label htmlFor="read-kind">Function</Label>
            <Select value={selectedKind} onValueChange={(value) => onKindChange(value as ReadKind)}>
              <SelectTrigger className="w-full" id="read-kind">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {readKinds.map((kind) => (
                  <SelectItem key={kind.value} value={kind.value}>
                    {kind.code} | {kind.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="grid gap-1 md:col-span-1">
            <Label htmlFor="start-address">Start address</Label>
            <Input
              id="start-address"
              type="text"
              value={addressInput}
              onChange={(event) => onAddressChange(event.target.value)}
              placeholder="0x10 or 16"
            />
          </div>
          <div className="grid gap-1 md:col-span-1">
            <Label htmlFor="quantity">Quantity</Label>
            <Input
              id="quantity"
              type="number"
              value={quantity}
              onChange={(event) => onQuantityChange(Number(event.target.value))}
            />
            {quantityError && <Badge variant="destructive">{quantityError}</Badge>}
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <small>Prefix hex address with 0x, decimal otherwise.</small>
          {addressError && <Badge variant="destructive">{addressError}</Badge>}
        </div>

        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <span>Values</span>
            <Badge variant="outline">{lastResult ? new Date(lastResult.completedAt).toLocaleTimeString() : '—'}</Badge>
          </div>
          {lastResult?.errorMessage ? (
            <Badge variant="destructive">{lastResult.errorMessage}</Badge>
          ) : rows.length === 0 ? (
            <p>No data read yet.</p>
          ) : (
            <Table className="table-fixed">
              <TableHeader>
                <TableRow>
                  <TableHead className="w-24">Base</TableHead>
                  {Array.from({ length: columns }).map((_, index) => (
                    <TableHead key={index}>+{index}</TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {rows.map((row, rowIndex) => (
                  <TableRow key={`${row.label}-${rowIndex}`}>
                    <TableCell className="w-24">{row.label}</TableCell>
                    {row.cells.map((cell, cellIndex) => (
                      <TableCell key={`${rowIndex}-${cellIndex}`} colSpan={cell.colSpan}>
                        {cell.fullValue ? (
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <code className="cursor-help">{cell.value}</code>
                            </TooltipTrigger>
                            <TooltipContent>{cell.fullValue}</TooltipContent>
                          </Tooltip>
                        ) : (
                          <code>{cell.value}</code>
                        )}
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </div>
      </CardContent>
    </Card>
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
  const orderIndex = (type: string) => {
    const index = decoderTypeOrder.indexOf(type as (typeof decoderTypeOrder)[number])
    return index === -1 ? Number.MAX_SAFE_INTEGER : index
  }

  for (let offset = 0; offset < values.length; offset += columns) {
    const slice = values.slice(offset, offset + columns)
    const baseAddress = result.address + offset
    rows.push({
      label: formatAddress(baseAddress, addressBase, addressFormat),
      cells: slice.map((value) => ({ value: formatValue(value, valueBase), colSpan: 1 })),
    })

    const enabledDecoders = decoders
      .filter((decoder) => decoder.enabled)
      .sort((left, right) => orderIndex(left.type) - orderIndex(right.type))
    if (enabledDecoders.length && result.regValues) {
      const regSlice = result.regValues.slice(offset, offset + columns)
      for (const decoder of enabledDecoders) {
        const decoded = decodeRegisters(regSlice, decoder, columns)
        rows.push({
          label: `↳ ${decoder.type}`,
          cells: decoded,
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
      let fullValue: string | undefined
      if (decoder.type === 'uint32') {
        value = `${view.getUint32(0, false)}`
      } else if (decoder.type === 'int32') {
        value = `${view.getInt32(0, false)}`
      } else {
        const raw = view.getFloat32(0, false)
        value = formatFloat(raw)
        fullValue = `${raw}`
      }
      cells.push({ value, colSpan: 2, fullValue })
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

function formatFloat(value: number) {
  if (!Number.isFinite(value)) {
    return String(value)
  }
  if (Math.abs(value) >= 1e21) {
    return value.toExponential(3)
  }
  return value.toFixed(3)
}
