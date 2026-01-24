import type { Config } from '../types'

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
  return (
    <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
      <h2 className="text-xl font-semibold">Addressing</h2>
      <div className="mt-4 flex flex-wrap items-center gap-3 text-xs font-semibold uppercase tracking-[0.15em] text-slate-500">
        <div className="flex items-center gap-2">
          <span>Base</span>
          <div className="flex gap-2">
            <button
              className={`rounded-xl border px-3 py-2 ${
                config?.addressBase === 1 ? 'border-slate-900 bg-slate-900 text-white' : 'border-slate-200 bg-white'
              }`}
              onClick={() => onAddressBaseChange(1)}
            >
              1
            </button>
            <button
              className={`rounded-xl border px-3 py-2 ${
                config?.addressBase === 0 ? 'border-slate-900 bg-slate-900 text-white' : 'border-slate-200 bg-white'
              }`}
              onClick={() => onAddressBaseChange(0)}
            >
              0
            </button>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <span>Addr</span>
          <div className="flex gap-2">
            <button
              className={`rounded-xl border px-3 py-2 ${
                config?.addressFormat === 10 ? 'border-slate-900 bg-slate-900 text-white' : 'border-slate-200 bg-white'
              }`}
              onClick={() => onAddressFormatChange(10)}
            >
              Dec
            </button>
            <button
              className={`rounded-xl border px-3 py-2 ${
                config?.addressFormat === 16 ? 'border-slate-900 bg-slate-900 text-white' : 'border-slate-200 bg-white'
              }`}
              onClick={() => onAddressFormatChange(16)}
            >
              Hex
            </button>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <span>Value</span>
          <div className="flex gap-2">
            <button
              className={`rounded-xl border px-3 py-2 ${
                config?.valueBase === 10 ? 'border-slate-900 bg-slate-900 text-white' : 'border-slate-200 bg-white'
              }`}
              onClick={() => onValueBaseChange(10)}
            >
              Dec
            </button>
            <button
              className={`rounded-xl border px-3 py-2 ${
                config?.valueBase === 16 ? 'border-slate-900 bg-slate-900 text-white' : 'border-slate-200 bg-white'
              }`}
              onClick={() => onValueBaseChange(16)}
            >
              Hex
            </button>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <span>Cols</span>
          <div className="flex gap-2">
            {[8, 16].map((count) => (
              <button
                key={count}
                className={`rounded-xl border px-3 py-2 ${
                  columns === count ? 'border-slate-900 bg-slate-900 text-white' : 'border-slate-200 bg-white'
                }`}
                onClick={() => onColumnsChange(count)}
              >
                {count}
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
