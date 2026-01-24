import type { DecoderConfig } from '../types'
import { decoderTypeOrder } from './decoder-order'
import { Checkbox } from './ui/checkbox'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select'

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
    <div className="grid grid-cols-[auto_minmax(0,1fr)_auto_auto] items-center gap-2">
      {decoderTypeOrder.map((type) => {
        const decoder = findDecoder(type)
        return (
          <div key={type} className="contents">
            <Checkbox
              checked={Boolean(decoder.enabled)}
              onCheckedChange={(value) => onUpdate({ ...decoder, enabled: Boolean(value) })}
            />
            <span className="truncate">{type}</span>
            <Select value={decoder.endianness} onValueChange={(value) => onUpdate({ ...decoder, endianness: value })}>
              <SelectTrigger size="sm" className="w-16">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {endiannessOptions.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={decoder.wordOrder} onValueChange={(value) => onUpdate({ ...decoder, wordOrder: value })}>
              <SelectTrigger size="sm" className="w-16">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {wordOrderOptions.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )
      })}
    </div>
  )
}
