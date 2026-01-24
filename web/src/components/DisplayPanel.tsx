import type { Config } from '../types'
import { Button } from './ui/button'

type Props = {
  config: Config | null
  columns: number
  onColumnsChange: (value: number) => void
  onAddressBaseChange: (value: number) => void
  onAddressFormatChange: (value: number) => void
  onValueBaseChange: (value: number) => void
}

export default function DisplayPanel({
  config,
  columns,
  onColumnsChange,
  onAddressBaseChange,
  onAddressFormatChange,
  onValueBaseChange,
}: Props) {
  const activeVariant = (active: boolean) => (active ? 'default' : 'outline')

  return (
    <div className="space-y-2">
      <h3>Addressing</h3>
      <div className="space-y-2">
        <div className="flex items-center justify-between gap-2">
          <span>Base</span>
          <div className="flex gap-2">
            <Button size="sm" variant={activeVariant(config?.addressBase === 1)} onClick={() => onAddressBaseChange(1)}>
              1
            </Button>
            <Button size="sm" variant={activeVariant(config?.addressBase === 0)} onClick={() => onAddressBaseChange(0)}>
              0
            </Button>
          </div>
        </div>
        <div className="flex items-center justify-between gap-2">
          <span>Addr</span>
          <div className="flex gap-2">
            <Button
              size="sm"
              variant={activeVariant(config?.addressFormat === 10)}
              onClick={() => onAddressFormatChange(10)}
            >
              Dec
            </Button>
            <Button
              size="sm"
              variant={activeVariant(config?.addressFormat === 16)}
              onClick={() => onAddressFormatChange(16)}
            >
              Hex
            </Button>
          </div>
        </div>
        <div className="flex items-center justify-between gap-2">
          <span>Value</span>
          <div className="flex gap-2">
            <Button size="sm" variant={activeVariant(config?.valueBase === 10)} onClick={() => onValueBaseChange(10)}>
              Dec
            </Button>
            <Button size="sm" variant={activeVariant(config?.valueBase === 16)} onClick={() => onValueBaseChange(16)}>
              Hex
            </Button>
          </div>
        </div>
        <div className="flex items-center justify-between gap-2">
          <span>Cols</span>
          <div className="flex gap-2">
            {[4, 8, 16].map((count) => (
              <Button size="sm" key={count} variant={activeVariant(columns === count)} onClick={() => onColumnsChange(count)}>
                {count}
              </Button>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
