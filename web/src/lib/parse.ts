export function parseAddress(input: string): number | null {
  const trimmed = input.trim().toLowerCase()
  if (trimmed === '') {
    return null
  }
  if (trimmed.startsWith('0x')) {
    const hex = trimmed.slice(2)
    if (!/^[0-9a-f]+$/.test(hex)) {
      return null
    }
    return Number.parseInt(hex, 16)
  }
  if (!/^[0-9]+$/.test(trimmed)) {
    return null
  }
  const value = Number.parseInt(trimmed, 10)
  if (Number.isNaN(value)) {
    return null
  }
  return value
}
