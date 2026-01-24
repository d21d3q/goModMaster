import type { DecoderConfig } from '../types'

const decoderTypes = ['uint16', 'int16', 'uint32', 'int32', 'float32']

const endiannessOptions = [
  { label: 'BE', value: 'big' },
  { label: 'LE', value: 'little' },
]

const wordOrderOptions = [
  { label: 'HF', value: 'high-first' },
  { label: 'LF', value: 'low-first' },
]

type Props = {
  decoders: DecoderConfig[]
  onUpdate: (decoder: DecoderConfig) => void
}

export default function DecoderPanel({ decoders, onUpdate }: Props) {
  const findDecoder = (type: string) =>
    decoders.find((decoder) => decoder.type === type) ?? {
      type,
      endianness: 'big',
      wordOrder: 'high-first',
      enabled: false,
    }

  return (
    <div className="flex flex-wrap items-center gap-3">
      {decoderTypes.map((type) => {
        const decoder = findDecoder(type)
        return (
          <div key={type} className="flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-2 text-xs">
            <input
              type="checkbox"
              checked={Boolean(decoder.enabled)}
              onChange={() => onUpdate({ ...decoder, enabled: !decoder.enabled })}
              className="h-4 w-4 accent-slate-900"
            />
            <span className="uppercase tracking-[0.2em] text-slate-500">{type}</span>
            <select
              className="rounded-md border border-slate-200 bg-white px-2 py-1 text-[11px]"
              value={decoder.endianness}
              onChange={(event) => onUpdate({ ...decoder, endianness: event.target.value })}
            >
              {endiannessOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            <select
              className="rounded-md border border-slate-200 bg-white px-2 py-1 text-[11px]"
              value={decoder.wordOrder}
              onChange={(event) => onUpdate({ ...decoder, wordOrder: event.target.value })}
            >
              {wordOrderOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>
        )
      })}
    </div>
  )
}
