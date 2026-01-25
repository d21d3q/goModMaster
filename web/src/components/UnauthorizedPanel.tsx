type Props = {
}

export default function UnauthorizedPanel(_props: Props) {
  return (
    <div className="flex min-h-svh items-center justify-center p-6">
      <div className="max-w-md rounded-xl border bg-background p-6 text-center shadow-sm">
        <h1 className="text-lg font-semibold">Session expired</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          This link is no longer authorized. Please copy the fresh URL printed in the terminal and open it again.
        </p>
      </div>
    </div>
  )
}
