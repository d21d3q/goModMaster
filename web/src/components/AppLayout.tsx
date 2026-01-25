import ConfigForm from './ConfigForm'
import DecoderPanel from './DecoderPanel'
import DisplayPanel from './DisplayPanel'
import RawLog from './RawLog'
import ReadPanel from './ReadPanel'
import StatsPanel from './StatsPanel'
import { Badge } from './ui/badge'
import { Button } from './ui/button'
import { Tooltip, TooltipContent, TooltipTrigger } from './ui/tooltip'
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
} from './ui/sidebar'
import type { Config } from '../types'
import type { LogEntry, ReadKind, ReadResult, Stats } from '../view-models'

type AppLayoutProps = {
  config: Config | null
  connected: boolean
  connecting: boolean
  connectionError: string
  logs: LogEntry[]
  showLogs: boolean
  stats: Stats
  version: string
  invocation: string
  columns: number
  selectedKind: ReadKind
  addressInput: string
  quantity: number
  lastResult: ReadResult | null
  autoConnect: boolean
  addressError: string
  quantityError: string
  onSaveConfig: (next: Config) => void
  onUnauthorized: () => void
  onUpdateDecoder: (nextDecoder: { type: string; endianness: string; wordOrder: string; enabled: boolean }) => void
  onColumnsChange: (next: number) => void
  onAddressBaseChange: (base: number) => void
  onAddressFormatChange: (format: number) => void
  onValueBaseChange: (base: number) => void
  onKindChange: (kind: ReadKind) => void
  onAddressChange: (value: string) => void
  onQuantityChange: (value: number) => void
  onRead: () => void
  onAutoConnectChange: (value: boolean) => void
  onToggleLogs: () => void
  onConnect: () => void
  onDisconnect: () => void
}

function AppLayout({
  config,
  connected,
  connecting,
  connectionError,
  logs,
  showLogs,
  stats,
  version,
  invocation,
  columns,
  selectedKind,
  addressInput,
  quantity,
  lastResult,
  autoConnect,
  addressError,
  quantityError,
  onSaveConfig,
  onUnauthorized,
  onUpdateDecoder,
  onColumnsChange,
  onAddressBaseChange,
  onAddressFormatChange,
  onValueBaseChange,
  onKindChange,
  onAddressChange,
  onQuantityChange,
  onRead,
  onAutoConnectChange,
  onToggleLogs,
  onConnect,
  onDisconnect,
}: AppLayoutProps) {
  const statusLabel = connected ? 'online' : connecting ? 'connecting' : 'offline'
  const statusVariant = connected ? 'default' : connecting ? 'secondary' : 'outline'
  const actionVariant = connected || connecting ? 'secondary' : 'default'
  const addressBase = config?.addressBase ?? 0
  const addressFormat = config?.addressFormat ?? 10
  const valueBase = config?.valueBase ?? 10
  const decoders = config?.decoders ?? []

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
                onSave={onSaveConfig}
                connected={connected}
                connecting={connecting}
                onUnauthorized={onUnauthorized}
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
              <DecoderPanel decoders={decoders} onUpdate={onUpdateDecoder} />
            </SidebarGroupContent>
          </SidebarGroup>
          <SidebarSeparator />
          <SidebarGroup>
            <SidebarGroupLabel>Addressing</SidebarGroupLabel>
            <SidebarGroupContent>
              <DisplayPanel
                config={config}
                columns={columns}
                onColumnsChange={onColumnsChange}
                onAddressBaseChange={onAddressBaseChange}
                onAddressFormatChange={onAddressFormatChange}
                onValueBaseChange={onValueBaseChange}
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
                <Button size="sm" variant={actionVariant} onClick={connected || connecting ? onDisconnect : onConnect}>
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
              addressBase={addressBase}
              addressFormat={addressFormat}
              valueBase={valueBase}
              decoders={decoders}
              lastResult={lastResult}
              columns={columns}
              connected={connected}
              autoConnect={autoConnect}
              addressError={addressError}
              quantityError={quantityError}
              onKindChange={onKindChange}
              onAddressChange={onAddressChange}
              onQuantityChange={onQuantityChange}
              onRead={onRead}
              onAutoConnectChange={onAutoConnectChange}
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
                <Button size="sm" variant="outline" onClick={onToggleLogs}>
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

export default AppLayout
